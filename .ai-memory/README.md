# .ai-memory — Shared AI Agent Memory

Project knowledge for AI agents (Claude Code, OMC, Codex).
This directory is checked into git so any contributor gets full context.

## Structure

| File | Content |
|------|---------|
| `patterns.md` | Established code patterns (backend + frontend) |
| `gotchas.md` | Known pitfalls, common bugs, deployment traps |
| `decisions.md` | Key architecture and business logic decisions |
| `status.md` | Current dev status, phases, pending work |
| `testing.md` | Test coverage map, testing patterns |

## Usage

- `CLAUDE.md` references this directory — agents auto-load it
- OMC project memory (`.omc/project-memory.json`) handles auto-detection separately
- Keep files updated as the project evolves
- Only store non-obvious knowledge: patterns, gotchas, decisions that can't be derived from reading code

## Maintenance

- Verify facts against current code before trusting — memory can go stale
- After major changes (new migration, architecture shift), update relevant files
- Remove entries that are no longer true
