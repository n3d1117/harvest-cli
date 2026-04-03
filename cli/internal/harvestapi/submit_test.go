package harvestapi

import (
	"context"
	"encoding/json"
	"mime"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWeeklySummaryUsesSelectedDate(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/day_entries/weekly_summary" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("start_date"); got != "2026-04-01" {
			t.Fatalf("unexpected start date: %q", got)
		}
		if got := r.URL.Query().Get("end_date"); got != "2026-04-01" {
			t.Fatalf("unexpected end date: %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(weeklySummaryResponse{
			Weeks: []WeeklySummaryWeek{
				{
					StartDate: "2026-03-30",
					EndDate:   "2026-04-05",
					Submitted: true,
				},
			},
		})
	}))
	defer server.Close()

	client := New("123", "token", server.Client())
	client.BaseURL = server.URL

	summary, err := client.WeeklySummary(context.Background(), time.Date(2026, 4, 1, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("weekly summary failed: %v", err)
	}
	if summary.StartDate != "2026-03-30" || !summary.Submitted {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

func TestSubmitWeekForApprovalSendsExpectedForm(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/daily/submit" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Fatalf("unexpected auth header: %q", got)
		}
		if got := r.Header.Get("Harvest-Account-Id"); got != "123" {
			t.Fatalf("unexpected account header: %q", got)
		}
		if got := r.Header.Get("X-Harvest-Webview-Client"); got != "true" {
			t.Fatalf("unexpected webview header: %q", got)
		}

		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			t.Fatalf("parse content-type: %v", err)
		}
		if mediaType != "multipart/form-data" {
			t.Fatalf("unexpected media type: %q", mediaType)
		}
		if err := r.ParseMultipartForm(4096); err != nil {
			t.Fatalf("parse multipart form: %v", err)
		}
		if got := r.FormValue("of_user"); got != "4833590" {
			t.Fatalf("unexpected of_user: %q", got)
		}
		if got := r.FormValue("submitted_date"); got != "2026-04-01" {
			t.Fatalf("unexpected submitted_date: %q", got)
		}
		if got := r.FormValue("submitted_date_year"); got != "2026" {
			t.Fatalf("unexpected submitted_date_year: %q", got)
		}
		if got := r.FormValue("period_begin"); got != "89" {
			t.Fatalf("unexpected period_begin: %q", got)
		}
		if got := r.FormValue("period_begin_year"); got != "2026" {
			t.Fatalf("unexpected period_begin_year: %q", got)
		}
		if params["boundary"] == "" {
			t.Fatalf("expected multipart boundary")
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New("123", "token", server.Client())
	client.BaseURL = server.URL

	err := client.SubmitWeekForApproval(context.Background(), SubmitWeekInput{
		UserID:     4833590,
		TargetDate: time.Date(2026, 4, 1, 12, 0, 0, 0, time.Local),
		WeekStart:  time.Date(2026, 3, 30, 12, 0, 0, 0, time.Local),
	})
	if err != nil {
		t.Fatalf("submit week failed: %v", err)
	}
}

func TestSubmitWeekForApprovalAcceptsRedirect(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "/time?day=1&month=4&year=2026")
		w.WriteHeader(http.StatusFound)
	}))
	defer server.Close()

	client := New("123", "token", server.Client())
	client.BaseURL = server.URL

	err := client.SubmitWeekForApproval(context.Background(), SubmitWeekInput{
		UserID:     4833590,
		TargetDate: time.Date(2026, 4, 1, 12, 0, 0, 0, time.Local),
		WeekStart:  time.Date(2026, 3, 30, 12, 0, 0, 0, time.Local),
	})
	if err != nil {
		t.Fatalf("submit week failed: %v", err)
	}
}
