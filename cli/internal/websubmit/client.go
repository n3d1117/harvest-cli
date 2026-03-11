package websubmit

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"
)

const (
	DefaultLoginBaseURL = "https://id.getharvest.com"
	defaultUserAgent    = "harvest-cli"
)

var ErrUnauthenticated = errors.New("submit auth is missing or expired")

type Cookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Domain   string    `json:"domain"`
	Path     string    `json:"path"`
	Expires  time.Time `json:"expires,omitempty"`
	Secure   bool      `json:"secure,omitempty"`
	HTTPOnly bool      `json:"http_only,omitempty"`
}

type Session struct {
	BaseURL         string   `json:"base_url"`
	HarvestCookies  []Cookie `json:"harvest_cookies,omitempty"`
	IdentityCookies []Cookie `json:"identity_cookies,omitempty"`
}

type AuthState struct {
	BaseURL              string    `json:"base_url"`
	Email                string    `json:"email,omitempty"`
	Name                 string    `json:"name,omitempty"`
	SessionExpiresAt     time.Time `json:"session_expires_at,omitempty"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at,omitempty"`
}

type SubmitResult struct {
	Action          string `json:"action"`
	WeekStart       string `json:"week_start"`
	WeekEnd         string `json:"week_end"`
	ReturnTo        string `json:"return_to"`
	SubmittedBefore bool   `json:"submitted_before"`
	SubmittedAfter  bool   `json:"submitted_after"`
}

type Client struct {
	AccountID      string
	BaseURL        string
	LoginBaseURL   string
	HTTPClient     *http.Client
	UserAgent      string
	trackedCookies map[string]Cookie
}

func New(accountID string, httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, fmt.Errorf("create cookie jar: %w", err)
		}
		httpClient = &http.Client{Jar: jar}
	}
	if httpClient.Jar == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, fmt.Errorf("create cookie jar: %w", err)
		}
		httpClient.Jar = jar
	}

	return &Client{
		AccountID:      accountID,
		LoginBaseURL:   DefaultLoginBaseURL,
		HTTPClient:     httpClient,
		UserAgent:      defaultUserAgent,
		trackedCookies: map[string]Cookie{},
	}, nil
}
