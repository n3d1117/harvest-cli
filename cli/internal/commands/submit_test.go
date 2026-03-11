package commands

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"harvest/internal/config"
	"harvest/internal/secretstore"
	"harvest/internal/websubmit"
)

type fakeSubmitSecretStore struct {
	items map[string]string
}

func (f *fakeSubmitSecretStore) Save(_ context.Context, service, account, secret string) error {
	if f.items == nil {
		f.items = map[string]string{}
	}
	f.items[service+"|"+account] = secret
	return nil
}

func (f *fakeSubmitSecretStore) Load(_ context.Context, service, account string) (string, error) {
	value, ok := f.items[service+"|"+account]
	if !ok {
		return "", secretstore.ErrNotFound
	}
	return value, nil
}

func (f *fakeSubmitSecretStore) Delete(_ context.Context, service, account string) error {
	delete(f.items, service+"|"+account)
	return nil
}

func (f *fakeSubmitSecretStore) Exists(_ context.Context, service, account string) (bool, error) {
	_, ok := f.items[service+"|"+account]
	return ok, nil
}

type fakeSubmitClient struct {
	loginState       websubmit.AuthState
	loginCalls       int
	loginEmail       string
	loginPassword    string
	exportedSession  websubmit.Session
	restoreSession   websubmit.Session
	restoreWasCalled bool
	submitResult     websubmit.SubmitResult
	submitErr        error
	submitCalls      int
	submitDate       time.Time
	submitNow        time.Time
}

func (f *fakeSubmitClient) RestoreSession(session websubmit.Session) error {
	f.restoreSession = session
	f.restoreWasCalled = true
	return nil
}

func (f *fakeSubmitClient) ExportSession() (websubmit.Session, error) {
	return f.exportedSession, nil
}

func (f *fakeSubmitClient) SessionState() websubmit.AuthState {
	return f.loginState
}

func (f *fakeSubmitClient) Login(_ context.Context, email, password string) (websubmit.AuthState, error) {
	f.loginCalls++
	f.loginEmail = email
	f.loginPassword = password
	return f.loginState, nil
}

func (f *fakeSubmitClient) SubmitWeek(_ context.Context, date, now time.Time) (websubmit.SubmitResult, error) {
	f.submitCalls++
	f.submitDate = date
	f.submitNow = now
	err := f.submitErr
	f.submitErr = nil
	if err != nil {
		return websubmit.SubmitResult{}, err
	}
	return f.submitResult, nil
}

