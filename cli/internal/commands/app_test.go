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

type fakePrompt struct {
	values []string
}

func (f *fakePrompt) Prompt(_ string) (string, error) {
	value := f.values[0]
	f.values = f.values[1:]
	return value, nil
}

func (f *fakePrompt) PromptSecret(_ string) (string, error) {
	return f.Prompt("")
}

type fakeClient struct {
	me                 harvestapi.User
	assignments        []harvestapi.ProjectAssignment
	createResponse     harvestapi.TimeEntry
	updateResponse     harvestapi.TimeEntry
	timeEntry          harvestapi.TimeEntry
	timeEntries        []harvestapi.TimeEntry
	weeklySummaries    []harvestapi.WeeklySummaryWeek
	weeklySummaryErr   error
	submitWeekErr      error
	createInput        harvestapi.CreateTimeEntryInput
	updateInput        harvestapi.UpdateTimeEntryInput
	submitWeekInput    harvestapi.SubmitWeekInput
	timeEntryID        int64
	updateID           int64
	deleteID           int64
	weeklySummaryFor   time.Time
	projectsCalled     bool
	createWasCalled    bool
	updateWasCalled    bool
	deleteWasCalled    bool
	timeEntryWasRead   bool
	submitWeekCalls    int
	weeklySummaryCalls int
}

func (f *fakeClient) Me(context.Context) (harvestapi.User, error) {
	return f.me, nil
}

func (f *fakeClient) ProjectAssignments(context.Context) ([]harvestapi.ProjectAssignment, error) {
	f.projectsCalled = true
	return f.assignments, nil
}

func (f *fakeClient) CreateTimeEntry(_ context.Context, input harvestapi.CreateTimeEntryInput) (harvestapi.TimeEntry, error) {
	f.createInput = input
	f.createWasCalled = true
	return f.createResponse, nil
}

func (f *fakeClient) TimeEntry(_ context.Context, id int64) (harvestapi.TimeEntry, error) {
	f.timeEntryWasRead = true
	f.timeEntryID = id
	return f.timeEntry, nil
}

func (f *fakeClient) UpdateTimeEntry(_ context.Context, id int64, input harvestapi.UpdateTimeEntryInput) (harvestapi.TimeEntry, error) {
	f.updateID = id
	f.updateInput = input
	f.updateWasCalled = true
	return f.updateResponse, nil
}

func (f *fakeClient) DeleteTimeEntry(_ context.Context, id int64) error {
	f.deleteID = id
	f.deleteWasCalled = true
	return nil
}

func (f *fakeClient) TimeEntries(context.Context, string, string) ([]harvestapi.TimeEntry, error) {
	return f.timeEntries, nil
}

func (f *fakeClient) WeeklySummary(_ context.Context, targetDate time.Time) (harvestapi.WeeklySummaryWeek, error) {
	f.weeklySummaryCalls++
	f.weeklySummaryFor = targetDate
	if f.weeklySummaryErr != nil {
		return harvestapi.WeeklySummaryWeek{}, f.weeklySummaryErr
	}
	if len(f.weeklySummaries) == 0 {
		return harvestapi.WeeklySummaryWeek{}, nil
	}
	summary := f.weeklySummaries[0]
	if len(f.weeklySummaries) > 1 {
		f.weeklySummaries = f.weeklySummaries[1:]
	}
	return summary, nil
}

func (f *fakeClient) SubmitWeekForApproval(_ context.Context, input harvestapi.SubmitWeekInput) error {
	f.submitWeekCalls++
	f.submitWeekInput = input
	return f.submitWeekErr
}

