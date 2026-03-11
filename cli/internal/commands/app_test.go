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
	me              harvestapi.User
	assignments     []harvestapi.ProjectAssignment
	createResponse  harvestapi.TimeEntry
	timeEntries     []harvestapi.TimeEntry
	createInput     harvestapi.CreateTimeEntryInput
	projectsCalled  bool
	createWasCalled bool
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

func (f *fakeClient) TimeEntries(context.Context, string, string) ([]harvestapi.TimeEntry, error) {
	return f.timeEntries, nil
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

	if err := app.Execute([]string{"log", "--duration", "45m", "--json"}); err != nil {
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
	if !strings.Contains(stdout.String(), "Usage:\n  harvest log") {
		t.Fatalf("unexpected help output: %q", stdout.String())
	}
}
