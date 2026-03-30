# Harvest Help And Errors

Use this file for help entry points, JSON wrappers, and common CLI errors.

## Help Entry Points

Supported help commands:

- `harvest help`
- `harvest help config`
- `harvest help log`
- `harvest help log create`
- `harvest help log update`
- `harvest help log delete`
- `harvest help submit`

Use this reference for `submit auth` details.

## Top-Level Commands

- `login`
- `config`
- `submit`
- `whoami`
- `projects`
- `recent`
- `log`
- `today`
- `help`

`log` subcommands:

- `create`
- `update`
- `delete`

## JSON Wrappers

- `harvest config show --json`: `{ "ok": true, "config_path": "...", "config": { ... } }`
- `harvest submit auth status --json`: `{ "ok": true, "status": { ... } }`
- `harvest submit week --json`: `{ "ok": true, "result": { ... } }`
- `harvest submit week --dry-run --json`: `{ "ok": true, "dry_run": true, "result": { ... } }`
- `harvest whoami --json`: `{ "ok": true, "user": { ... } }`
- `harvest projects --json`: `{ "ok": true, "projects": [ ... ] }`
- `harvest recent --json`: `{ "ok": true, "from": "YYYY-MM-DD", "to": "YYYY-MM-DD", "entries": [ ... ] }`
- `harvest log create --json`: `{ "ok": true, "entry": { ... } }`
- `harvest log create --dry-run --json`: `{ "ok": true, "dry_run": true, "entry": { ... } }`
- `harvest log update --json`: `{ "ok": true, "entry": { ... } }`
- `harvest log update --dry-run --json`: `{ "ok": true, "dry_run": true, "entry": { ... } }`
- `harvest log delete --json`: `{ "ok": true, "entry": { ... } }`
- `harvest log delete --dry-run --json`: `{ "ok": true, "dry_run": true, "entry": { ... } }`
- `harvest today --json`: `{ "ok": true, "date": "YYYY-MM-DD", "total_hours": 0, "entries": [ ... ] }`

## Common Errors

Missing API credentials:

```text
error: missing Harvest credentials; run `harvest login` or `harvest config set --account-id ... --token ...`
```

Missing submit auth:

```text
error: submit auth is not configured; run `harvest submit auth login` first
```

Expired submit auth without a saved password:

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

Missing log subcommand:

```text
error: missing log subcommand
```

Missing entry ID:

```text
error: --id is required
```

Missing update fields:

```text
error: pass at least one field to update
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

Incomplete project/task update:

```text
error: if changing project or task, pass both --project and --task
```