func TestLoginSavesValidatedCredentials(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(configPath, nil)
	client := &fakeClient{
		me: harvestapi.User{
			ID:        1,
			FirstName: "Ned",
			LastName:  "Tester",
			Email:     "ned@example.com",
		},
	}

	var stdout bytes.Buffer
	app := &App{
		Store:   store,
		Prompt:  &fakePrompt{values: []string{"123", "secret-token"}},
		Stdout:  &stdout,
		Stderr:  &bytes.Buffer{},
		Now:     time.Now,
		Context: context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) {
			if values.AccountID != "123" || values.Token != "secret-token" {
				t.Fatalf("unexpected credentials: %+v", values)
			}
			return client, nil
		},
	}

	if err := app.Execute([]string{"login"}); err != nil {
		t.Fatalf("login failed: %v", err)
	}

	values, err := store.LoadFile()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if values.Token != "secret-token" {
		t.Fatalf("expected token to be saved, got %q", values.Token)
	}
	if !strings.Contains(stdout.String(), "Saved Harvest credentials") {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
}

func TestWhoamiJSON(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	app := &App{
		Store:   config.NewStore(filepath.Join(t.TempDir(), "config.json"), nil),
		Prompt:  &fakePrompt{},
		Stdout:  &stdout,
		Stderr:  &bytes.Buffer{},
		Now:     time.Now,
		Context: context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) {
			return &fakeClient{
				me: harvestapi.User{ID: 1, Email: "ned@example.com"},
			}, nil
		},
	}

	accountID := "123"
	token := "secret"
	if _, err := app.Store.Save(config.Update{AccountID: &accountID, Token: &token}); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	if err := app.Execute([]string{"whoami", "--json"}); err != nil {
		t.Fatalf("whoami failed: %v", err)
	}

	var payload struct {
		OK   bool            `json:"ok"`
		User harvestapi.User `json:"user"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if !payload.OK || payload.User.Email != "ned@example.com" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestLogUsesDefaultsAndOptionalNotes(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(configPath, nil)
	accountID := "123"
	token := "secret"
	defaultProject := "Acme"
	defaultTask := "Development"
	if _, err := store.Save(config.Update{
		AccountID:      &accountID,
		Token:          &token,
		DefaultProject: &defaultProject,
		DefaultTask:    &defaultTask,
	}); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	client := &fakeClient{
		assignments: []harvestapi.ProjectAssignment{
			{
				ID:       1,
				IsActive: true,
				Project:  harvestapi.Project{ID: 11, Name: "Acme"},
				TaskAssignments: []harvestapi.TaskAssignment{
					{ID: 2, IsActive: true, Task: harvestapi.Task{ID: 22, Name: "Development"}},
				},
			},
		},
		createResponse: harvestapi.TimeEntry{ID: 44},
	}

	var stdout bytes.Buffer
	app := &App{
		Store:  store,
		Prompt: &fakePrompt{},
		Stdout: &stdout,
		Stderr: &bytes.Buffer{},
		Now: func() time.Time {
			return time.Date(2026, 3, 11, 9, 0, 0, 0, time.Local)
		},
		Context: context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) {
			return client, nil
		},
	}

	if err := app.Execute([]string{"log", "create", "--duration", "45m", "--json"}); err != nil {
		t.Fatalf("log failed: %v", err)
	}
	if !client.createWasCalled {
		t.Fatalf("expected create to be called")
	}
	if client.createInput.Notes != "" {
		t.Fatalf("expected empty notes, got %q", client.createInput.Notes)
	}
	if client.createInput.SpentDate != "2026-03-11" {
		t.Fatalf("unexpected spent date: %q", client.createInput.SpentDate)
	}
	if client.createInput.Hours != 0.75 {
		t.Fatalf("unexpected hours: %v", client.createInput.Hours)
	}
}

func TestLogAcceptsTodayLiteral(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(configPath, nil)
	accountID := "123"
	token := "secret"
	defaultProject := "Acme"
	defaultTask := "Development"
	if _, err := store.Save(config.Update{
		AccountID:      &accountID,
		Token:          &token,
		DefaultProject: &defaultProject,
		DefaultTask:    &defaultTask,
	}); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	client := &fakeClient{
		assignments: []harvestapi.ProjectAssignment{
			{
				ID:       1,
				IsActive: true,
				Project:  harvestapi.Project{ID: 11, Name: "Acme"},
				TaskAssignments: []harvestapi.TaskAssignment{
					{ID: 2, IsActive: true, Task: harvestapi.Task{ID: 22, Name: "Development"}},
				},
			},
		},
		createResponse: harvestapi.TimeEntry{ID: 44},
	}

	app := &App{
		Store:  store,
		Prompt: &fakePrompt{},
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Now: func() time.Time {
			return time.Date(2026, 3, 11, 9, 0, 0, 0, time.Local)
		},
		Context: context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) {
			return client, nil
		},
	}

	if err := app.Execute([]string{"log", "create", "--duration", "30m", "--date", "today"}); err != nil {
		t.Fatalf("log failed: %v", err)
	}
	if client.createInput.SpentDate != "2026-03-11" {
		t.Fatalf("unexpected spent date: %q", client.createInput.SpentDate)
	}
}

func TestLogDryRunJSONResolvesWithoutCreatingEntry(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(configPath, nil)
	accountID := "123"
	token := "secret"
	defaultProject := "Acme"
	defaultTask := "Development"
	if _, err := store.Save(config.Update{
		AccountID:      &accountID,
		Token:          &token,
		DefaultProject: &defaultProject,
		DefaultTask:    &defaultTask,
	}); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	client := &fakeClient{
		assignments: []harvestapi.ProjectAssignment{
			{
				ID:       1,
				IsActive: true,
				Project:  harvestapi.Project{ID: 11, Name: "Acme"},
				TaskAssignments: []harvestapi.TaskAssignment{
					{ID: 2, IsActive: true, Task: harvestapi.Task{ID: 22, Name: "Development"}},
				},
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
			return time.Date(2026, 3, 11, 9, 0, 0, 0, time.Local)
		},
		Context: context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) {
			return client, nil
		},
	}

	if err := app.Execute([]string{"log", "create", "--duration", "45m", "--dry-run", "--json"}); err != nil {
		t.Fatalf("log dry run failed: %v", err)
	}
	if client.createWasCalled {
		t.Fatalf("expected dry run to skip create")
	}
	if !client.projectsCalled {
		t.Fatalf("expected dry run to resolve project assignments")
	}

	var payload struct {
		OK     bool `json:"ok"`
		DryRun bool `json:"dry_run"`
		Entry  struct {
			ID        *int64  `json:"id"`
			Date      string  `json:"date"`
			Hours     float64 `json:"hours"`
			ProjectID int64   `json:"project_id"`
			Project   string  `json:"project"`
			TaskID    int64   `json:"task_id"`
			Task      string  `json:"task"`
		} `json:"entry"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if !payload.OK || !payload.DryRun {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if payload.Entry.ID != nil {
		t.Fatalf("expected no entry id in dry run, got %v", *payload.Entry.ID)
	}
	if payload.Entry.Date != "2026-03-11" || payload.Entry.Hours != 0.75 {
		t.Fatalf("unexpected entry payload: %+v", payload.Entry)
	}
	if payload.Entry.ProjectID != 11 || payload.Entry.TaskID != 22 {
		t.Fatalf("unexpected resolved ids: %+v", payload.Entry)
	}
}

func TestLogUpdateDryRunJSONChangesOnlyDate(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(configPath, nil)
	accountID := "123"
	token := "secret"
	if _, err := store.Save(config.Update{AccountID: &accountID, Token: &token}); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	client := &fakeClient{
		timeEntry: harvestapi.TimeEntry{
			ID:        44,
			SpentDate: "2026-03-11",
			Hours:     1.5,
			Notes:     "Original note",
			Project:   harvestapi.Project{ID: 11, Name: "Acme"},
			Task:      harvestapi.Task{ID: 22, Name: "Development"},
		},
	}

	var stdout bytes.Buffer
	app := &App{
		Store:  store,
		Prompt: &fakePrompt{},
		Stdout: &stdout,
		Stderr: &bytes.Buffer{},
		Now: func() time.Time {
			return time.Date(2026, 3, 12, 9, 0, 0, 0, time.Local)
		},
		Context: context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) {
			return client, nil
		},
	}

	if err := app.Execute([]string{"log", "update", "--id", "44", "--date", "today", "--dry-run", "--json"}); err != nil {
		t.Fatalf("log update dry run failed: %v", err)
	}
	if client.updateWasCalled {
		t.Fatalf("expected dry run to skip update")
	}
	if !client.timeEntryWasRead || client.timeEntryID != 44 {
		t.Fatalf("expected time entry read, got %+v", client)
	}

	var payload struct {
		OK     bool `json:"ok"`
		DryRun bool `json:"dry_run"`
		Entry  struct {
			ID        *int64  `json:"id"`
			Date      string  `json:"date"`
			Hours     float64 `json:"hours"`
			Notes     *string `json:"notes"`
			ProjectID int64   `json:"project_id"`
			Project   string  `json:"project"`
			TaskID    int64   `json:"task_id"`
			Task      string  `json:"task"`
		} `json:"entry"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if !payload.OK || !payload.DryRun || payload.Entry.ID == nil || *payload.Entry.ID != 44 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if payload.Entry.Date != "2026-03-12" || payload.Entry.Hours != 1.5 {
		t.Fatalf("unexpected entry payload: %+v", payload.Entry)
	}
	if payload.Entry.Notes == nil || *payload.Entry.Notes != "Original note" {
		t.Fatalf("unexpected notes: %+v", payload.Entry.Notes)
	}
}

func TestLogUpdateChangesHoursAndNotes(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(configPath, nil)
	accountID := "123"
	token := "secret"
	if _, err := store.Save(config.Update{AccountID: &accountID, Token: &token}); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	client := &fakeClient{
		updateResponse: harvestapi.TimeEntry{
			ID:        44,
			SpentDate: "2026-03-11",
			Hours:     0.75,
			Notes:     "",
			Project:   harvestapi.Project{ID: 11, Name: "Acme"},
			Task:      harvestapi.Task{ID: 22, Name: "Development"},
		},
	}

	app := &App{
		Store:   store,
		Prompt:  &fakePrompt{},
		Stdout:  &bytes.Buffer{},
		Stderr:  &bytes.Buffer{},
		Now:     time.Now,
		Context: context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) {
			return client, nil
		},
	}

	if err := app.Execute([]string{"log", "update", "--id", "44", "--duration", "45m", "--notes", ""}); err != nil {
		t.Fatalf("log update failed: %v", err)
	}
	if !client.updateWasCalled || client.updateID != 44 {
		t.Fatalf("expected update call, got %+v", client)
	}
	if client.updateInput.Hours == nil || *client.updateInput.Hours != 0.75 {
		t.Fatalf("unexpected hours update: %+v", client.updateInput)
	}
	if client.updateInput.Notes == nil || *client.updateInput.Notes != "" {
		t.Fatalf("expected notes clear, got %+v", client.updateInput.Notes)
	}
}

func TestLogUpdateChangesProjectAndTask(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(configPath, nil)
	accountID := "123"
	token := "secret"
	if _, err := store.Save(config.Update{AccountID: &accountID, Token: &token}); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	client := &fakeClient{
		assignments: []harvestapi.ProjectAssignment{
			{
				ID:       1,
				IsActive: true,
				Project:  harvestapi.Project{ID: 11, Name: "Acme"},
				TaskAssignments: []harvestapi.TaskAssignment{
					{ID: 2, IsActive: true, Task: harvestapi.Task{ID: 21, Name: "Design"}},
					{ID: 3, IsActive: true, Task: harvestapi.Task{ID: 22, Name: "Development"}},
				},
			},
		},
		updateResponse: harvestapi.TimeEntry{
			ID:        44,
			SpentDate: "2026-03-11",
			Hours:     1.5,
			Project:   harvestapi.Project{ID: 11, Name: "Acme"},
			Task:      harvestapi.Task{ID: 21, Name: "Design"},
		},
	}

	app := &App{
		Store:   store,
		Prompt:  &fakePrompt{},
		Stdout:  &bytes.Buffer{},
		Stderr:  &bytes.Buffer{},
		Now:     time.Now,
		Context: context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) {
			return client, nil
		},
	}

	if err := app.Execute([]string{"log", "update", "--id", "44", "--project", "Acme", "--task", "Design"}); err != nil {
		t.Fatalf("log update failed: %v", err)
	}
	if !client.projectsCalled {
		t.Fatalf("expected project lookup")
	}
	if client.updateInput.ProjectID == nil || *client.updateInput.ProjectID != 11 {
		t.Fatalf("unexpected project update: %+v", client.updateInput)
	}
	if client.updateInput.TaskID == nil || *client.updateInput.TaskID != 21 {
		t.Fatalf("unexpected task update: %+v", client.updateInput)
	}
}

func TestLogUpdateRequiresID(t *testing.T) {
	t.Parallel()

	app := &App{
		Store:         config.NewStore(filepath.Join(t.TempDir(), "config.json"), nil),
		Prompt:        &fakePrompt{},
		Stdout:        &bytes.Buffer{},
		Stderr:        &bytes.Buffer{},
		Now:           time.Now,
		Context:       context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) { return &fakeClient{}, nil },
	}

	err := app.Execute([]string{"log", "update", "--duration", "45m"})
	if err == nil || err.Error() != "--id is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLogUpdateRequiresAtLeastOneChange(t *testing.T) {
	t.Parallel()

	app := &App{
		Store:         config.NewStore(filepath.Join(t.TempDir(), "config.json"), nil),
		Prompt:        &fakePrompt{},
		Stdout:        &bytes.Buffer{},
		Stderr:        &bytes.Buffer{},
		Now:           time.Now,
		Context:       context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) { return &fakeClient{}, nil },
	}

	err := app.Execute([]string{"log", "update", "--id", "44"})
	if err == nil || err.Error() != "pass at least one field to update" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLogUpdateRequiresProjectAndTaskTogether(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(configPath, nil)
	accountID := "123"
	token := "secret"
	if _, err := store.Save(config.Update{AccountID: &accountID, Token: &token}); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	app := &App{
		Store:   store,
		Prompt:  &fakePrompt{},
		Stdout:  &bytes.Buffer{},
		Stderr:  &bytes.Buffer{},
		Now:     time.Now,
		Context: context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) {
			return &fakeClient{}, nil
		},
	}

	err := app.Execute([]string{"log", "update", "--id", "44", "--project", "Acme"})
	if err == nil || err.Error() != "if changing project or task, pass both --project and --task" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLogDeleteDeletesEntry(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(configPath, nil)
	accountID := "123"
	token := "secret"
	if _, err := store.Save(config.Update{AccountID: &accountID, Token: &token}); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	client := &fakeClient{
		timeEntry: harvestapi.TimeEntry{
			ID:        44,
			SpentDate: "2026-03-11",
			Hours:     1,
			Project:   harvestapi.Project{ID: 11, Name: "Acme"},
			Task:      harvestapi.Task{ID: 22, Name: "Development"},
		},
	}

	app := &App{
		Store:   store,
		Prompt:  &fakePrompt{},
		Stdout:  &bytes.Buffer{},
		Stderr:  &bytes.Buffer{},
		Now:     time.Now,
		Context: context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) {
			return client, nil
		},
	}

	if err := app.Execute([]string{"log", "delete", "--id", "44"}); err != nil {
		t.Fatalf("log delete failed: %v", err)
	}
	if !client.timeEntryWasRead || client.timeEntryID != 44 || !client.deleteWasCalled || client.deleteID != 44 {
		t.Fatalf("unexpected delete flow: %+v", client)
	}
}

func TestLogDeleteDryRunJSONFetchesWithoutDeleting(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(configPath, nil)
	accountID := "123"
	token := "secret"
	if _, err := store.Save(config.Update{AccountID: &accountID, Token: &token}); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	client := &fakeClient{
		timeEntry: harvestapi.TimeEntry{
			ID:        44,
			SpentDate: "2026-03-11",
			Hours:     1,
			Notes:     "Delete me",
			Project:   harvestapi.Project{ID: 11, Name: "Acme"},
			Task:      harvestapi.Task{ID: 22, Name: "Development"},
		},
	}

	var stdout bytes.Buffer
	app := &App{
		Store:   store,
		Prompt:  &fakePrompt{},
		Stdout:  &stdout,
		Stderr:  &bytes.Buffer{},
		Now:     time.Now,
		Context: context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) {
			return client, nil
		},
	}

	if err := app.Execute([]string{"log", "delete", "--id", "44", "--dry-run", "--json"}); err != nil {
		t.Fatalf("log delete dry run failed: %v", err)
	}
	if client.deleteWasCalled {
		t.Fatalf("expected dry run to skip delete")
	}

	var payload struct {
		OK     bool `json:"ok"`
		DryRun bool `json:"dry_run"`
		Entry  struct {
			ID *int64 `json:"id"`
		} `json:"entry"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if !payload.OK || !payload.DryRun || payload.Entry.ID == nil || *payload.Entry.ID != 44 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestTodayJSON(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(configPath, nil)
	accountID := "123"
	token := "secret"
	if _, err := store.Save(config.Update{AccountID: &accountID, Token: &token}); err != nil {
		t.Fatalf("seed config: %v", err)
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
		Context: context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) {
			return &fakeClient{
				timeEntries: []harvestapi.TimeEntry{
					{ID: 1, Hours: 1.25},
					{ID: 2, Hours: 0.75},
				},
			}, nil
		},
	}

	if err := app.Execute([]string{"today", "--json"}); err != nil {
		t.Fatalf("today failed: %v", err)
	}

	var payload struct {
		OK         bool                   `json:"ok"`
		TotalHours float64                `json:"total_hours"`
		Entries    []harvestapi.TimeEntry `json:"entries"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.TotalHours != 2 {
		t.Fatalf("unexpected total: %v", payload.TotalHours)
	}
	if len(payload.Entries) != 2 {
		t.Fatalf("unexpected entry count: %d", len(payload.Entries))
	}
}

func TestRecentJSON(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(configPath, nil)
	accountID := "123"
	token := "secret"
	if _, err := store.Save(config.Update{AccountID: &accountID, Token: &token}); err != nil {
		t.Fatalf("seed config: %v", err)
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
		Context: context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) {
			return &fakeClient{
				timeEntries: []harvestapi.TimeEntry{
					{ID: 2, SpentDate: "2026-03-09", Hours: 0.5, Project: harvestapi.Project{Name: "Acme"}, Task: harvestapi.Task{Name: "Design"}},
					{ID: 3, SpentDate: "2026-03-11", Hours: 1, Project: harvestapi.Project{Name: "Acme"}, Task: harvestapi.Task{Name: "Development"}},
					{ID: 1, SpentDate: "2026-03-11", Hours: 0.25, Project: harvestapi.Project{Name: "Acme"}, Task: harvestapi.Task{Name: "Review"}},
				},
			}, nil
		},
	}

	if err := app.Execute([]string{"recent", "--limit", "2", "--json"}); err != nil {
		t.Fatalf("recent failed: %v", err)
	}

	var payload struct {
		OK      bool                   `json:"ok"`
		From    string                 `json:"from"`
		To      string                 `json:"to"`
		Entries []harvestapi.TimeEntry `json:"entries"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if !payload.OK {
		t.Fatalf("expected ok payload")
	}
	if payload.From != "2025-12-12" || payload.To != "2026-03-11" {
		t.Fatalf("unexpected window: %+v", payload)
	}
	if len(payload.Entries) != 2 {
		t.Fatalf("unexpected entry count: %d", len(payload.Entries))
	}
	if payload.Entries[0].ID != 3 || payload.Entries[1].ID != 1 {
		t.Fatalf("unexpected entry order: %+v", payload.Entries)
	}
}

func TestRecentTextIncludesIDs(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(configPath, nil)
	accountID := "123"
	token := "secret"
	if _, err := store.Save(config.Update{AccountID: &accountID, Token: &token}); err != nil {
		t.Fatalf("seed config: %v", err)
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
		Context: context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) {
			return &fakeClient{
				timeEntries: []harvestapi.TimeEntry{
					{ID: 3, SpentDate: "2026-03-11", Hours: 1, Project: harvestapi.Project{Name: "Acme"}, Task: harvestapi.Task{Name: "Development"}},
				},
			}, nil
		},
	}

	if err := app.Execute([]string{"recent"}); err != nil {
		t.Fatalf("recent failed: %v", err)
	}
	if !strings.HasPrefix(stdout.String(), "ID") {
		t.Fatalf("expected ID column, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "3") {
		t.Fatalf("expected entry id in output, got %q", stdout.String())
	}
}

func TestTodayTextIncludesIDs(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(configPath, nil)
	accountID := "123"
	token := "secret"
	if _, err := store.Save(config.Update{AccountID: &accountID, Token: &token}); err != nil {
		t.Fatalf("seed config: %v", err)
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
		Context: context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) {
			return &fakeClient{
				timeEntries: []harvestapi.TimeEntry{
					{ID: 1, SpentDate: "2026-03-11", Hours: 1.25, Project: harvestapi.Project{Name: "Acme"}, Task: harvestapi.Task{Name: "Development"}},
				},
			}, nil
		},
	}

	if err := app.Execute([]string{"today"}); err != nil {
		t.Fatalf("today failed: %v", err)
	}
	if !strings.HasPrefix(stdout.String(), "ID") {
		t.Fatalf("expected ID column, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "1") {
		t.Fatalf("expected entry id in output, got %q", stdout.String())
	}
}

func TestHelpText(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	app := &App{
		Store:         config.NewStore(filepath.Join(t.TempDir(), "config.json"), nil),
		Prompt:        &fakePrompt{},
		Stdout:        &stdout,
		Stderr:        &bytes.Buffer{},
		Now:           time.Now,
		Context:       context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) { return &fakeClient{}, nil },
	}

	if err := app.Execute([]string{"help", "log"}); err != nil {
		t.Fatalf("help failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "Usage:\n  harvest log <command>") {
		t.Fatalf("unexpected help output: %q", stdout.String())
	}
}

func TestHelpTextForLogUpdate(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	app := &App{
		Store:         config.NewStore(filepath.Join(t.TempDir(), "config.json"), nil),
		Prompt:        &fakePrompt{},
		Stdout:        &stdout,
		Stderr:        &bytes.Buffer{},
		Now:           time.Now,
		Context:       context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) { return &fakeClient{}, nil },
	}

	if err := app.Execute([]string{"help", "log", "update"}); err != nil {
		t.Fatalf("help failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "Usage:\n  harvest log update --id <entry-id>") {
		t.Fatalf("unexpected help output: %q", stdout.String())
	}
}

func TestFlatLogFormIsRejected(t *testing.T) {
	t.Parallel()

	app := &App{
		Store:         config.NewStore(filepath.Join(t.TempDir(), "config.json"), nil),
		Prompt:        &fakePrompt{},
		Stdout:        &bytes.Buffer{},
		Stderr:        &bytes.Buffer{},
		Now:           time.Now,
		Context:       context.Background(),
		ClientFactory: func(values config.Values) (HarvestService, error) { return &fakeClient{}, nil },
	}

	err := app.Execute([]string{"log", "--duration", "45m"})
	if err == nil || err.Error() != "unknown log subcommand \"--duration\"" {
		t.Fatalf("unexpected error: %v", err)
	}
}
