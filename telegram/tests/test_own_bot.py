"""Probe test for configured client bot target from telegram/.env."""

import os

import pytest

from revisitr_telegram.response_collector import collect_bot_responses_after


@pytest.mark.asyncio(loop_scope="session")
async def test_start_command(tg_client, bot_entity, bot_username):
    me = await tg_client.get_me()
    latest = await tg_client.get_messages(bot_entity, limit=1)
    last_seen_id = latest[0].id if latest else 0

    await tg_client.send_message(bot_entity, "/start")
    bot_responses = await collect_bot_responses_after(
        tg_client,
        bot_entity,
        me.id,
        last_seen_id,
        timeout=15.0,
        quiet_period=1.5,
    )

    if not bot_responses:
        pytest.xfail(
            f"configured BOT_USERNAME '{bot_username}' did not respond to /start; target likely stale or offline"
        )

    has_content = any(r.text or r.media for r in bot_responses)
    assert has_content, "Bot response has no text or media content"
