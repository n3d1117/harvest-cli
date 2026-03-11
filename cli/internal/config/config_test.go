package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadAndEffectivePrecedence(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.json")
	store := NewStore(path, func(key string) string {
		switch key {
		case "HARVEST_TOKEN":
			return "env-token"
		case "HARVEST_DEFAULT_TASK":
			return "Env Task"
		default:
			return ""
		}
	})

	accountID := "file-account"
	token := "file-token"
	defaultProject := "File Project"

	if _, err := store.Save(Update{
		AccountID:      &accountID,
		Token:          &token,
		DefaultProject: &defaultProject,
	}); err != nil {
		t.Fatalf("save config: %v", err)
	}

	values, err := store.LoadFile()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if values.AccountID != "file-account" {
		t.Fatalf("unexpected account id: %q", values.AccountID)
	}

	effective, err := store.Effective(Values{
		DefaultProject: "Flag Project",
	})
	if err != nil {
		t.Fatalf("effective config: %v", err)
	}

	if effective.AccountID != "file-account" {
		t.Fatalf("expected file account id, got %q", effective.AccountID)
	}
	if effective.Token != "env-token" {
		t.Fatalf("expected env token, got %q", effective.Token)
	}
	if effective.DefaultProject != "Flag Project" {
		t.Fatalf("expected flag project, got %q", effective.DefaultProject)
	}
	if effective.DefaultTask != "Env Task" {
		t.Fatalf("expected env task, got %q", effective.DefaultTask)
	}
}

func TestSaveCanClearValues(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.json")
	store := NewStore(path, nil)

	accountID := "123"
	token := "token"
	defaultTask := "Task"

	if _, err := store.Save(Update{
		AccountID:   &accountID,
		Token:       &token,
		DefaultTask: &defaultTask,
	}); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	empty := ""
	values, err := store.Save(Update{
		DefaultTask: &empty,
	})
	if err != nil {
		t.Fatalf("clear task: %v", err)
	}
	if values.DefaultTask != "" {
		t.Fatalf("expected task to be cleared, got %q", values.DefaultTask)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file should exist: %v", err)
	}
}

func TestRedacted(t *testing.T) {
	t.Parallel()

	values := Redacted(Values{
		AccountID:      "123",
		Token:          "secret",
		DefaultProject: "Acme",
		DefaultTask:    "Dev",
	})

	if !values.TokenPresent {
		t.Fatalf("expected token to be marked as present")
	}
	if values.AccountID != "123" {
		t.Fatalf("unexpected account id: %q", values.AccountID)
	}
}
