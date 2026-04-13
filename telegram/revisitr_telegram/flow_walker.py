"""AI-driven Telegram bot flow walker using Telethon + Codex."""

import argparse
import asyncio
import hashlib
import json
import logging
import os
from datetime import datetime, timezone
from pathlib import Path

import httpx
from dotenv import load_dotenv
from telethon import TelegramClient
from telethon.errors import FloodWaitError
from telethon.tl.types import (
    KeyboardButtonCallback,
    KeyboardButtonUrl,
    ReplyInlineMarkup,
    ReplyKeyboardMarkup,
)

from revisitr_telegram.client import get_client
from revisitr_telegram.response_collector import collect_bot_responses

logger = logging.getLogger(__name__)

CODEX_SYSTEM_PROMPT = """\
You are a Telegram bot navigator. You explore bot menus by choosing which button to press next.

Current bot response:
{response_text}

Available actions:
{actions_list}

Previously visited paths (avoid revisiting):
{visited_paths}

Remaining depth: {depth_remaining}

Rules:
- Choose the most informative unexplored action
- If all actions have been visited, respond with "done"
- Never invent button labels that are not in the available actions list
- Respond ONLY with valid JSON

Respond with exactly one JSON object:
{{"action": "press_button", "target": "Button Label"}}
or {{"action": "send_text", "text": "/some_command"}}
or {{"action": "done", "reason": "all paths explored"}}"""

VALID_ACTIONS = {"press_button", "send_text", "done"}


def validate_codex_response(response: dict, available_buttons: list[str]) -> dict:
    """Validate Codex response and handle hallucinations."""
    action = response.get("action")
    if action not in VALID_ACTIONS:
        raise ValueError(f"Invalid action: {action}")
    if action == "press_button":
        target = response.get("target")
        if target not in available_buttons:
            logger.warning(
                "Codex hallucinated button '%s', falling back to first available", target
            )
            response["target"] = available_buttons[0] if available_buttons else None
    return response


def _extract_buttons(message) -> list[dict]:
    """Extract buttons from a Telegram message's reply_markup."""
    buttons = []
    markup = message.reply_markup
    if markup is None:
        return buttons

    if isinstance(markup, ReplyKeyboardMarkup):
        for row in markup.rows:
            for btn in row.buttons:
                buttons.append({"label": btn.text, "type": "reply_keyboard"})
    elif isinstance(markup, ReplyInlineMarkup):
        for row in markup.rows:
            for btn in row.buttons:
                entry = {"label": btn.text, "type": "inline_keyboard"}
                if isinstance(btn, KeyboardButtonUrl):
                    entry["url"] = btn.url
                elif isinstance(btn, KeyboardButtonCallback):
                    entry["callback_data"] = btn.data.decode() if isinstance(btn.data, bytes) else btn.data
                buttons.append(entry)
    return buttons


def _extract_media(message) -> list[dict]:
    """Extract media info from a Telegram message."""
    if not message.media:
        return []
    media_type = "unknown"
    if message.photo:
        media_type = "photo"
    elif message.video:
        media_type = "video"
    elif message.document:
        media_type = "document"
    elif message.gif:
        media_type = "animation"
    elif message.sticker:
        media_type = "sticker"
    elif message.audio:
        media_type = "audio"
    elif message.voice:
        media_type = "voice"
    return [{"type": media_type, "caption": message.text or None, "has_media": True}]


