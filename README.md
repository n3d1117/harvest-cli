![Harvest logo: timesheet card](assets/harvest-logo.svg)

# harvest

`harvest` is a small CLI for logging time to Harvest and submitting a week for approval.

The repo also contains an installable agent skill in [`skills/harvest`](skills/harvest/SKILL.md). You can install that skill in tools such as Claude Code, Codex, Cursor, Gemini CLI, and other Agent Skills-compatible tools. The skill tells an agent to use the `harvest` CLI instead of talking to Harvest directly.

## Install the CLI

```bash
brew install n3d1117/harvest-cli/harvest
```

## Install the skill

```bash
npx skills add n3d1117/harvest-cli --skill harvest
```

## Use the skill

How you invoke the skill depends on the host tool.

- Claude Code: `/harvest`
- Codex: `$harvest`
- Natural language: `Use the harvest skill to log 1h on Acme / Development today`

## First Use

Get your Harvest account ID and token:

1. Go to `https://id.getharvest.com/developers`
2. Click `Create new personal access token`
3. Give it a name
4. Copy the account ID and the token

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

## Submit Approval Setup

Harvest does not provide a public API for submitting time for approval.

Because of that, `harvest submit` uses normal Harvest website auth, not the personal access token used by the public API commands.

- `harvest login`, `harvest projects`, `harvest log`, and `harvest today` use your Harvest account ID and personal access token.
- `harvest submit ...` uses your Harvest website email and password to create a normal Harvest website session and then submits the same private form used by the web UI.
- Your submit email and password are sent only to Harvest login endpoints: `id.getharvest.com` and `*.harvestapp.com`.
- Saved submit secrets go to macOS Keychain, not `config.json`.

Create submit auth:

```bash
harvest submit auth login --email you@example.com --save-password
```

Check submit auth status:

```bash
harvest submit auth status
```

Observed Harvest website cookie lifetimes from a live login on 2026-03-11:

- `_harvest_sess`: about 15 days
- `production_access_token`: about 60 days

These are private website cookies and can change at any time. If you save your password with `--save-password`, the CLI can refresh the website session silently when the shorter-lived session cookie expires.

## Daily Use

Show project/task pairs you can log against:

```bash
harvest projects
```

See recent entries so you can reuse the same project/task:

```bash
harvest recent
```

Log time:

```bash
harvest log \
  --project "Acme" \
  --task "Development" \
  --duration 1h30m \
  --date today \
  --notes "CLI scaffolding"
```

See today:

```bash
harvest today
```

Submit the week that contains a given date:

```bash
harvest submit week --date today
```

## JSON Mode

Commands that return data also support `--json`.

```bash
harvest projects --json
harvest log --project "Acme" --task "Development" --duration 1h --json
harvest today --json
harvest submit auth status --json
harvest submit week --date 2026-03-09 --json
```

## Config location

The CLI stores config here:

```text
~/Library/Application Support/harvest/config.json
```

Saved submit passwords and website session cookies are stored in macOS Keychain, not in the config file.

Supported environment overrides:

- `HARVEST_ACCOUNT_ID`
- `HARVEST_TOKEN`
- `HARVEST_DEFAULT_PROJECT`
- `HARVEST_DEFAULT_TASK`

Precedence is:

1. command flags
2. environment variables
3. config file

## Development

```bash
cd cli
go test ./...
go build ./cmd/harvest
```
