package websubmit

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

func (c *Client) Login(ctx context.Context, email, password string) (AuthState, error) {
	loginURL := fmt.Sprintf("%s/harvest/sign_in?account_id=%s", strings.TrimRight(c.LoginBaseURL, "/"), url.QueryEscape(c.AccountID))
	signInPage, err := c.getPage(ctx, loginURL)
	if err != nil {
		return AuthState{}, err
	}

	formToken := firstMatch(formTokenPattern, signInPage.HTML)
	if formToken == "" {
		return AuthState{}, errors.New("Harvest login page did not include an authenticity token")
	}

	values := url.Values{}
	values.Set("authenticity_token", formToken)
	values.Set("email", email)
	values.Set("password", password)
	values.Set("account_id", c.AccountID)
	values.Set("product", "harvest")

	responsePage, err := c.postForm(ctx, strings.TrimRight(c.LoginBaseURL, "/")+"/sessions", values, signInPage.URL.String())
	if err != nil {
		return AuthState{}, err
	}
	if isSignInPage(responsePage) {
		return AuthState{}, errors.New("Harvest website login failed; check the email and password")
	}

	c.BaseURL = responsePage.URL.Scheme + "://" + responsePage.URL.Host
	responsePage, err = c.fetchTimesheetHome(ctx)
	if err != nil {
		return AuthState{}, err
	}
	if isSignInPage(responsePage) {
		return AuthState{}, ErrUnauthenticated
	}

	state := c.SessionState()
	state.Email = responsePage.PageData.CurrentUser.Email
	state.Name = responsePage.PageData.CurrentUser.FullName
	return state, nil
}
