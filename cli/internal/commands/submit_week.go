package commands

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"harvest/internal/config"
	"harvest/internal/output"
	"harvest/internal/secretstore"
	"harvest/internal/websubmit"
)

func (a *App) runSubmitWeek(args []string) error {
	fs := newFlagSet("submit week", submitWeekHelp, a.Stdout, a.Stderr)
	dateFlag := fs.String("date", "", "Date in YYYY-MM-DD or `today`")
	jsonOutput := fs.Bool("json", false, "Print JSON")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return &exitError{Code: 2, Message: "submit week does not accept positional arguments"}
	}

	targetDate, err := resolveDateInput(*dateFlag, a.Now())
	if err != nil {
		return &exitError{Code: 2, Message: err.Error()}
	}
	date, err := time.ParseInLocation("2006-01-02", targetDate, time.Local)
	if err != nil {
		return err
	}

	values, err := a.Store.Effective(config.Values{})
	if err != nil {
		return err
	}
	if strings.TrimSpace(values.AccountID) == "" {
		return &exitError{Code: 3, Message: "submit needs a Harvest account ID; run `harvest login` or `harvest config set --account-id ...` first"}
	}
	submitEmail := strings.TrimSpace(values.SubmitEmail)
	if submitEmail == "" {
		return &exitError{Code: 3, Message: "submit auth is not configured; run `harvest submit auth login` first"}
	}

	client, err := a.submitClient(values.AccountID)
	if err != nil {
		return err
	}

	if err := a.restoreSubmitSession(submitEmail, client); err != nil {
		return err
	}

	result, err := client.SubmitWeek(a.context(), date, a.Now())
	if errors.Is(err, websubmit.ErrUnauthenticated) {
		password, loadErr := a.submitSecretStore().Load(a.context(), secretstore.ServiceSubmitPassword, submitEmail)
		if loadErr != nil {
			if errors.Is(loadErr, secretstore.ErrNotFound) {
				return &exitError{Code: 3, Message: "submit auth expired; run `harvest submit auth login` again or save a password with `--save-password`"}
			}
			return loadErr
		}
		if _, err := client.Login(a.context(), submitEmail, password); err != nil {
			return &exitError{Code: 3, Message: fmt.Sprintf("submit auth refresh failed: %v", err)}
		}
		result, err = client.SubmitWeek(a.context(), date, a.Now())
	}
	if err != nil {
		return err
	}

	if err := a.saveSubmitSession(submitEmail, client); err != nil {
		return err
	}

	if *jsonOutput {
		return output.JSON(a.Stdout, struct {
			OK     bool                   `json:"ok"`
			Result websubmit.SubmitResult `json:"result"`
		}{
			OK:     true,
			Result: result,
		})
	}

	verb := map[string]string{
		"submitted":   "Submitted",
		"resubmitted": "Resubmitted",
	}[result.Action]
	if verb == "" {
		verb = "Submitted"
	}
	fmt.Fprintf(a.Stdout, "%s week %s to %s for approval.\n", verb, result.WeekStart, result.WeekEnd)
	return nil
}
