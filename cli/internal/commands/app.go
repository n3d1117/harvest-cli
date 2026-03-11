package commands

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"harvest/internal/config"
	"harvest/internal/harvestapi"
	"harvest/internal/output"
	"harvest/internal/prompt"
)

const rootHelp = `harvest logs time to Harvest.

Usage:
  harvest <command> [flags]
  harvest help [command]

Commands:
  login         Save Harvest credentials interactively
  config        Set or inspect stored config
  whoami        Verify Harvest auth
  projects      List active project/task pairs
  log           Create a time entry
  today         Show today's entries
  help          Show help for a command

Examples:
  harvest login
  harvest whoami
  harvest projects --json
  harvest log --project "Acme" --task "Development" --duration 1h30m --notes "CLI scaffolding"
  harvest today
`

const loginHelp = `Usage:
  harvest login

Prompt for a Harvest account ID and personal access token, validate them,
and save them to the local config file.

Examples:
  harvest login
`

const configHelp = `Usage:
  harvest config <command> [flags]

Commands:
  set           Save account, token, or defaults
  show          Show effective config without exposing the token

Examples:
  harvest config set --account-id 123456 --token abc123
  harvest config set --default-project "Acme" --default-task "Development"
  harvest config show
`

const configSetHelp = `Usage:
  harvest config set [flags]

Flags:
  --account-id string
  --token string
  --default-project string
  --default-task string

Examples:
  harvest config set --account-id 123456 --token abc123
  harvest config set --default-project "Acme" --default-task "Development"
`

const configShowHelp = `Usage:
  harvest config show [--json]

Show the effective config after applying environment overrides.
`

const whoamiHelp = `Usage:
  harvest whoami [--json]

Verify auth and print the current Harvest user.
`

const projectsHelp = `Usage:
  harvest projects [--json]

List active project/task pairs available for time logging.
`

const logHelp = `Usage:
  harvest log --project <name> --task <name> --duration <duration> [flags]

Flags:
  --project string
  --task string
  --duration string
  --date string
  --notes, -n string
  --json

Notes:
  - Duration uses Go duration strings like 45m, 1h30m, or 2h.
  - Date defaults to local today in YYYY-MM-DD.
  - Project and task can come from config defaults.

Examples:
  harvest log --project "Acme" --task "Development" --duration 1h30m
  harvest log --duration 45m --notes "Bug fix"
`

const todayHelp = `Usage:
  harvest today [--json]

Show today's time entries and the total logged hours.
`

type HarvestService interface {
	Me(context.Context) (harvestapi.User, error)
	ProjectAssignments(context.Context) ([]harvestapi.ProjectAssignment, error)
	CreateTimeEntry(context.Context, harvestapi.CreateTimeEntryInput) (harvestapi.TimeEntry, error)
	TimeEntries(context.Context, string, string) ([]harvestapi.TimeEntry, error)
}

type ClientFactory func(config.Values) (HarvestService, error)

type App struct {
	Store         *config.Store
	ClientFactory ClientFactory
	Prompt        prompt.Prompter
	Stdout        io.Writer
	Stderr        io.Writer
	Now           func() time.Time
	Context       context.Context
}

type exitError struct {
	Code    int
	Message string
}

type stringFlag struct {
	value string
	set   bool
}

func (f *stringFlag) String() string {
	return f.value
}

func (f *stringFlag) Set(value string) error {
	f.value = value
	f.set = true
	return nil
}

func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	path, err := config.DefaultPath()
	if err != nil {
		output.Errorf(stderr, "%v", err)
		return 1
	}

	app := &App{
		Store:         config.NewStore(path, os.Getenv),
		ClientFactory: defaultClientFactory,
		Prompt:        prompt.NewTerminal(stdin, stdout),
		Stdout:        stdout,
		Stderr:        stderr,
		Now:           time.Now,
		Context:       context.Background(),
	}

	if err := app.Execute(args); err != nil {
		var coded *exitError
		if errors.As(err, &coded) {
			if coded.Message != "" {
				output.Errorf(stderr, "%s", coded.Message)
			}
			return coded.Code
		}

		output.Errorf(stderr, "%v", err)
		return 1
	}

	return 0
}

