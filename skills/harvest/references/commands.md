# Harvest Command Index

Use this file when you only need the command surface.

## Setup

- `harvest login`
- `harvest config set`
- `harvest config show`

## Daily Work

- `harvest whoami`
- `harvest projects`
- `harvest recent`
- `harvest log create`
- `harvest log update`
- `harvest log delete`
- `harvest today`

## Submit

- `harvest submit auth login`
- `harvest submit auth status`
- `harvest submit auth logout`
- `harvest submit week`

## Help

- `harvest help`
- `harvest help config`
- `harvest help log`
- `harvest help log create`
- `harvest help log update`
- `harvest help log delete`
- `harvest help submit`

## Notes

- `config show`, `whoami`, `projects`, `recent`, `log create`, `log update`, `log delete`, `today`, `submit auth status`, and `submit week` support `--json`.
- `harvest log create`, `harvest log update`, `harvest log delete`, and `harvest submit week` also support `--dry-run`.
- Public API commands use Harvest account ID and personal access token.
- `submit` uses Harvest website auth.
- Saved submit passwords and website session cookies live in macOS Keychain.
