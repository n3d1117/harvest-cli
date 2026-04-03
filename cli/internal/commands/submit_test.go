package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"harvest/internal/config"
	"harvest/internal/harvestapi"
)

func TestSubmitWeekDryRunUsesPATAuthOnly(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(configPath, nil)
	accountID := "1450071"
	token := "secret-token"
	if _, err := store.Save(config.Update{AccountID: &accountID, Token: &token}); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	client := &fakeClient{
		weeklySummaries: []harvestapi.WeeklySummaryWeek{
			{
				StartDate: "2026-03-09",
				EndDate:   "2026-03-15",
			},
		},
	}

	var stdout bytes.Buffer
	app := &App{
		Store:   store,
		Prompt:  &fakePrompt{},
		Stdout:  &stdout,
		Stderr:  &bytes.Buffer{},
		Now:     func() time.Time { return time.Date(2026, 3, 11, 12, 0, 0, 0, time.Local) },
		Context: context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) {
			if values.AccountID != accountID || values.Token != token {
				t.Fatalf("unexpected credentials: %+v", values)
			}
			return client, nil
		},
	}

	if err := app.Execute([]string{"submit", "week", "--date", "today", "--dry-run", "--json"}); err != nil {
		t.Fatalf("submit week dry run failed: %v", err)
	}

	if client.submitWeekCalls != 0 {
		t.Fatalf("expected dry run to skip submit, got %d submit calls", client.submitWeekCalls)
	}
	if client.weeklySummaryCalls != 1 {
		t.Fatalf("expected one weekly summary call, got %d", client.weeklySummaryCalls)
	}
	if client.weeklySummaryFor.Format("2006-01-02") != "2026-03-11" {
		t.Fatalf("unexpected weekly summary date: %s", client.weeklySummaryFor.Format("2006-01-02"))
	}
	if !strings.Contains(stdout.String(), `"action": "would_submit"`) {
		t.Fatalf("expected dry run action in output, got %q", stdout.String())
	}
}

func TestSubmitWeekSubmitsAndVerifiesStatus(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(configPath, nil)
	accountID := "1450071"
	token := "secret-token"
	if _, err := store.Save(config.Update{AccountID: &accountID, Token: &token}); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	client := &fakeClient{
		me: harvestapi.User{ID: 4833590},
		weeklySummaries: []harvestapi.WeeklySummaryWeek{
			{
				StartDate: "2026-03-09",
				EndDate:   "2026-03-15",
			},
			{
				StartDate: "2026-03-09",
				EndDate:   "2026-03-15",
				Submitted: true,
			},
		},
	}

	var stdout bytes.Buffer
	app := &App{
		Store:   store,
		Prompt:  &fakePrompt{},
		Stdout:  &stdout,
		Stderr:  &bytes.Buffer{},
		Now:     func() time.Time { return time.Date(2026, 3, 11, 12, 0, 0, 0, time.Local) },
		Context: context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) {
			return client, nil
		},
	}

	if err := app.Execute([]string{"submit", "week", "--date", "today"}); err != nil {
		t.Fatalf("submit week failed: %v", err)
	}

	if client.weeklySummaryCalls != 2 {
		t.Fatalf("expected two weekly summary calls, got %d", client.weeklySummaryCalls)
	}
	if client.submitWeekCalls != 1 {
		t.Fatalf("expected one submit call, got %d", client.submitWeekCalls)
	}
	if client.submitWeekInput.UserID != 4833590 {
		t.Fatalf("unexpected user id: %d", client.submitWeekInput.UserID)
	}
	if client.submitWeekInput.TargetDate.Format("2006-01-02") != "2026-03-11" {
		t.Fatalf("unexpected target date: %s", client.submitWeekInput.TargetDate.Format("2006-01-02"))
	}
	if client.submitWeekInput.WeekStart.Format("2006-01-02") != "2026-03-09" {
		t.Fatalf("unexpected week start: %s", client.submitWeekInput.WeekStart.Format("2006-01-02"))
	}
	if !strings.Contains(stdout.String(), "Submitted week 2026-03-09 to 2026-03-15") {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
}

