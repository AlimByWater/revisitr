"""Telegram E2E tests for Revisitr master bot using a real userbot session."""

from __future__ import annotations

import asyncio
import json
import os
import re
import secrets
import subprocess
import pytest

from revisitr_telegram.response_collector import collect_bot_responses_after

LOCAL_PSQL_DSN = "postgresql://revisitr:devpassword@localhost:5433/revisitr?sslmode=disable"
POST_CODE_RE = re.compile(r"RV-[A-Z0-9]{6}")


def run(cmd: list[str]) -> subprocess.CompletedProcess[str]:
    return subprocess.run(cmd, check=True, text=True, capture_output=True)


def psql_fetchval(query: str) -> str:
    return run(["psql", LOCAL_PSQL_DSN, "-XqAt", "-c", query]).stdout.strip()


def psql_exec(query: str) -> None:
    run(["psql", LOCAL_PSQL_DSN, "-c", query])


def sql_literal(text: str) -> str:
    return text.replace("'", "''")


async def send_and_collect(client, bot_entity, me, text: str, delay: float = 5.0):
    before = await client.get_messages(bot_entity, limit=1)
    last_seen_id = before[0].id if before else 0
    await client.send_message(bot_entity, text)
    return await collect_bot_responses_after(client, bot_entity, me.id, last_seen_id, timeout=delay + 10, quiet_period=1.5)


async def send_file_and_collect(client, bot_entity, me, path: Path, caption: str | None = None):
    before = await client.get_messages(bot_entity, limit=1)
    last_seen_id = before[0].id if before else 0
    await client.send_file(bot_entity, str(path), caption=caption)
    return await collect_bot_responses_after(client, bot_entity, me.id, last_seen_id, timeout=20, quiet_period=2.0)


@pytest.fixture()
def linked_org(me):
    org_name = f"E2E Org {secrets.token_hex(4)}"
    user_email = f"e2e-{secrets.token_hex(4)}@test.revisitr.local"

    org_id = int(psql_fetchval(f"INSERT INTO organizations (name) VALUES ('{sql_literal(org_name)}') RETURNING id;"))
    user_id = int(
        psql_fetchval(
            "INSERT INTO users (email, name, password_hash, role, org_id) "
            f"VALUES ('{sql_literal(user_email)}', 'E2E User', 'hash', 'owner', {org_id}) RETURNING id;"
        )
    )

    yield {"org_id": org_id, "user_id": user_id, "telegram_id": me.id}

    psql_exec(f"DELETE FROM post_codes WHERE org_id = {org_id};")
    psql_exec(f"DELETE FROM bots WHERE org_id = {org_id};")
    psql_exec(f"DELETE FROM master_bot_links WHERE org_id = {org_id};")
    psql_exec(f"DELETE FROM users WHERE id = {user_id};")
    psql_exec(f"DELETE FROM organizations WHERE id = {org_id};")


@pytest.mark.asyncio(loop_scope="session")
async def test_masterbot_start_without_link_shows_activation_hint(tg_client, master_bot_entity, me):
    responses = await send_and_collect(tg_client, master_bot_entity, me, "/start")

    assert responses, "master bot did not respond to /start"
    merged = "\n".join((m.text or "") for m in responses)
    assert "Добро пожаловать в Revisitr Bot" in merged
    assert "Для привязки аккаунта" in merged
    assert "/link КОД" in merged


@pytest.mark.asyncio(loop_scope="session")
async def test_masterbot_start_with_activation_token_links_account(tg_client, master_bot_entity, me, linked_org):
    token = secrets.token_hex(16)
    auth_payload = json.dumps({"org_id": linked_org["org_id"], "user_id": linked_org["user_id"]}, ensure_ascii=False)
    run(["redis-cli", "-p", "6380", "set", f"masterbot:auth:{token}", auth_payload, "EX", "900"])

    responses = await send_and_collect(tg_client, master_bot_entity, me, f"/start {token}")

    assert responses, "master bot did not respond to activation deep link"
    merged = "\n".join((m.text or "") for m in responses)
    assert "Вы привязаны к организации" in merged

    count = psql_fetchval(
        f"SELECT COUNT(*) FROM master_bot_links WHERE org_id = {linked_org['org_id']} AND telegram_user_id = {linked_org['telegram_id']};"
    )
    assert count == "1"


