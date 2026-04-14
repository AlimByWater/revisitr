"""Pytest fixtures for Telethon Telegram bot E2E tests."""

from __future__ import annotations

import os
import subprocess
import time
from pathlib import Path

import httpx
import pytest
import pytest_asyncio
from dotenv import dotenv_values, load_dotenv

from revisitr_telegram.client import get_client

ROOT = Path(__file__).resolve().parents[2]
TELEGRAM_DIR = ROOT / "telegram"
BACKEND_DIR = ROOT / "backend"
LOCAL_DB_URL = "postgres://revisitr:devpassword@localhost:5433/revisitr?sslmode=disable"
LOCAL_PSQL_DSN = "postgresql://revisitr:devpassword@localhost:5433/revisitr?sslmode=disable"
REDIS_DIR = ROOT / "infra"
MASTERBOT_LOG = TELEGRAM_DIR / "results" / "masterbot-e2e.log"

load_dotenv(TELEGRAM_DIR / ".env")


def run(cmd: list[str], cwd: Path | None = None, env: dict[str, str] | None = None) -> subprocess.CompletedProcess[str]:
    return subprocess.run(cmd, cwd=cwd, env=env, check=True, text=True, capture_output=True)


def psql_fetchval(query: str) -> str:
    result = run(["psql", LOCAL_PSQL_DSN, "-tAc", query])
    return result.stdout.strip()


def psql_exec(query: str) -> None:
    run(["psql", LOCAL_PSQL_DSN, "-c", query])


@pytest.fixture(scope="session")
def telegram_env() -> dict[str, str]:
    env = dotenv_values(TELEGRAM_DIR / ".env")
    return {k: str(v) for k, v in env.items() if v is not None}


def resolve_probe_bot_username(env: dict[str, str]) -> str | None:
    return (
        env.get("DEMO_BOT_USERNAME")
        or env.get("CLIENT_BOT_USERNAME")
        or env.get("BOT_USERNAME")
    )


@pytest.fixture(scope="session")
def master_bot_token() -> str:
    backend_env = dotenv_values(BACKEND_DIR / ".env")
    token = backend_env.get("MASTER_BOT_TOKEN") or backend_env.get("ADMIN_BOT_TOKEN")
    if not token:
        pytest.skip("MASTER_BOT_TOKEN / ADMIN_BOT_TOKEN not found in backend/.env")
    return str(token)


@pytest.fixture(scope="session")
def master_bot_username(master_bot_token: str) -> str:
    response = httpx.get(f"https://api.telegram.org/bot{master_bot_token}/getMe", verify=False, timeout=15)
    response.raise_for_status()
    data = response.json()
    if not data.get("ok"):
        pytest.skip("Unable to resolve master bot username via Bot API")
    return data["result"]["username"]


@pytest.fixture(scope="session", autouse=True)
def local_infra() -> None:
    run(["docker", "compose", "up", "-d", "postgres", "redis"], cwd=REDIS_DIR)
    run(["goose", "-dir", "migrations", "postgres", LOCAL_DB_URL, "up"], cwd=BACKEND_DIR)
    run(["goose", "-dir", "migrations", "postgres", LOCAL_DB_URL, "up"], cwd=ROOT)
    yield
    subprocess.run(["docker", "compose", "stop", "postgres", "redis"], cwd=REDIS_DIR, check=False, text=True, capture_output=True)


@pytest.fixture(scope="session", autouse=True)
def local_masterbot(local_infra: None, master_bot_token: str) -> None:
    backend_env = dotenv_values(BACKEND_DIR / ".env")
    env = os.environ.copy()
    env.update({k: str(v) for k, v in backend_env.items() if v is not None})
    env.update(
        {
            "MASTER_BOT_TOKEN": master_bot_token,
            "POSTGRES_HOST": "localhost",
            "POSTGRES_PORT": "5433",
            "POSTGRES_USER": "revisitr",
            "POSTGRES_PASSWORD": "devpassword",
            "POSTGRES_DATABASE": "revisitr",
            "POSTGRES_SSLMODE": "disable",
            "REDIS_HOST": "localhost",
            "REDIS_PORT": "6380",
            "REDIS_PASSWORD": "",
        }
    )

    MASTERBOT_LOG.parent.mkdir(parents=True, exist_ok=True)
    proc = None
    startup_error = None

    for _ in range(3):
        with MASTERBOT_LOG.open("w") as log_file:
            proc = subprocess.Popen(
                ["go", "run", "./cmd/masterbot"],
                cwd=BACKEND_DIR,
                env=env,
                stdout=log_file,
                stderr=subprocess.STDOUT,
                text=True,
            )

        deadline = time.time() + 30
        while time.time() < deadline:
            if proc.poll() is not None:
                startup_error = MASTERBOT_LOG.read_text(errors="ignore")
                break
            if MASTERBOT_LOG.exists() and "master bot service started" in MASTERBOT_LOG.read_text(errors="ignore"):
                startup_error = None
                break
            time.sleep(0.5)

        if startup_error is None and proc and proc.poll() is None:
            break

        if proc and proc.poll() is None:
            proc.terminate()
            try:
                proc.wait(timeout=5)
            except subprocess.TimeoutExpired:
                proc.kill()
        time.sleep(1)
    else:
        raise RuntimeError(startup_error or "master bot did not start in time")

    yield

    if proc and proc.poll() is None:
        proc.terminate()
        try:
            proc.wait(timeout=10)
        except subprocess.TimeoutExpired:
            proc.kill()


@pytest_asyncio.fixture(scope="session")
async def tg_client(local_masterbot: None):
    try:
        async with get_client(str(TELEGRAM_DIR / ".env")) as client:
            yield client
    except (ValueError, RuntimeError) as e:
        pytest.skip(f"Telethon client unavailable: {e}")


@pytest_asyncio.fixture(scope="session")
async def me(tg_client):
    return await tg_client.get_me()


@pytest_asyncio.fixture(scope="session")
async def master_bot_entity(tg_client, master_bot_username):
    return await tg_client.get_entity(master_bot_username)


@pytest.fixture(scope="session")
def bot_username(telegram_env: dict[str, str]) -> str:
    username = resolve_probe_bot_username(telegram_env)
    if not username:
        pytest.skip("DEMO_BOT_USERNAME / CLIENT_BOT_USERNAME / BOT_USERNAME not set in telegram/.env")
    return username


@pytest_asyncio.fixture(scope="session")
async def bot_entity(tg_client, bot_username: str):
    return await tg_client.get_entity(bot_username)
