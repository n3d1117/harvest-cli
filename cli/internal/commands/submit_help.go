package commands

const submitHelp = `Usage:
  harvest submit <command> [flags]

Commands:
  auth         Sign in to Harvest website auth for submit
  week         Submit a week for approval

Notes:
  - ` + "`harvest submit`" + ` uses Harvest website auth, not the public Harvest API token.
  - Harvest does not provide a public API for submitting time for approval.
  - Your submit email and password are sent only to Harvest login endpoints.
  - Saved submit secrets are stored in macOS Keychain, not in the CLI config file.

Examples:
  harvest submit auth login --email you@example.com --save-password
  harvest submit auth status
  harvest submit week --date today
`

const submitAuthHelp = `Usage:
  harvest submit auth <command> [flags]

Commands:
  login        Save Harvest website submit auth
  status       Show Harvest website submit auth status
  logout       Delete saved Harvest website submit auth

Examples:
  harvest submit auth login --email you@example.com --save-password
  harvest submit auth status --json
  harvest submit auth logout
`

const submitAuthLoginHelp = `Usage:
  harvest submit auth login [--email <email>] [--save-password]

Create a Harvest website session for submit.

Notes:
  - Harvest does not expose submit-for-approval in the public API.
  - This command signs in to the Harvest website and saves a submit session.
  - Your email and password are sent only to Harvest (` + "`id.getharvest.com`" + ` and ` + "`*.harvestapp.com`" + `).
  - Saved passwords and session cookies go to macOS Keychain, not the CLI config file.
  - Without ` + "`--save-password`" + `, the CLI saves only the current website session.

Examples:
  harvest submit auth login
  harvest submit auth login --email you@example.com
  harvest submit auth login --email you@example.com --save-password
`

const submitAuthStatusHelp = `Usage:
  harvest submit auth status [--json]

Show Harvest website submit auth status.
`

const submitAuthLogoutHelp = `Usage:
  harvest submit auth logout

Delete the saved Harvest website submit email, password, and session.
`

const submitWeekHelp = `Usage:
  harvest submit week [--date today|YYYY-MM-DD] [--json]

Submit the week that contains the given date for approval.

Notes:
  - This uses Harvest website auth and the private web submit flow.
  - Run ` + "`harvest submit auth login`" + ` first.
  - ` + "`--date`" + ` defaults to local today.

Examples:
  harvest submit week
  harvest submit week --date today
  harvest submit week --date 2026-03-09 --json
`
