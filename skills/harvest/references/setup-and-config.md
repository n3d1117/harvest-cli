# Harvest Setup And Config

Use this file for auth, config, environment overrides, submit auth, and setup errors.

## Contents

1. `login`
2. `config`
3. `submit auth`
4. Config path and precedence
5. Secret storage
6. Common setup errors

## Public API Auth vs Submit Auth

The CLI now has two auth paths:

- Public API auth for `whoami`, `projects`, `recent`, `log`, and `today`
- Harvest website auth for `submit`

Why `submit` needs email/password:

- Harvest does not expose submit-for-approval in the public API.
- `harvest submit` signs in to the Harvest website and submits the same private form used by the web UI.
- Submit email/password are sent only to Harvest website endpoints: `id.getharvest.com` and `*.harvestapp.com`.

Storage rules:

- API account ID, API token, defaults, and submit email go in the config file.
- Saved submit passwords and website session cookies go in macOS Keychain.
- Submit secrets do not go in `config.json`.

## `harvest login`

Use for interactive public API auth.

Command:

```bash
harvest login
```

Example session:

```text
Harvest account ID: 123456
Harvest personal access token:
Saved Harvest credentials for Ned Tester.
```

Notes:

- This command validates the API credentials before saving them.
- It stores the config in the user config directory.
- It does not support `--json`.

## `harvest config`

Use to show config subcommands or to route to `set` and `show`.

Usage:

```bash
harvest config <command> [flags]
```

Example output:

```text
Usage:
  harvest config <command> [flags]

Commands:
  set           Save account, token, or defaults
  show          Show effective config without exposing the token

Examples:
  harvest config set --account-id 123456 --token abc123
  harvest config set --default-project "Acme" --default-task "Development"
  harvest config show
```

## `harvest config set`

Use for non-interactive public API setup or for changing defaults.

Usage:

```bash
harvest config set [flags]
```

Flags:

```text
--account-id string
--token string
--default-project string
--default-task string
```

Examples:

```bash
harvest config set --account-id 123456 --token abc123
harvest config set --default-project "Acme" --default-task "Development"
```

Example output:

```text
Saved config to /Users/ned/Library/Application Support/harvest/config.json.
Defaults: project="Acme" task="Development"
```

Notes:

- At least one flag is required.
- This command does not support `--json`.

## `harvest config show`

Use to inspect the effective config after environment overrides are applied.

Usage:

```bash
harvest config show [--json]
```

Human example:

```bash
harvest config show
```

Human output:

```text
Config file: /Users/ned/Library/Application Support/harvest/config.json
Account ID: 123456
Token: present
Default project: Acme
Default task: Development
Submit email: ned@example.com
```

JSON example:

```bash
harvest config show --json
```

JSON example output:

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

## `harvest submit auth`

Use this group for Harvest website auth used by `harvest submit`.

Subcommands:

- `harvest submit auth login`
- `harvest submit auth status`
- `harvest submit auth logout`

## `harvest submit auth login`

Use to create or refresh Harvest website submit auth.

Usage:

```bash
harvest submit auth login [--email <email>] [--save-password]
```

Flags:

```text
--email string
--save-password
```

Examples:

```bash
harvest submit auth login
harvest submit auth login --email you@example.com
harvest submit auth login --email you@example.com --save-password
```

Example output:

```text
Saved Harvest submit auth for Ned Tester.
Submit session expires: 2026-03-26T18:13:01Z
Password saved in macOS Keychain.
```

Notes:

- Harvest website auth is separate from the public API token.
- Without `--save-password`, the CLI saves only the current website session.
- With `--save-password`, the CLI can refresh the website session silently after the shorter-lived session cookie expires.

## `harvest submit auth status`

Use to inspect Harvest website submit auth.

Usage:

```bash
harvest submit auth status [--json]
```

Human output:

```text
Submit email: ned@example.com
Harvest base URL: https://shapegames.harvestapp.com
Session: saved (expires 2026-03-26T18:13:01Z)
Password: saved
Access token expires: 2026-05-10T18:13:00Z
```

JSON example output:

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

## `harvest submit auth logout`

Use to delete saved Harvest website submit auth.

Usage:

```bash
harvest submit auth logout
```

Human output:

```text
Removed saved Harvest submit auth.
```

## Config Path And Precedence

Default config path:

```text
~/Library/Application Support/harvest/config.json
```

Supported environment overrides:

- `HARVEST_ACCOUNT_ID`
- `HARVEST_TOKEN`
- `HARVEST_DEFAULT_PROJECT`
- `HARVEST_DEFAULT_TASK`

Precedence:

1. command flags
2. environment variables
3. config file

## Secret Storage

Submit secrets live in macOS Keychain:

- Harvest website password, if saved with `--save-password`
- Harvest website session cookies

Observed Harvest website cookie lifetimes from a live login on 2026-03-11:

- `_harvest_sess`: about 15 days
- `production_access_token`: about 60 days

These are private website cookies and may change.

## Common Setup Errors

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