class FlowWalker:
    """AI-driven bot navigation explorer."""

    def __init__(
        self,
        client: TelegramClient,
        bot_entity,
        codex_api_key: str,
        model: str = "codex-mini-latest",
        base_url: str = "https://api.openai.com/v1",
        max_depth: int = 5,
        step_delay: float = 2.0,
    ):
        self.client = client
        self.bot_entity = bot_entity
        self.codex_api_key = codex_api_key
        self.model = model
        self.base_url = base_url.rstrip("/")
        self.max_depth = max_depth
        self.step_delay = step_delay
        self.visited: set[tuple[str, str]] = set()
        self.http_client = httpx.AsyncClient(timeout=30.0)

    async def close(self):
        """Close the HTTP client."""
        await self.http_client.aclose()

    async def explore(self) -> dict:
        """Explore the bot's navigation tree starting from /start."""
        try:
            tree = await self._explore_node("/start", "command", depth=0)
            return tree
        finally:
            await self.close()

    async def _explore_node(self, action: str, action_type: str, depth: int) -> dict:
        """Recursively explore a single node in the bot's navigation tree."""
        node = {
            "action": action,
            "action_type": action_type,
            "response_text": "",
            "response_media": [],
            "buttons": [],
            "error": None,
            "children": [],
            "metadata": {
                "timestamp": datetime.now(timezone.utc).isoformat(),
                "depth": depth,
                "response_count": 0,
            },
        }

        # Rate limiting
        if depth > 0:
            await asyncio.sleep(self.step_delay)

        # Send action
        try:
            if action_type == "command" or action_type == "button_press":
                await self._send_action(action, action_type)
            elif action_type == "callback_query":
                # Handled separately via message.click()
                pass
        except FloodWaitError as e:
            logger.warning("FloodWaitError: sleeping %ds", e.seconds)
            await asyncio.sleep(e.seconds)
            try:
                await self._send_action(action, action_type)
            except FloodWaitError as e2:
                node["error"] = f"FloodWaitError: {e2.seconds}s (after retry)"
                return node

        # Collect responses
        try:
            responses = await collect_bot_responses(
                self.client, self.bot_entity, timeout=10.0, quiet_period=1.5
            )
        except Exception as e:
            node["error"] = f"collection_error: {e}"
            return node

        node["metadata"]["response_count"] = len(responses)

        if not responses:
            node["error"] = "timeout: no response in 10s"
            return node

        # Extract content from responses
        texts = []
        all_buttons = []
        all_media = []
        last_message_with_markup = None

        for msg in responses:
            if msg.text:
                texts.append(msg.text)
            all_media.extend(_extract_media(msg))
            msg_buttons = _extract_buttons(msg)
            if msg_buttons:
                all_buttons = msg_buttons  # last markup wins
                last_message_with_markup = msg

        node["response_text"] = "\n".join(texts)
        node["response_media"] = all_media
        node["buttons"] = all_buttons

        # Check loop detection
        response_hash = hashlib.md5(node["response_text"].encode()).hexdigest()[:8]
        visit_key = (action, response_hash)
        if visit_key in self.visited:
            return node  # already explored this path
        self.visited.add(visit_key)

        # Depth limit
        if depth >= self.max_depth:
            return node

        # No buttons = leaf node
        if not all_buttons:
            return node

        # Ask Codex what to explore next
        available_labels = [b["label"] for b in all_buttons]
        while True:
            decision = await self._ask_codex(node, available_labels, depth)
            if decision is None or decision.get("action") == "done":
                break

            if decision["action"] == "press_button":
                target = decision["target"]
                if target is None:
                    break

                # Find the button info
                btn_info = next((b for b in all_buttons if b["label"] == target), None)
                if btn_info is None:
                    break

                # Determine action type for the child
                child_action_type = "button_press"
                if btn_info["type"] == "inline_keyboard":
                    if "url" in btn_info:
                        # URL buttons: record but don't click
                        node["children"].append({
                            "action": target,
                            "action_type": "url_button",
                            "response_text": "",
                            "response_media": [],
                            "buttons": [],
                            "error": None,
                            "children": [],
                            "metadata": {
                                "timestamp": datetime.now(timezone.utc).isoformat(),
                                "depth": depth + 1,
                                "response_count": 0,
                                "url": btn_info["url"],
                            },
                        })
                        # Remove from available and continue
                        available_labels = [l for l in available_labels if l != target]
                        if not available_labels:
                            break
                        continue
                    elif "callback_data" in btn_info and last_message_with_markup:
                        # Click the inline button
                        try:
                            await last_message_with_markup.click(
                                data=btn_info["callback_data"].encode()
                                if isinstance(btn_info["callback_data"], str)
                                else btn_info["callback_data"]
                            )
                        except Exception as e:
                            node["children"].append({
                                "action": target,
                                "action_type": "callback_query",
                                "response_text": "",
                                "response_media": [],
                                "buttons": [],
                                "error": f"click_error: {e}",
                                "children": [],
                                "metadata": {
                                    "timestamp": datetime.now(timezone.utc).isoformat(),
                                    "depth": depth + 1,
                                    "response_count": 0,
                                },
                            })
                            available_labels = [l for l in available_labels if l != target]
                            if not available_labels:
                                break
                            continue

                        # Collect response after clicking
                        child_responses = await collect_bot_responses(
                            self.client, self.bot_entity, timeout=10.0, quiet_period=1.5
                        )
                        # Build child node from callback response
                        child_node = self._build_response_node(
                            target, "callback_query", child_responses, depth + 1
                        )
                        node["children"].append(child_node)
                        available_labels = [l for l in available_labels if l != target]
                        if not available_labels:
                            break
                        continue

                # Reply keyboard button: explore recursively
                child = await self._explore_node(target, child_action_type, depth + 1)
                node["children"].append(child)

                # Remove explored button from available
                available_labels = [l for l in available_labels if l != target]
                if not available_labels:
                    break

            elif decision["action"] == "send_text":
                text = decision.get("text", "")
                child = await self._explore_node(text, "command", depth + 1)
                node["children"].append(child)
                break  # free-text sends are one-off

        return node

    def _build_response_node(
        self, action: str, action_type: str, responses: list, depth: int
    ) -> dict:
        """Build a node dict from collected responses."""
        texts = []
        all_media = []
        all_buttons = []
        for msg in responses:
            if msg.text:
                texts.append(msg.text)
            all_media.extend(_extract_media(msg))
            msg_buttons = _extract_buttons(msg)
            if msg_buttons:
                all_buttons = msg_buttons

        return {
            "action": action,
            "action_type": action_type,
            "response_text": "\n".join(texts),
            "response_media": all_media,
            "buttons": all_buttons,
            "error": None if responses else "timeout: no response in 10s",
            "children": [],
            "metadata": {
                "timestamp": datetime.now(timezone.utc).isoformat(),
                "depth": depth,
                "response_count": len(responses),
            },
        }

    async def _send_action(self, action: str, action_type: str):
        """Send a message or press a reply keyboard button."""
        # Both commands and reply keyboard presses are sent as text messages
        await self.client.send_message(self.bot_entity, action)

    async def _ask_codex(
        self, node: dict, available_buttons: list[str], depth: int
    ) -> dict | None:
        """Ask Codex which action to take next. Falls back to sequential exploration if no API key."""
        if not available_buttons:
            return None

        # No API key → sequential BFS exploration of all buttons
        if not self.codex_api_key:
            return {"action": "press_button", "target": available_buttons[0]}

        actions_list = "\n".join(f"- {label}" for label in available_buttons)
        visited_paths = "\n".join(f"- {a} -> {h}" for a, h in self.visited) or "(none)"

        prompt = CODEX_SYSTEM_PROMPT.format(
            response_text=node["response_text"][:500],  # truncate long responses
            actions_list=actions_list,
            visited_paths=visited_paths,
            depth_remaining=self.max_depth - depth,
        )

        try:
            response_text = await self._call_codex(prompt)
            response = json.loads(response_text)
            return validate_codex_response(response, available_buttons)
        except (json.JSONDecodeError, ValueError, RuntimeError) as e:
            logger.warning("Codex error: %s. Falling back to first button.", e)
            return {"action": "press_button", "target": available_buttons[0]}

    async def _call_codex(self, prompt: str, retries: int = 3) -> str:
        """Call the Codex/OpenAI API with retry logic."""
        for attempt in range(retries):
            try:
                resp = await self.http_client.post(
                    f"{self.base_url}/chat/completions",
                    headers={
                        "Authorization": f"Bearer {self.codex_api_key}",
                        "Content-Type": "application/json",
                    },
                    json={
                        "model": self.model,
                        "messages": [{"role": "user", "content": prompt}],
                        "max_tokens": 256,
                        "temperature": 0.1,
                    },
                )
                if resp.status_code == 429 or resp.status_code >= 500:
                    wait = min(2**attempt * 1.0, 30.0)
                    logger.warning(
                        "Codex API %d, retry %d/%d in %.1fs",
                        resp.status_code, attempt + 1, retries, wait,
                    )
                    await asyncio.sleep(wait)
                    continue
                resp.raise_for_status()
                return resp.json()["choices"][0]["message"]["content"]
            except (httpx.TimeoutException, httpx.HTTPStatusError) as e:
                if attempt == retries - 1:
                    raise RuntimeError(f"Codex API failed after {retries} retries: {e}") from e
                wait = min(2**attempt * 1.0, 30.0)
                logger.warning("Codex error: %s, retry in %.1fs", e, wait)
                await asyncio.sleep(wait)
        raise RuntimeError("Codex API failed after all retries")


