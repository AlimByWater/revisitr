"""Response collection helpers for Telegram bot messages."""

import asyncio
import time

from telethon import TelegramClient


async def collect_bot_responses_after(
    client: TelegramClient,
    bot_entity,
    me_id: int,
    last_seen_id: int,
    timeout: float = 15.0,
    quiet_period: float = 1.5,
    poll_interval: float = 0.5,
    limit: int = 20,
) -> list:
    """Collect only fresh bot responses after a known message id.

    Polls Telegram until at least one new bot message appears and then waits
    for a short quiet period to capture multipart responses.
    """
    deadline = time.monotonic() + timeout
    latest_activity = None
    collected: list = []
    max_seen_id = last_seen_id

    while time.monotonic() < deadline:
        messages = await client.get_messages(bot_entity, limit=limit, min_id=max_seen_id)
        fresh = [m for m in reversed(messages) if m.sender_id != me_id]

        if fresh:
            collected.extend([m for m in fresh if m.id not in {x.id for x in collected}])
            max_seen_id = max(max_seen_id, *(m.id for m in fresh))
            latest_activity = time.monotonic()

        if collected and latest_activity is not None and time.monotonic() - latest_activity >= quiet_period:
            break

        await asyncio.sleep(poll_interval)

    return sorted(collected, key=lambda m: m.id)


async def collect_bot_responses(
    client: TelegramClient,
    bot_entity,
    timeout: float = 10.0,
    quiet_period: float = 2.0,
) -> list:
    """Backward-compatible wrapper for existing callers."""
    me = await client.get_me()
    latest = await client.get_messages(bot_entity, limit=1)
    last_seen_id = latest[0].id if latest else 0
    return await collect_bot_responses_after(
        client,
        bot_entity,
        me.id,
        last_seen_id,
        timeout=timeout,
        quiet_period=quiet_period,
    )
