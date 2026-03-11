# Harvest Daily Commands

Use this file for day-to-day CLI work: auth checks, listing projects, recent entries, logging time, reviewing today's entries, and submitting a week for approval.

## Contents

1. `whoami`
2. `projects`
3. `recent`
4. `log`
5. `today`
6. `submit week`

## `harvest whoami`

Use to verify public API auth and show the current Harvest user.

Usage:

```bash
harvest whoami [--json]
```

Human example:

```bash
harvest whoami
```

Human output:

```text
Ned Tester
Email: ned@example.com
User ID: 1
```

JSON example output:

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

## `harvest projects`

Use to list active project/task pairs you are allowed to log against.

Usage:

```bash
harvest projects [--json]
```

Human output:

```text
PROJECT  TASK         PROJECT ID  TASK ID
Acme     Design       11          21
Acme     Development  11          22
```

JSON example output:

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

Notes:

- Inactive projects and inactive tasks are filtered out.
- Use this command to resolve exact project/task names before calling `harvest log`.

## `harvest recent`

Use to inspect recent entries and reuse a known-good project/task pair.

Usage:

```bash
harvest recent [--limit <n>] [--days <n>] [--json]
```

Defaults:

- `--limit 10`
- `--days 90`

Human output:

```text
DATE        PROJECT  TASK         HOURS  NOTES
2026-03-11  Acme     Development  1.00   CLI scaffolding
2026-03-11  Acme     Review       0.25   PR review
2026-03-09  Acme     Design       0.50   Wireframes
```

JSON example output:

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

## `harvest log`

Use to create a time entry.

Usage:

```bash
harvest log --project <name> --task <name> --duration <duration> [flags]
```

Flags:

```text
--project string
--task string
--duration string
--date string
--notes, -n string
--json
```

Short aliases:

- `-p` for `--project`
- `-t` for `--task`
- `-n` for `--notes`

Rules:

- `--duration` is required.
- Duration uses Go duration strings like `45m`, `1h30m`, and `2h`.
- `--date` defaults to local today.
- `--date` accepts `today` or `YYYY-MM-DD`.
- `--project` and `--task` can come from config defaults or env vars.

Human example:

```bash
harvest log \
  --project "Acme" \
  --task "Development" \
  --duration 1h30m \
  --date today \
  --notes "CLI scaffolding"
```

Human output:

```text
Logged 1.50h on 2026-03-11 to Acme / Development (#44).
Notes: CLI scaffolding
```

JSON example output:

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

## `harvest today`

Use to review today's time entries and total.

Usage:

```bash
harvest today [--json]
```

Human output:

```text
DATE        PROJECT  TASK         HOURS  NOTES
2026-03-11  Acme     Development  1.25   CLI scaffolding
2026-03-11  Acme     Review       0.75   PR review
TOTAL                             2.00
```

JSON example output:

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

## `harvest submit week`

Use to submit the week that contains a date for approval.

This command needs Harvest website submit auth first:

```bash
harvest submit auth login --email you@example.com --save-password
```

Usage:

```bash
harvest submit week [--date today|YYYY-MM-DD] [--json]
```

Human examples:

```bash
harvest submit week
harvest submit week --date 2026-03-09
```

Human output:

```text
Submitted week 2026-03-09 to 2026-03-15 for approval.
```

JSON example output:

```json
{
  "ok": true,
  "result": {
    "action": "submitted",
    "week_start": "2026-03-09",
    "week_end": "2026-03-15",
    "return_to": "/time/day/2026/3/11/4833590",
    "submitted_before": false,
    "submitted_after": true
  }
}
```

Notes:

- This uses Harvest website auth, not the public API token.
- Harvest website email/password are needed because Harvest does not expose submit-for-approval in the public API.
- If the saved website session expires and a password is saved in Keychain, the CLI refreshes the session silently before submitting.