func defaultClientFactory(values config.Values) (HarvestService, error) {
	if strings.TrimSpace(values.AccountID) == "" || strings.TrimSpace(values.Token) == "" {
		return nil, &exitError{
			Code:    3,
			Message: "missing Harvest credentials; run `harvest login` or `harvest config set --account-id ... --token ...`",
		}
	}

	return harvestapi.New(values.AccountID, values.Token, nil), nil
}

func (e *exitError) Error() string {
	return e.Message
}

func (a *App) Execute(args []string) error {
	if len(args) == 0 {
		fmt.Fprint(a.Stdout, rootHelp)
		return &exitError{Code: 2, Message: "missing command"}
	}

	switch args[0] {
	case "help":
		return a.runHelp(args[1:])
	case "-h", "--help":
		fmt.Fprint(a.Stdout, rootHelp)
		return nil
	case "login":
		return a.runLogin(args[1:])
	case "config":
		return a.runConfig(args[1:])
	case "whoami":
		return a.runWhoami(args[1:])
	case "projects":
		return a.runProjects(args[1:])
	case "log":
		return a.runLog(args[1:])
	case "today":
		return a.runToday(args[1:])
	default:
		fmt.Fprint(a.Stdout, rootHelp)
		return &exitError{Code: 2, Message: fmt.Sprintf("unknown command %q", args[0])}
	}
}

func (a *App) runHelp(args []string) error {
	if len(args) == 0 {
		fmt.Fprint(a.Stdout, rootHelp)
		return nil
	}

	switch args[0] {
	case "login":
		fmt.Fprint(a.Stdout, loginHelp)
	case "config":
		fmt.Fprint(a.Stdout, configHelp)
	case "whoami":
		fmt.Fprint(a.Stdout, whoamiHelp)
	case "projects":
		fmt.Fprint(a.Stdout, projectsHelp)
	case "log":
		fmt.Fprint(a.Stdout, logHelp)
	case "today":
		fmt.Fprint(a.Stdout, todayHelp)
	default:
		return &exitError{Code: 2, Message: fmt.Sprintf("unknown help topic %q", args[0])}
	}
	return nil
}

func (a *App) runLogin(args []string) error {
	fs := newFlagSet("login", loginHelp, a.Stdout, a.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return &exitError{Code: 2, Message: "login does not accept positional arguments"}
	}

	accountID, err := a.Prompt.Prompt("Harvest account ID: ")
	if err != nil {
		return fmt.Errorf("read account id: %w", err)
	}
	token, err := a.Prompt.PromptSecret("Harvest personal access token: ")
	if err != nil {
		return fmt.Errorf("read token: %w", err)
	}

	accountID = strings.TrimSpace(accountID)
	token = strings.TrimSpace(token)
	if accountID == "" || token == "" {
		return &exitError{Code: 2, Message: "account ID and token are required"}
	}

	client, err := a.ClientFactory(config.Values{
		AccountID: accountID,
		Token:     token,
	})
	if err != nil {
		return err
	}

	user, err := client.Me(a.context())
	if err != nil {
		return &exitError{Code: 3, Message: fmt.Sprintf("login failed: %v", err)}
	}

	if _, err := a.Store.Save(config.Update{
		AccountID: &accountID,
		Token:     &token,
	}); err != nil {
		return err
	}

	name := fullName(user)
	if name == "" {
		name = user.Email
	}
	fmt.Fprintf(a.Stdout, "Saved Harvest credentials for %s.\n", name)
	return nil
}

