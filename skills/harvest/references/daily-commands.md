# Harvest Daily Commands

Use this file for `whoami`, `projects`, `recent`, `log create`, `log update`, `log delete`, `today`, and `submit week`.

## Commands

Check auth:

```bash
harvest whoami
harvest whoami --json
```

List active project/task pairs:

```bash
harvest projects
harvest projects --json
```

Show recent entries:

```bash
harvest recent
harvest recent --limit 5 --days 30 --json
```

Create a time entry:

```bash
harvest log create --project "Acme" --task "Development" --duration 1h30m
harvest log create --project "Acme" --task "Development" --duration 1h30m --dry-run
harvest log create --duration 45m --date today --notes "Bug fix"
harvest log create --project "Acme" --task "Development" --duration 1h --dry-run --json
harvest log create --project "Acme" --task "Development" --duration 1h --json
```

Update a time entry:

```bash
harvest log update --id 44 --duration 45m
harvest log update --id 44 --date 2026-03-12
harvest log update --id 44 --notes ""
harvest log update --id 44 --project "Acme" --task "Review" --dry-run --json
```

Delete a time entry:

```bash
harvest log delete --id 44
harvest log delete --id 44 --dry-run --json
```

Review today:

```bash
harvest today
harvest today --json
```

Submit a week:

```bash
harvest submit week
harvest submit week --date today
harvest submit week --date today --dry-run
harvest submit week --date 2026-03-09 --dry-run --json
harvest submit week --date 2026-03-09 --json
```

## Flags

`harvest whoami`

- `--json`

`harvest projects`

- `--json`

`harvest recent`

- `--limit <n>` default `10`
- `--days <n>` default `90`
- `--json`

`harvest log create`

- `--project string`
- `--task string`
- `--duration string`
- `--date string`
- `--notes, -n string`
- `--dry-run`
- `--json`

Short aliases:

- `-p` for `--project`
- `-t` for `--task`
- `-n` for `--notes`

`harvest log update`

- `--id int`
- `--project string`
- `--task string`
- `--duration string`
- `--date string`
- `--notes, -n string`
- `--dry-run`
- `--json`

`harvest log delete`

- `--id int`
- `--dry-run`
- `--json`

`harvest today`

- `--json`

`harvest submit week`

- `--date today|YYYY-MM-DD`
- `--dry-run`
- `--json`

## Output Shape

`harvest whoami`

```text
Ned Tester
Email: ned@example.com
User ID: 1
```

`harvest whoami --json`

```json
{
  "ok": true,
  "user": {
    "id": 1,
    "first_name": "Ned",
    "last_name": "Tester",
    "email": "ned@example.com"
  }
}
```

`harvest projects`

```text
PROJECT  TASK         PROJECT ID  TASK ID
Acme     Design       11          21
Acme     Development  11          22
```

`harvest projects --json`

```json
{
  "ok": true,
  "projects": [
    {
      "project_id": 11,
      "project": "Acme",
      "task_id": 21,
      "task": "Design"
    },
    {
      "project_id": 11,
      "project": "Acme",
      "task_id": 22,
      "task": "Development"
    }
  ]
}
```

`harvest recent`

```text
ID  DATE        PROJECT  TASK         HOURS  NOTES
3   2026-03-11  Acme     Development  1.00   CLI scaffolding
1   2026-03-11  Acme     Review       0.25   PR review
2   2026-03-09  Acme     Design       0.50   Wireframes
```

`harvest recent --json`

```json
{
  "ok": true,
  "from": "2025-12-12",
  "to": "2026-03-11",
  "entries": [
    {
      "id": 3,
      "spent_date": "2026-03-11",
      "hours": 1,
      "project": {
        "id": 11,
        "name": "Acme"
      },
      "task": {
        "id": 22,
        "name": "Development"
      },
      "notes": "CLI scaffolding"
    }
  ]
}
```

`harvest log create`

```text
Logged 1.50h on 2026-03-11 to Acme / Development (#44).
Notes: CLI scaffolding
```

`harvest log create --json`

```json
{
  "ok": true,
  "entry": {
    "id": 44,
    "date": "2026-03-11",
    "hours": 0.75,
    "project_id": 11,
    "project": "Acme",
    "task_id": 22,
    "task": "Development"
  }
}
```

`harvest log create --dry-run`

```text
Dry run: would log 1.50h on 2026-03-11 to Acme / Development.
Notes: CLI scaffolding
```

