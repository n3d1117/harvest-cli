package websubmit

import (
	"fmt"
	"net/url"
	"strings"
)

func (c *Client) RestoreSession(session Session) error {
	if strings.TrimSpace(session.BaseURL) != "" {
		c.BaseURL = strings.TrimRight(session.BaseURL, "/")
	}

	harvestURL, err := url.Parse(c.baseURLOrFallback("https://harvestapp.com"))
	if err != nil {
		return fmt.Errorf("parse Harvest base url: %w", err)
	}
	c.HTTPClient.Jar.SetCookies(harvestURL, cookiesToHTTP(session.HarvestCookies))

	loginURL, err := url.Parse(c.LoginBaseURL)
	if err != nil {
		return fmt.Errorf("parse login base url: %w", err)
	}
	c.HTTPClient.Jar.SetCookies(loginURL, cookiesToHTTP(session.IdentityCookies))
	c.trackedCookies = map[string]Cookie{}
	for _, cookie := range append(session.HarvestCookies, session.IdentityCookies...) {
		c.trackedCookies[c.cookieKey(cookie)] = cookie
	}
	return nil
}

func (c *Client) ExportSession() (Session, error) {
	loginURL, err := url.Parse(c.LoginBaseURL)
	if err != nil {
		return Session{}, fmt.Errorf("parse login base url: %w", err)
	}

	harvestURL, err := url.Parse(c.baseURLOrFallback("https://harvestapp.com"))
	if err != nil {
		return Session{}, fmt.Errorf("parse harvest base url: %w", err)
	}

	return Session{
		BaseURL:         c.BaseURL,
		HarvestCookies:  c.exportTrackedCookies(harvestURL.Host),
		IdentityCookies: c.exportTrackedCookies(loginURL.Host),
	}, nil
}

func (c *Client) SessionState() AuthState {
	state := AuthState{BaseURL: c.BaseURL}
	for _, cookie := range c.trackedCookies {
		switch cookie.Name {
		case "_harvest_sess":
			state.SessionExpiresAt = cookie.Expires
		case "production_access_token":
			state.AccessTokenExpiresAt = cookie.Expires
		}
	}
	return state
}

func (c *Client) baseURLOrFallback(fallback string) string {
	if strings.TrimSpace(c.BaseURL) != "" {
		return c.BaseURL
	}
	return fallback
}
