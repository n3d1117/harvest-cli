package commands

const submitHelp = `Usage:
  harvest submit <command> [flags]

Commands:
  week         Submit a week for approval

Notes:
  - ` + "`harvest submit`" + ` uses your Harvest account ID and personal access token.
  - ` + "`submit week`" + ` calls Harvest's private submit endpoint.
  - Run ` + "`harvest login`" + ` first.

Examples:
  harvest submit week --date today
`

const submitWeekHelp = `Usage:
  harvest submit week [--date today|YYYY-MM-DD] [--dry-run] [--json]

Submit the week that contains the given date for approval.

Notes:
  - This uses your Harvest account ID and personal access token.
  - ` + "`--date`" + ` defaults to local today.
  - ` + "`--dry-run`" + ` resolves the week and current submit state without posting the final submit request.

Examples:
  harvest submit week
  harvest submit week --date today
  harvest submit week --date today --dry-run
  harvest submit week --date 2026-03-09 --json
`
