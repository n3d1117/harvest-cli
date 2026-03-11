package commands

import (
	"fmt"
	"strings"
	"time"

	"harvest/internal/config"
	"harvest/internal/output"
	"harvest/internal/secretstore"
)

func (a *App) runSubmitAuthLogin(args []string) error {
	fs := newFlagSet("submit auth login", submitAuthLoginHelp, a.Stdout, a.Stderr)
	email := fs.String("email", "", "Harvest website email")
	savePassword := fs.Bool("save-password", false, "Save the Harvest website password in macOS Keychain")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return &exitError{Code: 2, Message: "submit auth login does not accept positional arguments"}
	}

	values, err := a.Store.Effective(config.Values{})
	if err != nil {
		return err
	}
	if strings.TrimSpace(values.AccountID) == "" {
		return &exitError{Code: 3, Message: "submit auth needs a Harvest account ID; run `harvest login` or `harvest config set --account-id ...` first"}
	}

	submitEmail := strings.TrimSpace(*email)
	if submitEmail == "" {
		submitEmail = strings.TrimSpace(values.SubmitEmail)
	}
	if submitEmail == "" {
		submitEmail, err = a.Prompt.Prompt("Harvest website email: ")
		if err != nil {
			return fmt.Errorf("read submit email: %w", err)
		}
		submitEmail = strings.TrimSpace(submitEmail)
	}
	if submitEmail == "" {
		return &exitError{Code: 2, Message: "submit email is required"}
	}

	password, err := a.Prompt.PromptSecret("Harvest website password: ")
	if err != nil {
		return fmt.Errorf("read submit password: %w", err)
	}
	password = strings.TrimSpace(password)
	if password == "" {
		return &exitError{Code: 2, Message: "submit password is required"}
	}

	client, err := a.submitClient(values.AccountID)
	if err != nil {
		return err
	}

	state, err := client.Login(a.context(), submitEmail, password)
	if err != nil {
		return &exitError{Code: 3, Message: fmt.Sprintf("submit auth login failed: %v", err)}
	}

	if err := a.saveSubmitEmail(submitEmail); err != nil {
		return err
	}
	if err := a.saveSubmitSession(submitEmail, client); err != nil {
		return err
	}
	if *savePassword {
		if err := a.submitSecretStore().Save(a.context(), secretstore.ServiceSubmitPassword, submitEmail, password); err != nil {
			return err
		}
	}

	name := state.Name
	if name == "" {
		name = submitEmail
	}
	fmt.Fprintf(a.Stdout, "Saved Harvest submit auth for %s.\n", name)
	if !state.SessionExpiresAt.IsZero() {
		fmt.Fprintf(a.Stdout, "Submit session expires: %s\n", state.SessionExpiresAt.Format(time.RFC3339))
	}
	if *savePassword {
		fmt.Fprintln(a.Stdout, "Password saved in macOS Keychain.")
	}
	return nil
}

func (a *App) runSubmitAuthStatus(args []string) error {
	fs := newFlagSet("submit auth status", submitAuthStatusHelp, a.Stdout, a.Stderr)
	jsonOutput := fs.Bool("json", false, "Print JSON")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return &exitError{Code: 2, Message: "submit auth status does not accept positional arguments"}
	}

	values, err := a.Store.Effective(config.Values{})
	if err != nil {
		return err
	}
	status, err := a.submitAuthStatus(values.SubmitEmail)
	if err != nil {
		return err
	}

	if *jsonOutput {
		return output.JSON(a.Stdout, struct {
			OK     bool             `json:"ok"`
			Status submitAuthStatus `json:"status"`
		}{
			OK:     true,
			Status: status,
		})
	}

	fmt.Fprintf(a.Stdout, "Submit email: %s\n", printableValue(status.Email))
	fmt.Fprintf(a.Stdout, "Harvest base URL: %s\n", printableValue(status.BaseURL))
	fmt.Fprintf(a.Stdout, "Session: %s\n", presentWithExpiry(status.SessionSaved, status.SessionExpiresAt))
	fmt.Fprintf(a.Stdout, "Password: %s\n", map[bool]string{true: "saved", false: "not saved"}[status.PasswordSaved])
	if status.AccessTokenExpiresAt != nil {
		fmt.Fprintf(a.Stdout, "Access token expires: %s\n", status.AccessTokenExpiresAt.Format(time.RFC3339))
	}
	return nil
}

func (a *App) runSubmitAuthLogout(args []string) error {
	fs := newFlagSet("submit auth logout", submitAuthLogoutHelp, a.Stdout, a.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return &exitError{Code: 2, Message: "submit auth logout does not accept positional arguments"}
	}

	values, err := a.Store.Effective(config.Values{})
	if err != nil {
		return err
	}
	email := strings.TrimSpace(values.SubmitEmail)
	if email != "" {
		if err := a.submitSecretStore().Delete(a.context(), secretstore.ServiceSubmitPassword, email); err != nil {
			return err
		}
		if err := a.submitSecretStore().Delete(a.context(), secretstore.ServiceSubmitSession, email); err != nil {
			return err
		}
	}

	empty := ""
	if _, err := a.Store.Save(config.Update{SubmitEmail: &empty}); err != nil {
		return err
	}

	fmt.Fprintln(a.Stdout, "Removed saved Harvest submit auth.")
	return nil
}
