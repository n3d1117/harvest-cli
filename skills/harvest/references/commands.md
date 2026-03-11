# Harvest Command Index

Use this file when you only need the command surface at a glance.

## Commands

- `harvest help`
- `harvest config`
- `harvest submit`
- `harvest login`
- `harvest config set`
- `harvest config show`
- `harvest submit auth login`
- `harvest submit auth status`
- `harvest submit auth logout`
- `harvest submit week`
- `harvest whoami`
- `harvest projects`
- `harvest recent`
- `harvest log`
- `harvest today`

## Notes

- Public API commands use Harvest account ID + personal access token.
- `harvest submit ...` uses Harvest website email/password because Harvest does not expose submit-for-approval in the public API.
- Submit email/password are sent only to Harvest website endpoints.
- Saved submit secrets live in macOS Keychain, not the CLI config file.
