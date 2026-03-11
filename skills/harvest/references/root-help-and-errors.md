# Harvest Errors And JSON Summary

Use this file for top-level JSON wrapper summaries and common CLI error strings.

If you need the real help text, run `harvest help` or `harvest help <command>`.

## JSON Shapes

These are the top-level JSON wrappers:

- `config show --json`: `{ "ok": true, "config_path": "...", "config": { ... } }`
- `whoami --json`: `{ "ok": true, "user": { ... } }`
- `projects --json`: `{ "ok": true, "projects": [ ... ] }`
- `recent --json`: `{ "ok": true, "from": "YYYY-MM-DD", "to": "YYYY-MM-DD", "entries": [ ... ] }`
- `log --json`: `{ "ok": true, "entry": { ... } }`
- `today --json`: `{ "ok": true, "date": "YYYY-MM-DD", "total_hours": 0, "entries": [ ... ] }`

## Common Errors

These strings come from the CLI and are safe to quote back to the user.

Missing credentials:

```text
error: missing Harvest credentials; run `harvest login` or `harvest config set --account-id ... --token ...`
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
