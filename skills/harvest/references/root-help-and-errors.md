# Harvest Errors And JSON Summary

Use this file for top-level JSON wrapper summaries and common CLI error strings.

If you need the real help text, run `harvest help`, `harvest help config`, or `harvest help submit`.

Nested help topics such as `harvest help config show` and `harvest help submit auth` are not supported.

## Top-Level Commands

Current top-level commands:

- `login`
- `config`
- `submit`
- `whoami`
- `projects`
- `recent`
- `log`
- `today`
- `help`

## JSON Shapes

These are the top-level JSON wrappers:

- `config show --json`: `{ "ok": true, "config_path": "...", "config": { ... } }`
- `submit auth status --json`: `{ "ok": true, "status": { ... } }`
- `submit week --json`: `{ "ok": true, "result": { ... } }`
- `whoami --json`: `{ "ok": true, "user": { ... } }`
- `projects --json`: `{ "ok": true, "projects": [ ... ] }`
- `recent --json`: `{ "ok": true, "from": "YYYY-MM-DD", "to": "YYYY-MM-DD", "entries": [ ... ] }`
- `log --json`: `{ "ok": true, "entry": { ... } }`
- `today --json`: `{ "ok": true, "date": "YYYY-MM-DD", "total_hours": 0, "entries": [ ... ] }`

## Common Errors

These strings come from the CLI and are safe to quote back to the user.

Missing API credentials:

```text
error: missing Harvest credentials; run `harvest login` or `harvest config set --account-id ... --token ...`
```

Missing submit auth:

```text
error: submit auth is not configured; run `harvest submit auth login` first
```

Expired submit auth without saved password:

```text
error: submit auth expired; run `harvest submit auth login` again or save a password with `--save-password`
```

Missing submit account ID:

```text
error: submit needs a Harvest account ID; run `harvest login` or `harvest config set --account-id ...` first
```

Missing project:

```text
error: project is required; pass --project or set a default project
```

Missing task:

```text
error: task is required; pass --task or set a default task
```

Missing duration:

```text
error: --duration is required
```

Bad date:

```text
error: date must use YYYY-MM-DD or `today`
```

Bad recent limit:

```text
error: --limit must be greater than zero
```

Bad recent days:

```text
error: --days must be greater than zero
```

Ambiguous project:

```text
error: project "Acme" is ambiguous: Acme (#11), Acme (#18)
```

Ambiguous task:

```text
error: task "Development" is ambiguous under project "Acme": Development (#22), Development (#29)
```