async def run_flow_walker(
    bot_username: str,
    output_dir: str = "results",
    max_depth: int = 5,
) -> dict:
    """Run the flow walker against a bot and save results."""
    load_dotenv()

    codex_api_key = os.getenv("CODEX_API_KEY", "")
    codex_model = os.getenv("CODEX_MODEL", "codex-mini-latest")
    codex_base_url = os.getenv("CODEX_BASE_URL", "https://api.openai.com/v1")

    async with get_client() as client:
        bot_entity = await client.get_entity(bot_username)

        walker = FlowWalker(
            client=client,
            bot_entity=bot_entity,
            codex_api_key=codex_api_key,
            model=codex_model,
            base_url=codex_base_url,
            max_depth=max_depth,
        )

        logger.info("Starting flow walker for @%s (max_depth=%d)", bot_username, max_depth)
        tree = await walker.explore()

    # Save results
    output_path = Path(output_dir)
    output_path.mkdir(parents=True, exist_ok=True)
    timestamp = datetime.now(timezone.utc).strftime("%Y%m%d_%H%M%S")
    filename = output_path / f"{bot_username}_{timestamp}.json"

    with open(filename, "w", encoding="utf-8") as f:
        json.dump(tree, f, indent=2, ensure_ascii=False)

    logger.info("Results saved to %s", filename)
    return tree


def main():
    """CLI entry point."""
    parser = argparse.ArgumentParser(description="AI-driven Telegram bot flow walker")
    parser.add_argument("bot_username", help="Bot username to explore (e.g. @MyBot)")
    parser.add_argument("--output-dir", default="results", help="Output directory for JSON results")
    parser.add_argument("--max-depth", type=int, default=5, help="Maximum exploration depth")
    parser.add_argument("--verbose", "-v", action="store_true", help="Enable verbose logging")
    args = parser.parse_args()

    logging.basicConfig(
        level=logging.DEBUG if args.verbose else logging.INFO,
        format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
    )

    asyncio.run(run_flow_walker(args.bot_username, args.output_dir, args.max_depth))


if __name__ == "__main__":
    main()