`harvest log create --dry-run --json`

```json
{
  "ok": true,
  "dry_run": true,
  "entry": {
    "date": "2026-03-11",
    "hours": 1.5,
    "project_id": 11,
    "project": "Acme",
    "task_id": 22,
    "task": "Development",
    "notes": "CLI scaffolding"
  }
}
```

`harvest log update`

```text
Updated entry #44 to 0.75h on 2026-03-12 for Acme / Development.
Notes cleared.
```

`harvest log update --json`

```json
{
  "ok": true,
  "entry": {
    "id": 44,
    "date": "2026-03-12",
    "hours": 0.75,
    "notes": "",
    "project_id": 11,
    "project": "Acme",
    "task_id": 22,
    "task": "Development"
  }
}
```

`harvest log update --dry-run --json`

```json
{
  "ok": true,
  "dry_run": true,
  "entry": {
    "id": 44,
    "date": "2026-03-12",
    "hours": 0.75,
    "notes": "",
    "project_id": 11,
    "project": "Acme",
    "task_id": 22,
    "task": "Development"
  }
}
```

`harvest log delete`

```text
Deleted 0.75h on 2026-03-12 from Acme / Development (#44).
```

`harvest log delete --dry-run --json`

```json
{
  "ok": true,
  "dry_run": true,
  "entry": {
    "id": 44,
    "date": "2026-03-12",
    "hours": 0.75,
    "project_id": 11,
    "project": "Acme",
    "task_id": 22,
    "task": "Development"
  }
}
```

`harvest today`

```text
ID     DATE        PROJECT  TASK         HOURS  NOTES
1      2026-03-11  Acme     Development  1.25   CLI scaffolding
2      2026-03-11  Acme     Review       0.75   PR review
TOTAL                                    2.00
```

`harvest today --json`

```json
{
  "ok": true,
  "date": "2026-03-11",
  "total_hours": 2,
  "entries": [
    {
      "id": 1,
      "hours": 1.25
    },
    {
      "id": 2,
      "hours": 0.75
    }
  ]
}
```

`harvest submit week`

```text
Submitted week 2026-03-09 to 2026-03-15 for approval.
```

`harvest submit week --json`

```json
{
  "ok": true,
  "result": {
    "action": "submitted",
    "week_start": "2026-03-09",
    "week_end": "2026-03-15",
    "return_to": "/time?day=11&month=3&year=2026",
    "submitted_before": false,
    "submitted_after": true
  }
}
```

`harvest submit week --dry-run`

```text
Dry run: would submit week 2026-03-09 to 2026-03-15 for approval.
```

`harvest submit week --dry-run --json`

```json
{
  "ok": true,
  "dry_run": true,
  "result": {
    "action": "would_submit",
    "week_start": "2026-03-09",
    "week_end": "2026-03-15",
    "return_to": "/time?day=11&month=3&year=2026",
    "submitted_before": false
  }
}
```

## Workflow

1. Ask what the user worked on for each day in the target period unless they already gave exact entries or explicitly asked to reuse recent entries.
2. Run `harvest projects --json` if project or task names are unknown.
3. Run `harvest recent --json` if the user wants to reuse a recent pair.
4. Run `harvest log create --dry-run ...`, `harvest log update --dry-run ...`, or `harvest log delete --dry-run ...` when the user wants a preview.
5. Wait about 2 seconds, then verify with `harvest today --json` or `harvest recent --json`.
6. Make sure public API auth is present before `harvest submit week`.
7. Run `harvest submit week --dry-run ...` when the user wants a submit preview.

## Notes

- `harvest projects` filters out inactive projects and tasks.
- Before `harvest log create` or `harvest submit week`, gather the user's day-by-day work details unless exact entries are already known.
- `harvest log create` requires `--duration`.
- `harvest log update` and `harvest log delete` require `--id`.
- Duration uses Go duration strings like `45m`, `1h30m`, and `2h`.
- `--date` defaults to local today.
- `harvest log create` and `harvest log update` accept `today` or `YYYY-MM-DD` for `--date`.
- `--project` and `--task` can come from config defaults or environment variables on `harvest log create`.
- After a create, update, or delete, Harvest can lag briefly on reads. Wait about 2 seconds before using `recent` or `today` to verify the change.
- `submit week` uses the same Harvest account ID and personal access token as the rest of the CLI.
- `submit week --dry-run` reads the week summary without submitting.
