package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"harvest/internal/config"
	"harvest/internal/secretstore"
	"harvest/internal/websubmit"
)

type submitAuthStatus struct {
	Email                string     `json:"email,omitempty"`
	BaseURL              string     `json:"base_url,omitempty"`
	SessionSaved         bool       `json:"session_saved"`
	SessionExpiresAt     *time.Time `json:"session_expires_at,omitempty"`
	PasswordSaved        bool       `json:"password_saved"`
	AccessTokenExpiresAt *time.Time `json:"access_token_expires_at,omitempty"`
}

func (a *App) saveSubmitEmail(email string) error {
	if _, err := a.Store.Save(config.Update{SubmitEmail: &email}); err != nil {
		return err
	}
	return nil
}

func (a *App) saveSubmitSession(email string, client SubmitClient) error {
	session, err := client.ExportSession()
	if err != nil {
		return err
	}
	payload, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("encode submit session: %w", err)
	}
	return a.submitSecretStore().Save(a.context(), secretstore.ServiceSubmitSession, email, string(payload))
}

func (a *App) restoreSubmitSession(email string, client SubmitClient) error {
	payload, err := a.submitSecretStore().Load(a.context(), secretstore.ServiceSubmitSession, email)
	if err != nil {
		if errors.Is(err, secretstore.ErrNotFound) {
			return nil
		}
		return err
	}

	var session websubmit.Session
	if err := json.Unmarshal([]byte(payload), &session); err != nil {
		return fmt.Errorf("decode submit session: %w", err)
	}
	return client.RestoreSession(session)
}

func (a *App) submitAuthStatus(email string) (submitAuthStatus, error) {
	status := submitAuthStatus{Email: strings.TrimSpace(email)}
	if status.Email == "" {
		return status, nil
	}

	sessionPayload, err := a.submitSecretStore().Load(a.context(), secretstore.ServiceSubmitSession, status.Email)
	if err != nil && !errors.Is(err, secretstore.ErrNotFound) {
		return submitAuthStatus{}, err
	}
	if err == nil {
		var session websubmit.Session
		if decodeErr := json.Unmarshal([]byte(sessionPayload), &session); decodeErr != nil {
			return submitAuthStatus{}, fmt.Errorf("decode submit session: %w", decodeErr)
		}
		status.SessionSaved = true
		status.BaseURL = session.BaseURL
		for _, cookie := range session.HarvestCookies {
			switch cookie.Name {
			case "_harvest_sess":
				status.SessionExpiresAt = timePointer(cookie.Expires)
			case "production_access_token":
				status.AccessTokenExpiresAt = timePointer(cookie.Expires)
			}
		}
	}

	passwordSaved, err := a.submitSecretStore().Exists(a.context(), secretstore.ServiceSubmitPassword, status.Email)
	if err != nil {
		return submitAuthStatus{}, err
	}
	status.PasswordSaved = passwordSaved
	return status, nil
}

func presentWithExpiry(present bool, expiresAt *time.Time) string {
	if !present {
		return "not saved"
	}
	if expiresAt == nil {
		return "saved"
	}
	return fmt.Sprintf("saved (expires %s)", expiresAt.Format(time.RFC3339))
}

func timePointer(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	copy := value
	return &copy
}
