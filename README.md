# harvest

`harvest` is a small local-first CLI for logging time to Harvest.

It is built for one main flow:

1. log in once
2. inspect valid project/task pairs
3. log time fast
4. verify what got logged today

The repo also contains an installable agent skill in [`skills/harvest-log`](/Users/ned/Downloads/harvest/skills/harvest-log/SKILL.md). The skill is generic: it tells an agent to use the local `harvest` CLI instead of talking to Harvest directly.

## Why this exists

Harvest already has APIs and a few community CLIs. This project stays intentionally small:

- one Go binary
- local config file
- human-friendly output by default
- `--json` for agents and scripts
- no timers, no aliases, no interactive pickers in v1

## Requirements

- Go 1.26+
- A Harvest account ID
- A Harvest personal access token

Go is installed locally on this machine already:

```bash
go version
```

## Build

```bash
cd cli
go build -o ../bin/harvest ./cmd/harvest
```

Or run it without building:

```bash
cd cli
go run ./cmd/harvest --help
```

## First-time setup

Interactive login for humans:

```bash
cd cli
go run ./cmd/harvest login
```

Non-interactive config:

```bash
cd cli
go run ./cmd/harvest config set \
  --account-id 123456 \
  --token YOUR_TOKEN \
  --default-project "Acme" \
  --default-task "Development"
```

Check auth:

```bash
cd cli
go run ./cmd/harvest whoami
```

## Daily commands

Show project/task pairs you can log against:

```bash
cd cli
go run ./cmd/harvest projects
```

Log time:

```bash
cd cli
go run ./cmd/harvest log \
  --project "Acme" \
  --task "Development" \
  --duration 1h30m \
  --notes "CLI scaffolding"
```

Notes are optional:

```bash
cd cli
go run ./cmd/harvest log \
  --project "Acme" \
  --task "Development" \
  --duration 45m
```

See today:

```bash
cd cli
go run ./cmd/harvest today
```

## Agent-friendly output

Commands that return data also support `--json`.

Examples:

```bash
cd cli
go run ./cmd/harvest projects --json
go run ./cmd/harvest log --project "Acme" --task "Development" --duration 1h --json
go run ./cmd/harvest today --json
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

## Tests

```bash
cd cli
go test ./...
```

## Homebrew release scaffolding

This repo includes:

- [`.goreleaser.yaml`](/Users/ned/Downloads/harvest/.goreleaser.yaml)
- [`.github/workflows/release.yml`](/Users/ned/Downloads/harvest/.github/workflows/release.yml)

The setup assumes:

1. this repo will publish GitHub Releases
2. a separate Homebrew tap repo exists, for example `YOUR_GITHUB_USER/homebrew-tap`
3. the tap repo will receive the generated formula on each tagged release

Required GitHub repository settings:

- repository variable `HOMEBREW_TAP_OWNER`
- repository variable `HOMEBREW_TAP_REPO`
- repository secret `HOMEBREW_TAP_GITHUB_TOKEN`

The token must be able to write to the tap repo.

Release flow:

```bash
git tag v0.1.0
git push origin v0.1.0
```

That workflow will:

1. run tests
2. build release archives
3. publish a GitHub Release
4. update the Homebrew formula in the tap repo

## Installing the skill

The skill lives in the standard repo layout under `skills/`.

The expected install shape is:

```bash
npx skills add n3d1117/harvest-cli --skill harvest-log
```

If your tool uses a slightly different installer command, the important part is that the skill is in `skills/harvest-log/SKILL.md`.
