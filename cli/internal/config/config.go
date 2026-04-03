package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Values struct {
	AccountID      string `json:"account_id,omitempty"`
	Token          string `json:"token,omitempty"`
	DefaultProject string `json:"default_project,omitempty"`
	DefaultTask    string `json:"default_task,omitempty"`
}

type Update struct {
	AccountID      *string
	Token          *string
	DefaultProject *string
	DefaultTask    *string
}

type RedactedValues struct {
	AccountID      string `json:"account_id,omitempty"`
	TokenPresent   bool   `json:"token_present"`
	DefaultProject string `json:"default_project,omitempty"`
	DefaultTask    string `json:"default_task,omitempty"`
}

type Store struct {
	Path   string
	Getenv func(string) string
}

func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}

	return filepath.Join(dir, "harvest", "config.json"), nil
}

func NewStore(path string, getenv func(string) string) *Store {
	if getenv == nil {
		getenv = os.Getenv
	}

	return &Store{
		Path:   path,
		Getenv: getenv,
	}
}

func (s *Store) LoadFile() (Values, error) {
	data, err := os.ReadFile(s.Path)
	if errors.Is(err, os.ErrNotExist) {
		return Values{}, nil
	}
	if err != nil {
		return Values{}, fmt.Errorf("read config: %w", err)
	}
	if len(data) == 0 {
		return Values{}, nil
	}

	var values Values
	if err := json.Unmarshal(data, &values); err != nil {
		return Values{}, fmt.Errorf("decode config: %w", err)
	}

	return values, nil
}

func (s *Store) Effective(overrides Values) (Values, error) {
	values, err := s.LoadFile()
	if err != nil {
		return Values{}, err
	}

	values = applyEnv(values, s.Getenv)
	values = mergeValues(values, overrides)
	return values, nil
}

func (s *Store) Save(update Update) (Values, error) {
	values, err := s.LoadFile()
	if err != nil {
		return Values{}, err
	}

	values = applyUpdate(values, update)

	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return Values{}, fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		return Values{}, fmt.Errorf("encode config: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(s.Path, data, 0o600); err != nil {
		return Values{}, fmt.Errorf("write config: %w", err)
	}

	return values, nil
}

func Redacted(values Values) RedactedValues {
	return RedactedValues{
		AccountID:      values.AccountID,
		TokenPresent:   values.Token != "",
		DefaultProject: values.DefaultProject,
		DefaultTask:    values.DefaultTask,
	}
}

func applyEnv(values Values, getenv func(string) string) Values {
	if value := getenv("HARVEST_ACCOUNT_ID"); value != "" {
		values.AccountID = value
	}
	if value := getenv("HARVEST_TOKEN"); value != "" {
		values.Token = value
	}
	if value := getenv("HARVEST_DEFAULT_PROJECT"); value != "" {
		values.DefaultProject = value
	}
	if value := getenv("HARVEST_DEFAULT_TASK"); value != "" {
		values.DefaultTask = value
	}
	return values
}

func mergeValues(base Values, overrides Values) Values {
	if overrides.AccountID != "" {
		base.AccountID = overrides.AccountID
	}
	if overrides.Token != "" {
		base.Token = overrides.Token
	}
	if overrides.DefaultProject != "" {
		base.DefaultProject = overrides.DefaultProject
	}
	if overrides.DefaultTask != "" {
		base.DefaultTask = overrides.DefaultTask
	}
	return base
}

func applyUpdate(values Values, update Update) Values {
	if update.AccountID != nil {
		values.AccountID = *update.AccountID
	}
	if update.Token != nil {
		values.Token = *update.Token
	}
	if update.DefaultProject != nil {
		values.DefaultProject = *update.DefaultProject
	}
	if update.DefaultTask != nil {
		values.DefaultTask = *update.DefaultTask
	}
	return values
}