func TestSubmitAuthLoginSavesEmailSessionAndOptionalPassword(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(configPath, nil)
	accountID := "1450071"
	if _, err := store.Save(config.Update{AccountID: &accountID}); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	secrets := &fakeSubmitSecretStore{}
	submitClient := &fakeSubmitClient{
		loginState: websubmit.AuthState{
			BaseURL:          "https://shapegames.harvestapp.com",
			Name:             "Ned Tester",
			Email:            "ned@example.com",
			SessionExpiresAt: time.Date(2026, 3, 26, 18, 13, 1, 0, time.UTC),
		},
		exportedSession: websubmit.Session{
			BaseURL: "https://shapegames.harvestapp.com",
			HarvestCookies: []websubmit.Cookie{
				{Name: "_harvest_sess", Value: "session", Expires: time.Date(2026, 3, 26, 18, 13, 1, 0, time.UTC)},
			},
		},
	}

	var stdout bytes.Buffer
	app := &App{
		Store:         store,
		Prompt:        &fakePrompt{values: []string{"secret-password"}},
		Stdout:        &stdout,
		Stderr:        &bytes.Buffer{},
		Now:           time.Now,
		Context:       context.Background(),
		SubmitSecrets: secrets,
		SubmitClientFactory: func(accountID string) (SubmitClient, error) {
			if accountID != "1450071" {
				t.Fatalf("unexpected account id: %q", accountID)
			}
			return submitClient, nil
		},
	}

	if err := app.Execute([]string{"submit", "auth", "login", "--email", "ned@example.com", "--save-password"}); err != nil {
		t.Fatalf("submit auth login failed: %v", err)
	}

	values, err := store.LoadFile()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if values.SubmitEmail != "ned@example.com" {
		t.Fatalf("expected submit email to be saved, got %q", values.SubmitEmail)
	}
	if submitClient.loginCalls != 1 || submitClient.loginPassword != "secret-password" {
		t.Fatalf("unexpected login calls: %+v", submitClient)
	}
	if _, ok := secrets.items[secretstore.ServiceSubmitPassword+"|ned@example.com"]; !ok {
		t.Fatalf("expected password in secret store")
	}
	if _, ok := secrets.items[secretstore.ServiceSubmitSession+"|ned@example.com"]; !ok {
		t.Fatalf("expected session in secret store")
	}
	if !strings.Contains(stdout.String(), "Saved Harvest submit auth") {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
}

func TestSubmitWeekRefreshesExpiredSessionWithSavedPassword(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(configPath, nil)
	accountID := "1450071"
	submitEmail := "ned@example.com"
	if _, err := store.Save(config.Update{
		AccountID:   &accountID,
		SubmitEmail: &submitEmail,
	}); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	secrets := &fakeSubmitSecretStore{
		items: map[string]string{
			secretstore.ServiceSubmitPassword + "|" + submitEmail: "secret-password",
			secretstore.ServiceSubmitSession + "|" + submitEmail:  `{"base_url":"https://shapegames.harvestapp.com","harvest_cookies":[{"name":"_harvest_sess","value":"old"}]}`,
		},
	}

	submitClient := &fakeSubmitClient{
		loginState: websubmit.AuthState{
			BaseURL:          "https://shapegames.harvestapp.com",
			SessionExpiresAt: time.Date(2026, 3, 26, 18, 13, 1, 0, time.UTC),
		},
		submitErr: websubmit.ErrUnauthenticated,
		submitResult: websubmit.SubmitResult{
			Action:    "submitted",
			WeekStart: "2026-03-09",
			WeekEnd:   "2026-03-15",
		},
		exportedSession: websubmit.Session{
			BaseURL: "https://shapegames.harvestapp.com",
			HarvestCookies: []websubmit.Cookie{
				{Name: "_harvest_sess", Value: "new", Expires: time.Date(2026, 3, 26, 18, 13, 1, 0, time.UTC)},
			},
		},
	}

	var stdout bytes.Buffer
	app := &App{
		Store:  store,
		Prompt: &fakePrompt{},
		Stdout: &stdout,
		Stderr: &bytes.Buffer{},
		Now: func() time.Time {
			return time.Date(2026, 3, 11, 12, 0, 0, 0, time.Local)
		},
		Context:       context.Background(),
		SubmitSecrets: secrets,
		SubmitClientFactory: func(accountID string) (SubmitClient, error) {
			return submitClient, nil
		},
	}

	if err := app.Execute([]string{"submit", "week", "--date", "today"}); err != nil {
		t.Fatalf("submit week failed: %v", err)
	}

	if !submitClient.restoreWasCalled {
		t.Fatalf("expected saved session to be restored")
	}
	if submitClient.loginCalls != 1 {
		t.Fatalf("expected login refresh, got %d", submitClient.loginCalls)
	}
	if submitClient.submitCalls != 2 {
		t.Fatalf("expected submit retry, got %d", submitClient.submitCalls)
	}
	if submitClient.submitDate.Format("2006-01-02") != "2026-03-11" {
		t.Fatalf("unexpected submit date: %s", submitClient.submitDate)
	}
	if !strings.Contains(stdout.String(), "Submitted week 2026-03-09 to 2026-03-15") {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
	if !strings.Contains(secrets.items[secretstore.ServiceSubmitSession+"|"+submitEmail], `"value":"new"`) {
		t.Fatalf("expected refreshed session to be saved, got %q", secrets.items[secretstore.ServiceSubmitSession+"|"+submitEmail])
	}
}
