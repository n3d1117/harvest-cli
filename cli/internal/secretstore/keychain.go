package secretstore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

const (
	ServiceSubmitPassword = "harvest.submit.password"
	ServiceSubmitSession  = "harvest.submit.session"
)

var ErrNotFound = errors.New("secret not found")

type Store struct {
	CommandRunner func(context.Context, string, ...string) *exec.Cmd
}

func New() *Store {
	return &Store{
		CommandRunner: exec.CommandContext,
	}
}

func (s *Store) Save(ctx context.Context, service, account, secret string) error {
	cmd := s.command(ctx,
		"add-generic-password",
		"-U",
		"-a", account,
		"-s", service,
		"-w", secret,
	)
	return s.runDiscard(cmd)
}

func (s *Store) Load(ctx context.Context, service, account string) (string, error) {
	cmd := s.command(ctx,
		"find-generic-password",
		"-a", account,
		"-s", service,
		"-w",
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if isNotFound(stderr.String()) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("read keychain item %q for %q: %w", service, account, err)
	}

	return strings.TrimRight(stdout.String(), "\n"), nil
}

func (s *Store) Delete(ctx context.Context, service, account string) error {
	cmd := s.command(ctx,
		"delete-generic-password",
		"-a", account,
		"-s", service,
	)

	var stderr bytes.Buffer
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if isNotFound(stderr.String()) {
			return nil
		}
		return fmt.Errorf("delete keychain item %q for %q: %w", service, account, err)
	}

	return nil
}

func (s *Store) Exists(ctx context.Context, service, account string) (bool, error) {
	_, err := s.Load(ctx, service, account)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	return false, err
}

func (s *Store) command(ctx context.Context, args ...string) *exec.Cmd {
	return s.CommandRunner(ctx, "security", args...)
}

func (s *Store) runDiscard(cmd *exec.Cmd) error {
	var stderr bytes.Buffer
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("write keychain item: %s: %w", strings.TrimSpace(stderr.String()), err)
	}
	return nil
}

func isNotFound(stderr string) bool {
	value := strings.ToLower(stderr)
	return strings.Contains(value, "could not be found") || strings.Contains(value, "item not found")
}
