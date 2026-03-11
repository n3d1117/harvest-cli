---
name: harvest-log
description: Use when logging time to Harvest, checking today's Harvest entries, or working through a local Harvest CLI. Prefer the local `harvest` command over direct Harvest API calls. Use `--json` when structured output is useful for another tool or agent.
---

# Harvest Log

Prefer the local `harvest` CLI over direct Harvest API calls.

Assume `harvest` is available on `PATH`. If it is not, tell the user to install or build it first.

## Workflow

1. Verify auth first:

```bash
harvest whoami
```

2. If the project or task is unknown, inspect valid pairs:

```bash
harvest projects --json
```

3. If the user wants to reuse a recent project/task, inspect recent entries:

```bash
harvest recent --json
```

4. Log time:

```bash
harvest log --project "Acme" --task "Development" --duration 1h30m --notes "CLI scaffolding"
```

5. Verify same-day results:

```bash
harvest today --json
```

## Rules

- Prefer the CLI over direct API calls.
- Prefer `--json` when another tool or agent needs structured output.
- If auth is missing, ask the user to run `harvest login` or `harvest config set`.
- If project or task names are ambiguous, stop and ask for a more specific value.
- Do not invent Harvest IDs. Resolve names through `harvest projects --json`.

## Useful examples

Human-first:

```bash
harvest log --project "Acme" --task "Development" --duration 45m
```

Structured:

```bash
harvest log --project "Acme" --task "Development" --duration 1h --json
```
