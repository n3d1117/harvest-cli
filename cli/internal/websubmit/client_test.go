package websubmit

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestLoginAndSubmitWeek(t *testing.T) {
	t.Parallel()

	var submitForm url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/harvest/sign_in":
			fmt.Fprint(w, `<!DOCTYPE html><html><head><title>Sign in - Harvest</title></head><body><form action="/sessions" method="post"><input type="hidden" name="authenticity_token" value="login-token"></form></body></html>`)
		case r.Method == http.MethodPost && r.URL.Path == "/sessions":
			if err := r.ParseForm(); err != nil {
				t.Fatalf("parse login form: %v", err)
			}
			if got := r.Form.Get("email"); got != "ned@example.com" {
				t.Fatalf("unexpected email: %q", got)
			}
			http.SetCookie(w, &http.Cookie{
				Name:     "production_access_token",
				Value:    "access",
				Path:     "/",
				Expires:  time.Date(2026, 5, 10, 18, 13, 0, 0, time.UTC),
				HttpOnly: true,
			})
			http.SetCookie(w, &http.Cookie{
				Name:     "_harvest_sess",
				Value:    "session",
				Path:     "/",
				Expires:  time.Date(2026, 3, 26, 18, 13, 1, 0, time.UTC),
				HttpOnly: true,
			})
			http.Redirect(w, r, "/time", http.StatusFound)
		case r.Method == http.MethodGet && r.URL.Path == "/time":
			fmt.Fprint(w, pageHTML("/time", "2026-03-09", false, false))
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/time/day/2026/3/11/"):
			fmt.Fprint(w, pageHTML(r.URL.Path, "2026-03-09", false, false))
		case r.Method == http.MethodPost && r.URL.Path == "/daily/review":
			if err := r.ParseForm(); err != nil {
				t.Fatalf("parse submit form: %v", err)
			}
			submitForm = r.Form
			fmt.Fprint(w, pageHTML("/time/day/2026/3/11/99", "2026-03-09", true, false))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("create cookie jar: %v", err)
	}
	client, err := New("1450071", &http.Client{Jar: jar})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	client.LoginBaseURL = server.URL

	state, err := client.Login(context.Background(), "ned@example.com", "secret")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if state.BaseURL != server.URL {
		t.Fatalf("unexpected base url: %q", state.BaseURL)
	}

	result, err := client.SubmitWeek(context.Background(), time.Date(2026, 3, 11, 9, 0, 0, 0, time.Local), time.Date(2026, 3, 11, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("submit week failed: %v", err)
	}
	if result.Action != "submitted" {
		t.Fatalf("unexpected action: %q", result.Action)
	}
	if got := submitForm.Get("period_begin"); got != "68" {
		t.Fatalf("unexpected period_begin: %q", got)
	}
	if got := submitForm.Get("submitted_date"); got != "70" {
		t.Fatalf("unexpected submitted_date: %q", got)
	}

	session, err := client.ExportSession()
	if err != nil {
		t.Fatalf("export session: %v", err)
	}
	if session.BaseURL != server.URL {
		t.Fatalf("unexpected session base url: %q", session.BaseURL)
	}

	status := client.SessionState()
	if status.SessionExpiresAt.IsZero() || status.AccessTokenExpiresAt.IsZero() {
		t.Fatalf("expected cookie expiries, got %+v", status)
	}
}

func pageHTML(path, weekOf string, submitted, approved bool) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
  <head>
    <title>Timesheet – Shape Games – Harvest</title>
    <meta name="csrf-token" content="page-token" />
  </head>
  <body>
    <script id="page_data-data-island" type="application/json">{"of_user_id":99,"approval_required":true,"current_user":{"id":99,"email":"ned@example.com","full_name":"Ned Tester"}}</script>
    <script id="timesheet-data-island" type="application/json">{"week_of":"%s","submitted":%t,"approved":%t}</script>
    <div>%s</div>
  </body>
</html>`, weekOf, submitted, approved, path)
}