func TestSubmitWeekDryRunReportsResubmit(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(configPath, nil)
	accountID := "1450071"
	token := "secret-token"
	if _, err := store.Save(config.Update{AccountID: &accountID, Token: &token}); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	client := &fakeClient{
		weeklySummaries: []harvestapi.WeeklySummaryWeek{
			{
				StartDate: "2026-03-09",
				EndDate:   "2026-03-15",
				Submitted: true,
			},
		},
	}

	var stdout bytes.Buffer
	app := &App{
		Store:   store,
		Prompt:  &fakePrompt{},
		Stdout:  &stdout,
		Stderr:  &bytes.Buffer{},
		Now:     func() time.Time { return time.Date(2026, 3, 11, 12, 0, 0, 0, time.Local) },
		Context: context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) {
			return client, nil
		},
	}

	if err := app.Execute([]string{"submit", "week", "--date", "today", "--dry-run"}); err != nil {
		t.Fatalf("submit week dry run failed: %v", err)
	}

	if !strings.Contains(stdout.String(), "Dry run: would resubmit week 2026-03-09 to 2026-03-15") {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
}

func TestSubmitWeekApprovedFails(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(configPath, nil)
	accountID := "1450071"
	token := "secret-token"
	if _, err := store.Save(config.Update{AccountID: &accountID, Token: &token}); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	client := &fakeClient{
		weeklySummaries: []harvestapi.WeeklySummaryWeek{
			{
				StartDate: "2026-03-09",
				EndDate:   "2026-03-15",
				Approved:  true,
			},
		},
	}

	app := &App{
		Store:   store,
		Prompt:  &fakePrompt{},
		Stdout:  &bytes.Buffer{},
		Stderr:  &bytes.Buffer{},
		Now:     func() time.Time { return time.Date(2026, 3, 11, 12, 0, 0, 0, time.Local) },
		Context: context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) {
			return client, nil
		},
	}

	err := app.Execute([]string{"submit", "week", "--date", "today"})
	if err == nil || err.Error() != "week starting 2026-03-09 is already approved" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSubmitWeekJSONPreservesResultShape(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(configPath, nil)
	accountID := "1450071"
	token := "secret-token"
	if _, err := store.Save(config.Update{AccountID: &accountID, Token: &token}); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	client := &fakeClient{
		me: harvestapi.User{ID: 4833590},
		weeklySummaries: []harvestapi.WeeklySummaryWeek{
			{
				StartDate: "2026-03-09",
				EndDate:   "2026-03-15",
				Submitted: true,
			},
			{
				StartDate: "2026-03-09",
				EndDate:   "2026-03-15",
				Submitted: true,
			},
		},
	}

	var stdout bytes.Buffer
	app := &App{
		Store:   store,
		Prompt:  &fakePrompt{},
		Stdout:  &stdout,
		Stderr:  &bytes.Buffer{},
		Now:     func() time.Time { return time.Date(2026, 3, 11, 12, 0, 0, 0, time.Local) },
		Context: context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) {
			return client, nil
		},
	}

	if err := app.Execute([]string{"submit", "week", "--date", "today", "--json"}); err != nil {
		t.Fatalf("submit week json failed: %v", err)
	}

	var payload struct {
		OK     bool                    `json:"ok"`
		Result harvestapi.SubmitResult `json:"result"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if !payload.OK {
		t.Fatalf("expected ok payload")
	}
	if payload.Result.Action != "resubmitted" {
		t.Fatalf("unexpected action: %+v", payload.Result)
	}
	if payload.Result.ReturnTo != "/time?day=11&month=3&year=2026" {
		t.Fatalf("unexpected return_to: %+v", payload.Result)
	}
}
