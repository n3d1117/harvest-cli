---
name: harvest
description: Use when working with the `harvest` CLI to log time, check auth, inspect config, list active project/task pairs, reuse recent entries, review today's Harvest time entries, or submit a week for approval. Prefer the CLI over direct Harvest API calls. Read the specific file in `references/` when you need exact flags, output shapes, or example output.
---

# Harvest

Use the `harvest` CLI, not direct Harvest API calls.

Assume `harvest` is on `PATH`. If it is not, install or build it first.

## Skill Layout

- `SKILL.md`: trigger rules, workflow, and guardrails.
- `agents/openai.yaml`: UI metadata for the skill.
- `references/setup-and-config.md`: login, config, env overrides, and setup errors.
- `references/daily-commands.md`: day-to-day commands and example output.
- `references/commands.md`: short command index.
- `references/root-help-and-errors.md`: root help, JSON wrapper summary, and common errors.

Open only the reference file you need.

## Default Workflow

1. Check auth first with `harvest whoami` or `harvest config show`.
2. If credentials are missing, use `harvest login` or `harvest config set`.
3. If project or task names are unknown, run `harvest projects --json`.
4. If the user wants to reuse a recent pair, run `harvest recent --json`.
5. Log time with `harvest log ...`.
6. Verify the result with `harvest today --json`.
7. For approval submit, run `harvest submit auth status`.
8. If submit auth is missing or expired, run `harvest submit auth login`.
9. Submit a week with `harvest submit week --date ...`.

## Rules

- Prefer the CLI over direct API calls.
- Prefer `--json` when another tool or agent will read the result.
- Do not invent Harvest IDs. Resolve project/task pairs from `harvest projects --json`.
- Treat missing credentials, missing project/task names, invalid durations, and ambiguous project/task matches as real blockers.
- Treat missing submit auth as a real blocker for `harvest submit`.
- `harvest log` accepts Go duration strings like `45m`, `1h30m`, and `2h`.
- `harvest log` accepts `--date today` or `--date YYYY-MM-DD`.
- `harvest log` can use default project/task values from config or environment variables.
- `harvest submit week` uses Harvest website auth, not the public API token.
- Harvest website email and password are needed for submit because Harvest does not expose submit-for-approval in the public API.
- Submit credentials are sent only to Harvest website endpoints and saved secrets live in macOS Keychain.

## When To Read The Reference

Read the relevant reference when you need any of the following:

- [`references/setup-and-config.md`](references/setup-and-config.md) for login, config, and env rules
- [`references/daily-commands.md`](references/daily-commands.md) for command usage and example output
- [`references/commands.md`](references/commands.md) for a short command index
- [`references/root-help-and-errors.md`](references/root-help-and-errors.md) for root help, JSON wrappers, and common errors