func (a *App) runConfig(args []string) error {
	if len(args) == 0 {
		fmt.Fprint(a.Stdout, configHelp)
		return &exitError{Code: 2, Message: "missing config subcommand"}
	}

	switch args[0] {
	case "help":
		fmt.Fprint(a.Stdout, configHelp)
		return nil
	case "set":
		return a.runConfigSet(args[1:])
	case "show":
		return a.runConfigShow(args[1:])
	default:
		fmt.Fprint(a.Stdout, configHelp)
		return &exitError{Code: 2, Message: fmt.Sprintf("unknown config subcommand %q", args[0])}
	}
}

func (a *App) runConfigSet(args []string) error {
	fs := newFlagSet("config set", configSetHelp, a.Stdout, a.Stderr)
	var accountID stringFlag
	var token stringFlag
	var defaultProject stringFlag
	var defaultTask stringFlag

	fs.Var(&accountID, "account-id", "Harvest account ID")
	fs.Var(&token, "token", "Harvest personal access token")
	fs.Var(&defaultProject, "default-project", "Default project name")
	fs.Var(&defaultTask, "default-task", "Default task name")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return &exitError{Code: 2, Message: "config set does not accept positional arguments"}
	}

	update := config.Update{}
	changed := false

	if accountID.set {
		update.AccountID = &accountID.value
		changed = true
	}
	if token.set {
		update.Token = &token.value
		changed = true
	}
	if defaultProject.set {
		update.DefaultProject = &defaultProject.value
		changed = true
	}
	if defaultTask.set {
		update.DefaultTask = &defaultTask.value
		changed = true
	}

	if !changed {
		return &exitError{Code: 2, Message: "nothing to save; pass at least one flag"}
	}

	values, err := a.Store.Save(update)
	if err != nil {
		return err
	}

	fmt.Fprintf(a.Stdout, "Saved config to %s.\n", a.Store.Path)
	if values.DefaultProject != "" || values.DefaultTask != "" {
		fmt.Fprintf(a.Stdout, "Defaults: project=%q task=%q\n", values.DefaultProject, values.DefaultTask)
	}
	return nil
}

func (a *App) runConfigShow(args []string) error {
	fs := newFlagSet("config show", configShowHelp, a.Stdout, a.Stderr)
	jsonOutput := fs.Bool("json", false, "Print JSON")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return &exitError{Code: 2, Message: "config show does not accept positional arguments"}
	}

	values, err := a.Store.Effective(config.Values{})
	if err != nil {
		return err
	}

	redacted := config.Redacted(values)
	if *jsonOutput {
		return output.JSON(a.Stdout, struct {
			OK         bool                  `json:"ok"`
			ConfigPath string                `json:"config_path"`
			Config     config.RedactedValues `json:"config"`
		}{
			OK:         true,
			ConfigPath: a.Store.Path,
			Config:     redacted,
		})
	}

	fmt.Fprintf(a.Stdout, "Config file: %s\n", a.Store.Path)
	fmt.Fprintf(a.Stdout, "Account ID: %s\n", printableValue(redacted.AccountID))
	fmt.Fprintf(a.Stdout, "Token: %s\n", map[bool]string{true: "present", false: "missing"}[redacted.TokenPresent])
	fmt.Fprintf(a.Stdout, "Default project: %s\n", printableValue(redacted.DefaultProject))
	fmt.Fprintf(a.Stdout, "Default task: %s\n", printableValue(redacted.DefaultTask))
	return nil
}

