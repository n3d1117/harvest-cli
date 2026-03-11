package harvestapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestMeSendsAuthHeaders(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("unexpected auth header: %q", got)
		}
		if got := r.Header.Get("Harvest-Account-Id"); got != "123" {
			t.Fatalf("unexpected account header: %q", got)
		}
		if got := r.Header.Get("User-Agent"); got != "harvest-cli" {
			t.Fatalf("unexpected user agent: %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(User{
			ID:        1,
			FirstName: "Ned",
			LastName:  "Tester",
			Email:     "ned@example.com",
		})
	}))
	defer server.Close()

	client := New("123", "test-token", server.Client())
	client.BaseURL = server.URL

	user, err := client.Me(context.Background())
	if err != nil {
		t.Fatalf("me request failed: %v", err)
	}
	if user.Email != "ned@example.com" {
		t.Fatalf("unexpected user email: %q", user.Email)
	}
}

func TestCreateTimeEntrySendsExpectedBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/time_entries" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if payload["project_id"].(float64) != 7 {
			t.Fatalf("unexpected project id: %v", payload["project_id"])
		}
		if payload["task_id"].(float64) != 9 {
			t.Fatalf("unexpected task id: %v", payload["task_id"])
		}
		if payload["spent_date"].(string) != "2026-03-11" {
			t.Fatalf("unexpected date: %v", payload["spent_date"])
		}
		if payload["hours"].(float64) != 1.5 {
			t.Fatalf("unexpected hours: %v", payload["hours"])
		}
		if _, ok := payload["notes"]; ok {
			t.Fatalf("expected notes to be omitted when empty")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TimeEntry{ID: 88, Hours: 1.5})
	}))
	defer server.Close()

	client := New("123", "token", server.Client())
	client.BaseURL = server.URL

	entry, err := client.CreateTimeEntry(context.Background(), CreateTimeEntryInput{
		ProjectID: 7,
		TaskID:    9,
		SpentDate: "2026-03-11",
		Hours:     1.5,
	})
	if err != nil {
		t.Fatalf("create time entry: %v", err)
	}
	if entry.ID != 88 {
		t.Fatalf("unexpected entry id: %d", entry.ID)
	}
}

func TestTimeEntriesIncludesDateRange(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			t.Fatalf("parse query: %v", err)
		}
		if values.Get("from") != "2026-03-11" {
			t.Fatalf("unexpected from date: %q", values.Get("from"))
		}
		if values.Get("to") != "2026-03-11" {
			t.Fatalf("unexpected to date: %q", values.Get("to"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(timeEntriesPage{
			TimeEntries: []TimeEntry{
				{ID: 1, Hours: 0.5, Notes: "test"},
			},
		})
	}))
	defer server.Close()

	client := New("123", "token", server.Client())
	client.BaseURL = server.URL

	entries, err := client.TimeEntries(context.Background(), "2026-03-11", "2026-03-11")
	if err != nil {
		t.Fatalf("time entries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("unexpected entry count: %d", len(entries))
	}
}

func TestDecodeAPIErrorPrefersStructuredMessage(t *testing.T) {
	t.Parallel()

	response := &http.Response{
		Status:     "401 Unauthorized",
		StatusCode: http.StatusUnauthorized,
		Body:       ioNopCloser(strings.NewReader(`{"error":"bad token"}`)),
	}

	err := decodeAPIError(response)
	if err == nil || err.Error() != "401 Unauthorized: bad token" {
		t.Fatalf("unexpected error: %v", err)
	}
}

type readCloser struct {
	*strings.Reader
}

func (r readCloser) Close() error { return nil }

func ioNopCloser(reader *strings.Reader) readCloser {
	return readCloser{Reader: reader}
}
