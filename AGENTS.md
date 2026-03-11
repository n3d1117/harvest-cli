# AGENTS.md

This repo contains a small `harvest` CLI for logging time to Harvest. It also ships an agent skill in `skills/harvest` that tells agents how to use the CLI safely and correctly.

When you change the CLI, update the skill in the same change.

Keep these files in sync with the real command surface:

- `skills/harvest/SKILL.md`
- `skills/harvest/references/commands.md`
- `skills/harvest/agents/openai.yaml` when the skill name or scope changes
- `README.md` when install or usage examples change

Update the skill when any of these change:

- command names
- flags
- help text
- config keys or environment variables
- output format, especially JSON shape
- auth flow

Do not merge a CLI change that makes the skill docs stale.

Use Conventional Commits for commit messages.