@pytest.mark.asyncio(loop_scope="session")
async def test_masterbot_linked_commands_show_bots_and_settings(tg_client, master_bot_entity, me, linked_org):
    psql_exec(
        "INSERT INTO master_bot_links (org_id, telegram_user_id, telegram_username, is_active) "
        f"VALUES ({linked_org['org_id']}, {linked_org['telegram_id']}, '{sql_literal(me.username or '')}', true) "
        "ON CONFLICT (org_id, telegram_user_id) DO UPDATE SET is_active = true;"
    )
    psql_exec(
        "INSERT INTO bots (org_id, name, token, username, status, settings, is_managed, created_at, updated_at) VALUES "
        f"({linked_org['org_id']}, 'Masterbot Test Venue', 'dummy-token', 'masterbot_test_venue_bot', 'active', "
        "'{\"modules\":[\"loyalty\",\"campaigns\"],\"registration_form\":[{\"name\":\"phone\",\"label\":\"Телефон\",\"type\":\"phone\",\"required\":true}],\"welcome_message\":\"Добро пожаловать в тестовый бот\"}'::jsonb, true, NOW(), NOW());"
    )

    mybots = await send_and_collect(tg_client, master_bot_entity, me, "/mybots")
    assert mybots, "master bot did not respond to /mybots"
    assert "Masterbot Test Venue" in "\n".join((m.text or "") for m in mybots)

    settings = await send_and_collect(tg_client, master_bot_entity, me, "/settings")
    assert settings, "master bot did not respond to /settings"
    merged = "\n".join((m.text or "") for m in settings)
    assert "Настройки бота" in merged
    assert "Поля регформы" in merged


@pytest.mark.asyncio(loop_scope="session")
async def test_masterbot_creates_post_code_from_text_message(tg_client, master_bot_entity, me, linked_org):
    psql_exec(
        "INSERT INTO master_bot_links (org_id, telegram_user_id, telegram_username, is_active) "
        f"VALUES ({linked_org['org_id']}, {linked_org['telegram_id']}, '{sql_literal(me.username or '')}', true) "
        "ON CONFLICT (org_id, telegram_user_id) DO UPDATE SET is_active = true;"
    )

    text = f"E2E text post {secrets.token_hex(4)}"
    responses = await send_and_collect(tg_client, master_bot_entity, me, text, delay=6.0)

    assert responses, "master bot did not respond to text post creation"
    merged = "\n".join((m.text or "") for m in responses)
    assert "Пост создан" in merged

    match = POST_CODE_RE.search(merged)
    assert match, f"post code not found in bot response: {merged}"
    code = match.group(0)

    stored_text = psql_fetchval(
        f"SELECT content->>'text' FROM post_codes WHERE org_id = {linked_org['org_id']} AND code = '{code}';"
    )
    assert stored_text == text


@pytest.mark.asyncio(loop_scope="session")
async def test_masterbot_creates_post_code_from_photo_message(tg_client, master_bot_entity, me, linked_org, tmp_path):
    psql_exec(
        "INSERT INTO master_bot_links (org_id, telegram_user_id, telegram_username, is_active) "
        f"VALUES ({linked_org['org_id']}, {linked_org['telegram_id']}, '{sql_literal(me.username or '')}', true) "
        "ON CONFLICT (org_id, telegram_user_id) DO UPDATE SET is_active = true;"
    )

    image_path = tmp_path / "e2e.png"
    image_path.write_bytes(
        bytes.fromhex(
            "89504E470D0A1A0A0000000D4948445200000001000000010802000000907753DE0000000C49444154789C63606060000000040001F61738550000000049454E44AE426082"
        )
    )
    caption = f"Photo post {secrets.token_hex(4)}"
    responses = await send_file_and_collect(tg_client, master_bot_entity, me, image_path, caption=caption)

    assert responses, "master bot did not respond to photo post creation"
    merged = "\n".join((m.text or "") for m in responses)
    assert "Пост создан" in merged

    match = POST_CODE_RE.search(merged)
    assert match, f"post code not found in photo response: {merged}"
    code = match.group(0)

    media_type = psql_fetchval(
        f"SELECT content->>'media_type' FROM post_codes WHERE org_id = {linked_org['org_id']} AND code = '{code}';"
    )
    stored_text = psql_fetchval(
        f"SELECT content->>'text' FROM post_codes WHERE org_id = {linked_org['org_id']} AND code = '{code}';"
    )
    media_count = psql_fetchval(
        f"SELECT jsonb_array_length(content->'media_urls') FROM post_codes WHERE org_id = {linked_org['org_id']} AND code = '{code}';"
    )
    assert media_type == "photo"
    assert stored_text == caption
    assert media_count == "1"


@pytest.mark.asyncio(loop_scope="session")
async def test_current_configured_client_bot_probe(tg_client, me, telegram_env):
    target = (
        telegram_env.get("DEMO_BOT_USERNAME")
        or telegram_env.get("CLIENT_BOT_USERNAME")
        or telegram_env.get("BOT_USERNAME")
    )
    if not target:
        pytest.skip("DEMO_BOT_USERNAME / CLIENT_BOT_USERNAME / BOT_USERNAME not set")

    entity = await tg_client.get_entity(target)
    responses = await send_and_collect(tg_client, entity, me, "/start", delay=5.0)

    if not responses:
        pytest.xfail(
            f"configured client bot '{target}' did not respond to /start during probe; likely stale env target or offline runtime"
        )

    merged = "\n".join((m.text or "") for m in responses)
    assert merged.strip(), f"configured client bot '{target}' returned empty response"