func (a *App) runWhoami(args []string) error {
	fs := newFlagSet("whoami", whoamiHelp, a.Stdout, a.Stderr)
	jsonOutput := fs.Bool("json", false, "Print JSON")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return &exitError{Code: 2, Message: "whoami does not accept positional arguments"}
	}

	client, _, err := a.client(config.Values{})
	if err != nil {
		return err
	}

	user, err := client.Me(a.context())
	if err != nil {
		return err
	}

	if *jsonOutput {
		return output.JSON(a.Stdout, struct {
			OK   bool            `json:"ok"`
			User harvestapi.User `json:"user"`
		}{
			OK:   true,
			User: user,
		})
	}

	name := fullName(user)
	if name == "" {
		name = user.Email
	}
	fmt.Fprintf(a.Stdout, "%s\n", name)
	fmt.Fprintf(a.Stdout, "Email: %s\n", user.Email)
	fmt.Fprintf(a.Stdout, "User ID: %d\n", user.ID)
	return nil
}

func (a *App) runProjects(args []string) error {
	fs := newFlagSet("projects", projectsHelp, a.Stdout, a.Stderr)
	jsonOutput := fs.Bool("json", false, "Print JSON")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return &exitError{Code: 2, Message: "projects does not accept positional arguments"}
	}

	client, _, err := a.client(config.Values{})
	if err != nil {
		return err
	}

	assignments, err := client.ProjectAssignments(a.context())
	if err != nil {
		return err
	}

	pairs := FlattenProjectTaskPairs(assignments)
	if *jsonOutput {
		return output.JSON(a.Stdout, struct {
			OK       bool              `json:"ok"`
			Projects []ProjectTaskPair `json:"projects"`
		}{
			OK:       true,
			Projects: pairs,
		})
	}

	if len(pairs) == 0 {
		fmt.Fprintln(a.Stdout, "No active project/task pairs found.")
		return nil
	}

	writer := tabwriter.NewWriter(a.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "PROJECT\tTASK\tPROJECT ID\tTASK ID")
	for _, pair := range pairs {
		fmt.Fprintf(writer, "%s\t%s\t%d\t%d\n", pair.ProjectName, pair.TaskName, pair.ProjectID, pair.TaskID)
	}
	return writer.Flush()
}

func (a *App) runLog(args []string) error {
	fs := newFlagSet("log", logHelp, a.Stdout, a.Stderr)
	var projectFlag stringFlag
	var taskFlag stringFlag

	duration := fs.String("duration", "", "Go duration string like 45m or 1h30m")
	date := fs.String("date", "", "Date in YYYY-MM-DD")
	notes := fs.String("notes", "", "Optional entry notes")
	fs.Var(&projectFlag, "project", "Project name")
	fs.Var(&taskFlag, "task", "Task name")
	fs.Var(&projectFlag, "p", "Project name")
	fs.Var(&taskFlag, "t", "Task name")
	fs.StringVar(notes, "n", "", "Optional entry notes")
	jsonOutput := fs.Bool("json", false, "Print JSON")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return &exitError{Code: 2, Message: "log does not accept positional arguments"}
	}
	if strings.TrimSpace(*duration) == "" {
		return &exitError{Code: 2, Message: "--duration is required"}
	}

	hours, err := ParseDurationHours(*duration)
	if err != nil {
		return &exitError{Code: 2, Message: err.Error()}
	}

	entryDate := strings.TrimSpace(*date)
	if entryDate == "" {
		entryDate = formatDate(a.Now().In(time.Local))
	} else if _, err := time.ParseInLocation("2006-01-02", entryDate, time.Local); err != nil {
		return &exitError{Code: 2, Message: "date must use YYYY-MM-DD"}
	}

	overrides := config.Values{
		DefaultProject: projectFlag.value,
		DefaultTask:    taskFlag.value,
	}

	client, values, err := a.client(overrides)
	if err != nil {
		return err
	}

	projectName := strings.TrimSpace(values.DefaultProject)
	taskName := strings.TrimSpace(values.DefaultTask)
	if projectName == "" {
		return &exitError{Code: 2, Message: "project is required; pass --project or set a default project"}
	}
	if taskName == "" {
		return &exitError{Code: 2, Message: "task is required; pass --task or set a default task"}
	}

	assignments, err := client.ProjectAssignments(a.context())
	if err != nil {
		return err
	}

	pair, err := ResolveProjectTask(assignments, projectName, taskName)
	if err != nil {
		return &exitError{Code: 2, Message: err.Error()}
	}

	entry, err := client.CreateTimeEntry(a.context(), harvestapi.CreateTimeEntryInput{
		ProjectID: pair.ProjectID,
		TaskID:    pair.TaskID,
		SpentDate: entryDate,
		Hours:     hours,
		Notes:     strings.TrimSpace(*notes),
	})
	if err != nil {
		return err
	}

	result := struct {
		ID        int64   `json:"id"`
		Date      string  `json:"date"`
		Hours     float64 `json:"hours"`
		Notes     string  `json:"notes,omitempty"`
		ProjectID int64   `json:"project_id"`
		Project   string  `json:"project"`
		TaskID    int64   `json:"task_id"`
		Task      string  `json:"task"`
	}{
		ID:        entry.ID,
		Date:      entryDate,
		Hours:     hours,
		Notes:     strings.TrimSpace(*notes),
		ProjectID: pair.ProjectID,
		Project:   pair.ProjectName,
		TaskID:    pair.TaskID,
		Task:      pair.TaskName,
	}

	if *jsonOutput {
		return output.JSON(a.Stdout, struct {
			OK    bool `json:"ok"`
			Entry any  `json:"entry"`
		}{
			OK:    true,
			Entry: result,
		})
	}

	fmt.Fprintf(a.Stdout, "Logged %.2fh on %s to %s / %s (#%d).\n", result.Hours, result.Date, result.Project, result.Task, result.ID)
	if result.Notes != "" {
		fmt.Fprintf(a.Stdout, "Notes: %s\n", result.Notes)
	}
	return nil
}

