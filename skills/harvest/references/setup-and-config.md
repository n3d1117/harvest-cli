# Harvest Setup And Config

Use this file for auth, config, environment overrides, and config inspection.

## Contents

1. `login`
2. `config`
3. Config path and precedence
4. Common setup errors

## `harvest login`

Use for interactive login.

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

- This command validates the credentials before saving them.
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

Use for non-interactive setup or for changing defaults.

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
    "default_task": "Development"
  }
}
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

## Common Setup Errors

Missing credentials:

```text
error: missing Harvest credentials; run `harvest login` or `harvest config set --account-id ... --token ...`
```
