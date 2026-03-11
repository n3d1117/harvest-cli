![Harvest logo: timesheet card](assets/harvest-logo.svg)

# harvest

`harvest` is a small local CLI for logging time to Harvest.

The repo also contains an installable agent skill in [`skills/harvest-log`](/Users/ned/Downloads/harvest/skills/harvest-log/SKILL.md). The skill is generic: it tells an agent to use the local `harvest` CLI instead of talking to Harvest directly.

## Install

```bash
brew install n3d1117/harvest-cli/harvest
```

That path works after the repo is public and the first tagged release has updated `Formula/harvest.rb`.

## First Use

Interactive login for humans:

```bash
harvest login
```

Non-interactive config:

```bash
harvest config set \
  --account-id 123456 \
  --token YOUR_TOKEN \
  --default-project "Acme" \
  --default-task "Development"
```

Check auth:

```bash
harvest whoami
```

## Daily Use

Show project/task pairs you can log against:

```bash
harvest projects
```

Log time:

```bash
harvest log \
  --project "Acme" \
  --task "Development" \
  --duration 1h30m \
  --notes "CLI scaffolding"
```

Notes are optional:

```bash
harvest log \
  --project "Acme" \
  --task "Development" \
  --duration 45m
```

See today:

```bash
harvest today
```

## JSON Mode

Commands that return data also support `--json`.

```bash
harvest projects --json
harvest log --project "Acme" --task "Development" --duration 1h --json
harvest today --json
```

## Config location

The CLI stores config here:

```text
~/Library/Application Support/harvest/config.json
```

Supported environment overrides:

- `HARVEST_ACCOUNT_ID`
- `HARVEST_TOKEN`
- `HARVEST_DEFAULT_PROJECT`
- `HARVEST_DEFAULT_TASK`

Precedence is:

1. command flags
2. environment variables
3. config file

## Installing the skill

```bash
npx skills add n3d1117/harvest-cli --skill harvest-log
```

## Development

```bash
cd cli
go test ./...
go build ./cmd/harvest
```
