# Harvest Setup And Config

Use this file for `harvest login`, `harvest config`, and submit prerequisites.

## Commands

```bash
harvest login
harvest config set --account-id 123456 --token abc123
harvest config set --default-project "Acme" --default-task "Development"
harvest config show
harvest config show --json
harvest whoami
```

## Flags

`harvest config set`

- `--account-id string`
- `--token string`
- `--default-project string`
- `--default-task string`

`harvest config show`

- `--json`

## Output Shape

`harvest config show`

```text
Config file: /Users/ned/Library/Application Support/harvest/config.json
Account ID: 123456
Token: present
Default project: Acme
Default task: Development
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
    "default_task": "Development"
  }
}
```

## Workflow

1. Run `harvest whoami` or `harvest config show`.
2. If public API auth is missing, run `harvest login` or `harvest config set`.
3. Set `--default-project` and `--default-task` if you log to the same pair often.
4. Before `harvest submit week`, make sure the account ID and token are set.

## Config

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

Public API commands and `submit week` use your Harvest account ID and personal access token.

## Common Errors

Missing API credentials:

```text
error: missing Harvest credentials; run `harvest login` or `harvest config set --account-id ... --token ...`
```
