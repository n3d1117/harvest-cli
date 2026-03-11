package commands

import (
	"testing"
	"time"

	"harvest/internal/harvestapi"
)

func TestParseDurationHours(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  float64
	}{
		{input: "45m", want: 0.75},
		{input: "1h30m", want: 1.5},
		{input: "2h", want: 2},
	}

	for _, test := range tests {
		test := test
		t.Run(test.input, func(t *testing.T) {
			t.Parallel()
			got, err := ParseDurationHours(test.input)
			if err != nil {
				t.Fatalf("parse duration: %v", err)
			}
			if got != test.want {
				t.Fatalf("expected %v, got %v", test.want, got)
			}
		})
	}
}

func TestParseDurationHoursRejectsInvalid(t *testing.T) {
	t.Parallel()

	if _, err := ParseDurationHours("nope"); err == nil {
		t.Fatalf("expected invalid duration error")
	}
	if _, err := ParseDurationHours("0m"); err == nil {
		t.Fatalf("expected zero duration error")
	}
}

func TestResolveProjectTask(t *testing.T) {
	t.Parallel()

	assignments := []harvestapi.ProjectAssignment{
		{
			ID:       1,
			IsActive: true,
			Project:  harvestapi.Project{ID: 10, Name: "Acme"},
			TaskAssignments: []harvestapi.TaskAssignment{
				{ID: 100, IsActive: true, Task: harvestapi.Task{ID: 101, Name: "Development"}},
			},
		},
	}

	pair, err := ResolveProjectTask(assignments, "acme", "development")
	if err != nil {
		t.Fatalf("resolve pair: %v", err)
	}
	if pair.ProjectID != 10 || pair.TaskID != 101 {
		t.Fatalf("unexpected pair: %+v", pair)
	}
}

func TestResolveProjectTaskMissingAndAmbiguous(t *testing.T) {
	t.Parallel()

	assignments := []harvestapi.ProjectAssignment{
		{
			ID:       1,
			IsActive: true,
			Project:  harvestapi.Project{ID: 10, Name: "Acme"},
			TaskAssignments: []harvestapi.TaskAssignment{
				{ID: 100, IsActive: true, Task: harvestapi.Task{ID: 101, Name: "Development"}},
			},
		},
		{
			ID:       2,
			IsActive: true,
			Project:  harvestapi.Project{ID: 20, Name: "Acme"},
			TaskAssignments: []harvestapi.TaskAssignment{
				{ID: 200, IsActive: true, Task: harvestapi.Task{ID: 201, Name: "Design"}},
			},
		},
	}

	if _, err := ResolveProjectTask(assignments, "Missing", "Development"); err == nil {
		t.Fatalf("expected missing project error")
	}
	if _, err := ResolveProjectTask(assignments, "Acme", "Development"); err == nil {
		t.Fatalf("expected ambiguous project error")
	}
}

func TestResolveDateInput(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 11, 12, 0, 0, 0, time.Local)

	got, err := resolveDateInput("", now)
	if err != nil {
		t.Fatalf("empty date should resolve: %v", err)
	}
	if got != "2026-03-11" {
		t.Fatalf("unexpected resolved date: %q", got)
	}

	got, err = resolveDateInput("today", now)
	if err != nil {
		t.Fatalf("today should resolve: %v", err)
	}
	if got != "2026-03-11" {
		t.Fatalf("unexpected today date: %q", got)
	}

	got, err = resolveDateInput("2026-03-09", now)
	if err != nil {
		t.Fatalf("explicit date should resolve: %v", err)
	}
	if got != "2026-03-09" {
		t.Fatalf("unexpected explicit date: %q", got)
	}
}

func TestResolveDateInputRejectsInvalid(t *testing.T) {
	t.Parallel()

	if _, err := resolveDateInput("11-03-2026", time.Now()); err == nil {
		t.Fatalf("expected invalid date error")
	}
}