func (a *App) runToday(args []string) error {
	fs := newFlagSet("today", todayHelp, a.Stdout, a.Stderr)
	jsonOutput := fs.Bool("json", false, "Print JSON")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return &exitError{Code: 2, Message: "today does not accept positional arguments"}
	}

	client, _, err := a.client(config.Values{})
	if err != nil {
		return err
	}

	date := formatDate(a.Now().In(time.Local))
	entries, err := client.TimeEntries(a.context(), date, date)
	if err != nil {
		return err
	}

	total := 0.0
	for _, entry := range entries {
		total += entry.Hours
	}

	if *jsonOutput {
		return output.JSON(a.Stdout, struct {
			OK         bool                   `json:"ok"`
			Date       string                 `json:"date"`
			TotalHours float64                `json:"total_hours"`
			Entries    []harvestapi.TimeEntry `json:"entries"`
		}{
			OK:         true,
			Date:       date,
			TotalHours: total,
			Entries:    entries,
		})
	}

	if len(entries) == 0 {
		fmt.Fprintf(a.Stdout, "No entries for %s.\n", date)
		return nil
	}

	writer := tabwriter.NewWriter(a.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "DATE\tPROJECT\tTASK\tHOURS\tNOTES")
	for _, entry := range entries {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%.2f\t%s\n", entry.SpentDate, entry.Project.Name, entry.Task.Name, entry.Hours, entry.Notes)
	}
	fmt.Fprintf(writer, "TOTAL\t\t\t%.2f\t\n", total)
	return writer.Flush()
}

func (a *App) client(overrides config.Values) (HarvestService, config.Values, error) {
	values, err := a.Store.Effective(overrides)
	if err != nil {
		return nil, config.Values{}, err
	}

	client, err := a.ClientFactory(values)
	if err != nil {
		return nil, config.Values{}, err
	}

	return client, values, nil
}

func (a *App) context() context.Context {
	if a.Context != nil {
		return a.Context
	}
	return context.Background()
}

func newFlagSet(name, help string, stdout, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprint(stdout, help)
	}
	return fs
}

func printableValue(value string) string {
	if value == "" {
		return "(not set)"
	}
	return value
}
