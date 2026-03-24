---
name: harvest
description: Use when working with the `harvest` CLI to log time, check auth, inspect config, list active project/task pairs, reuse recent entries, review today's Harvest time entries, or submit a week for approval. Prefer the CLI over direct Harvest API calls. Read the specific file in `references/` when you need exact flags, output shapes, or example output.
---

# Harvest

Use the `harvest` CLI.

Assume `harvest` is on `PATH`. If it is missing, build or install it first.

## Skill Layout

- `SKILL.md`: workflow and guardrails
- `agents/openai.yaml`: skill metadata
- `references/commands.md`: command index
- `references/setup-and-config.md`: login, config, env, and submit auth
- `references/daily-commands.md`: day-to-day commands
- `references/root-help-and-errors.md`: help entry points, JSON wrappers, and common errors

Open only the file you need.

## Default Workflow

1. Start with `harvest whoami` or `harvest config show`.
2. If public API auth is missing, run `harvest login` or `harvest config set`.
3. If project or task names are unknown, run `harvest projects --json`.
4. If the user wants a recent pair, run `harvest recent --json`.
5. Preview with `harvest log --dry-run ...` or log time with `harvest log ...`.
6. Verify with `harvest today --json`.
7. For approval submit, run `harvest submit auth status`.
8. If submit auth is missing or expired, run `harvest submit auth login`.
9. Preview with `harvest submit week --dry-run ...` or submit with `harvest submit week --date ...`.

## Rules

- Prefer the CLI over direct API calls.
- Prefer `--json` when another tool or agent will read the result.
- Do not invent Harvest IDs or project/task names. Resolve them from `harvest projects --json`.
- Treat missing credentials, ambiguous matches, invalid durations, and invalid dates as blockers.
- `harvest log` accepts Go duration strings like `45m`, `1h30m`, and `2h`.
- `harvest log` accepts `--date today` or `--date YYYY-MM-DD`.
- `harvest log` can use default project/task values from config or environment variables.
- `harvest log --dry-run` resolves the exact project/task pair and prints the entry without creating it.
- Public API commands use Harvest account ID and personal access token.
- `harvest submit` uses Harvest website auth.
- `harvest submit week --dry-run` validates submit auth and resolves the real week window without sending the final submit request.
- Saved submit passwords and website session cookies live in macOS Keychain.

## Reference Guide

Read [`references/commands.md`](references/commands.md) when you need:

- the command surface at a glance
- the command groups

Read [`references/setup-and-config.md`](references/setup-and-config.md) when you need:

- `login`, `config`, or `submit auth`
- config path, environment overrides, or precedence
- setup output or setup errors

Read [`references/daily-commands.md`](references/daily-commands.md) when you need:

- `whoami`, `projects`, `recent`, `log`, `today`, or `submit week`
- example text output
- example JSON output

Read [`references/root-help-and-errors.md`](references/root-help-and-errors.md) when you need:

- supported help entry points
- top-level JSON wrappers
- common CLI error strings
