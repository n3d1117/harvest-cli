# Harvest Setup And Config

Use this file for `harvest login`, `harvest config`, and `harvest submit auth`.

## Commands

Public API setup:

```bash
harvest login
harvest config set --account-id 123456 --token abc123
harvest config set --default-project "Acme" --default-task "Development"
harvest config show
harvest config show --json
```

Submit auth setup:

```bash
harvest submit auth login
harvest submit auth login --email you@example.com
harvest submit auth login --email you@example.com --save-password
harvest submit auth status
harvest submit auth status --json
harvest submit auth logout
```

## Flags

`harvest config set`

- `--account-id string`
- `--token string`
- `--default-project string`
- `--default-task string`

`harvest config show`

- `--json`

`harvest submit auth login`

- `--email string`
- `--save-password`

`harvest submit auth status`

- `--json`

## Output Shape

`harvest config show`

```text
Config file: /Users/ned/Library/Application Support/harvest/config.json
Account ID: 123456
Token: present
Default project: Acme
Default task: Development
Submit email: ned@example.com
```

`harvest config show --json`

```json
{
  "ok": true,
  "config_path": "/Users/ned/Library/Application Support/harvest/config.json",
  "config": {
    "account_id": "123456",
    "token_present": true,
    "default_project": "Acme",
    "default_task": "Development",
    "submit_email": "ned@example.com"
  }
}
```

`harvest submit auth login`

```text
Saved Harvest submit auth for Ned Tester.
Submit session expires: 2026-03-26T18:13:01Z
Password saved in macOS Keychain.
```

`harvest submit auth status`

```text
Submit email: ned@example.com
Harvest base URL: https://shapegames.harvestapp.com
Session: saved (expires 2026-03-26T18:13:01Z)
Password: saved
Access token expires: 2026-05-10T18:13:00Z
```

`harvest submit auth status --json`

```json
{
  "ok": true,
  "status": {
    "email": "ned@example.com",
    "base_url": "https://shapegames.harvestapp.com",
    "session_saved": true,
    "session_expires_at": "2026-03-26T18:13:01Z",
    "password_saved": true,
    "access_token_expires_at": "2026-05-10T18:13:00Z"
  }
}
```

## Workflow

1. Run `harvest whoami` or `harvest config show`.
2. If public API auth is missing, run `harvest login` or `harvest config set`.
3. Set `--default-project` and `--default-task` if you log to the same pair often.
4. If the user needs `submit`, run `harvest submit auth login`.
5. Check `harvest submit auth status` before `harvest submit week`.

## Config And Secrets

Config file:

```text
~/Library/Application Support/harvest/config.json
```

Environment overrides:

- `HARVEST_ACCOUNT_ID`
- `HARVEST_TOKEN`
- `HARVEST_DEFAULT_PROJECT`
- `HARVEST_DEFAULT_TASK`

Precedence:

1. command flags
2. environment variables
3. config file

Public API commands use your Harvest account ID and personal access token.
`submit` uses Harvest website auth because Harvest has no public submit-for-approval API.

Saved in macOS Keychain:

- Harvest website password when `--save-password` is used
- Harvest website session cookies

Observed Harvest website cookie lifetimes from a live login on 2026-03-11:

- `_harvest_sess`: about 15 days
- `production_access_token`: about 60 days

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
