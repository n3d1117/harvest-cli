# PAT-Based Submit Replacement

## Summary

Replace the old website-scraping submit flow with Harvest's private PAT-authenticated submit endpoint. Keep `harvest submit week` with the same flags and remove all `submit auth` commands, website session handling, and keychain-backed submit state.

## Implementation

- Add `weekly_summary` and `/daily/submit` support to `cli/internal/harvestapi`.
- Resolve submit state from `weekly_summary` and submit with PAT auth plus `x-harvest-webview-client: true`.
- Keep `submit week --dry-run` as a preview of the resolved week and current submit state.
- Re-read `weekly_summary` after submit and fail if Harvest still reports the week as not submitted.
- Remove `SubmitEmail`, submit-session helpers, keychain submit storage, and the entire `websubmit` package.

## Tests

- Cover `weekly_summary` query building and parsing.
- Cover `/daily/submit` multipart fields, required headers, `204`, and redirect-style success.
- Cover command-level dry-run, submit, resubmit, approved-week failure, and JSON output shape.

## Docs

- Update `README.md`.
- Update `skills/harvest/SKILL.md`.
- Update all files in `skills/harvest/references/` that mention submit auth or website sessions.
