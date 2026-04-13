"""Telethon client factory with async context manager lifecycle."""

import os
from contextlib import asynccontextmanager
from pathlib import Path

from dotenv import load_dotenv
from telethon import TelegramClient
from telethon.sessions import StringSession

REQUIRED_VARS = ("API_ID", "API_HASH")


def _resolve_session() -> StringSession | str:
    """Resolve Telethon session source.

    Priority:
    1. SESSION_STRING — portable StringSession
    2. SESSION_PATH — explicit path to a SQLite .session file
    3. SESSION_NAME — legacy session filename/name
    """
    session_string = os.getenv("SESSION_STRING")
    if session_string:
        return StringSession(session_string)

    session_path = os.getenv("SESSION_PATH")
    if session_path:
        return str(Path(session_path).expanduser())

    session_name = os.getenv("SESSION_NAME")
    if session_name:
        return session_name

    raise ValueError("Missing required env var: SESSION_STRING, SESSION_PATH, or SESSION_NAME")


def create_client(env_path: str = ".env") -> TelegramClient:
    """Create an unconnected TelegramClient from .env credentials."""
    load_dotenv(Path(env_path))

    missing = [v for v in REQUIRED_VARS if not os.getenv(v)]
    if missing:
        raise ValueError(f"Missing required env vars: {', '.join(missing)}")

    api_id = int(os.environ["API_ID"])
    api_hash = os.environ["API_HASH"]
    session = _resolve_session()

    return TelegramClient(session, api_id, api_hash)


@asynccontextmanager
async def get_client(env_path: str = ".env"):
    """Async context manager that yields a connected, authorized TelegramClient."""
    client = create_client(env_path)
    try:
        await client.connect()
        if not await client.is_user_authorized():
            raise RuntimeError(
                "Session not authorized. Provide a valid SESSION_STRING or SESSION_PATH/SESSION_NAME."
            )
        yield client
    finally:
        await client.disconnect()
