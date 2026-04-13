# Revisitr Telegram

Telethon infrastructure for E2E bot testing and AI-driven competitor analysis.

## Purpose

- **E2E Testing**: Send real Telegram messages to your bot and verify responses programmatically
- **Competitor Analysis**: AI-driven flow walker that explores any bot's navigation tree with Codex deciding which buttons to press

## Prerequisites

- Python 3.11+
- [uv](https://docs.astral.sh/uv/) package manager
- Telegram API credentials (`api_id` + `api_hash` from [my.telegram.org](https://my.telegram.org))
- Authorized Telethon `.session` file
- (For flow walker) OpenAI/Codex API key

## Setup

```bash
cd telegram/

# 1. Configure credentials
cp .env.example .env
# Edit .env with your values

# 2. Place your .session file
# The file should be named to match SESSION_NAME in .env (e.g., "my_account.session")
cp /path/to/your.session ./

# 3. Install dependencies
uv sync --dev
```

## Running Tests

```bash
# Run all tests
uv run pytest tests/ -v

# Run specific test
uv run pytest tests/test_own_bot.py -v
```

The test sends `/start` to your bot and verifies it responds. Response collection is event-driven (not sleep-based) to handle multi-part messages correctly.

## Running Flow Walker

```bash
# Explore a bot's navigation tree
uv run python -m revisitr_telegram.flow_walker @BotUsername

# With options
uv run python -m revisitr_telegram.flow_walker @BotUsername --max-depth 3 --verbose
```

Results are saved as JSON to `results/{bot_username}_{timestamp}.json`.

## .env Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `API_ID` | Yes | Telegram API ID from my.telegram.org |
| `API_HASH` | Yes | Telegram API hash from my.telegram.org |
| `SESSION_NAME` | Yes | Name of `.session` file (without extension) |
| `BOT_USERNAME` | For tests | Your bot's @username |
| `CODEX_API_KEY` | For walker | OpenAI API key |
| `CODEX_MODEL` | No | Model name (default: `codex-mini-latest`) |
| `CODEX_BASE_URL` | No | API base URL (default: `https://api.openai.com/v1`) |

## Flow Walker Output Format

```json
{
  "action": "/start",
  "action_type": "command",
  "response_text": "Welcome!",
  "response_media": [{"type": "photo", "caption": "Welcome!", "has_media": true}],
  "buttons": [
    {"label": "Balance", "type": "reply_keyboard"},
    {"label": "Visit site", "type": "inline_keyboard", "url": "https://..."},
    {"label": "Details", "type": "inline_keyboard", "callback_data": "details_1"}
  ],
  "error": null,
  "children": [],
  "metadata": {"timestamp": "...", "depth": 0, "response_count": 3}
}
```

**Button types**: `reply_keyboard` (text sent as message), `inline_keyboard` with `url` (recorded, not clicked), `inline_keyboard` with `callback_data` (clicked via Telethon).

**Error field**: `null` on success, or a description like `"FloodWaitError: 30s"`, `"timeout: no response in 10s"`.

## Creating a Session File

One-time manual step (requires SMS code):

```bash
cd telegram/
uv run python -c "
from telethon.sync import TelegramClient
import os
from dotenv import load_dotenv
load_dotenv()
client = TelegramClient(os.environ['SESSION_NAME'], int(os.environ['API_ID']), os.environ['API_HASH'])
client.start()
print('Session created:', client.session.filename)
client.disconnect()
"
```

## Token Budget

Flow walker uses ~500 tokens per navigation step. At max_depth=5 with branching factor ~4, worst case ~100 steps = ~50K tokens per exploration run.
