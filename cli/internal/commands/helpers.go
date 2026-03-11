package commands

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"harvest/internal/harvestapi"
)

type ProjectTaskPair struct {
	ProjectID   int64  `json:"project_id"`
	ProjectName string `json:"project"`
	TaskID      int64  `json:"task_id"`
	TaskName    string `json:"task"`
}

func ParseDurationHours(input string) (float64, error) {
	duration, err := time.ParseDuration(input)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", input, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("duration must be greater than zero")
	}

	hours := duration.Hours()
	return math.Round(hours*10000) / 10000, nil
}

func FlattenProjectTaskPairs(assignments []harvestapi.ProjectAssignment) []ProjectTaskPair {
	pairs := make([]ProjectTaskPair, 0)

	for _, assignment := range assignments {
		if !assignment.IsActive {
			continue
		}
		for _, taskAssignment := range assignment.TaskAssignments {
			if !taskAssignment.IsActive {
				continue
			}
			pairs = append(pairs, ProjectTaskPair{
				ProjectID:   assignment.Project.ID,
				ProjectName: assignment.Project.Name,
				TaskID:      taskAssignment.Task.ID,
				TaskName:    taskAssignment.Task.Name,
			})
		}
	}

	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].ProjectName == pairs[j].ProjectName {
			if pairs[i].TaskName == pairs[j].TaskName {
				return pairs[i].TaskID < pairs[j].TaskID
			}
			return pairs[i].TaskName < pairs[j].TaskName
		}
		return pairs[i].ProjectName < pairs[j].ProjectName
	})

	return pairs
}

func ResolveProjectTask(assignments []harvestapi.ProjectAssignment, projectName, taskName string) (ProjectTaskPair, error) {
	var projectMatches []harvestapi.ProjectAssignment
	for _, assignment := range assignments {
		if !assignment.IsActive {
			continue
		}
		if strings.EqualFold(assignment.Project.Name, projectName) {
			projectMatches = append(projectMatches, assignment)
		}
	}

	if len(projectMatches) == 0 {
		return ProjectTaskPair{}, fmt.Errorf("project %q was not found", projectName)
	}
	if len(projectMatches) > 1 {
		options := make([]string, 0, len(projectMatches))
		for _, match := range projectMatches {
			options = append(options, fmt.Sprintf("%s (#%d)", match.Project.Name, match.Project.ID))
		}
		return ProjectTaskPair{}, fmt.Errorf("project %q is ambiguous: %s", projectName, strings.Join(options, ", "))
	}

	project := projectMatches[0]
	var taskMatches []harvestapi.TaskAssignment
	for _, taskAssignment := range project.TaskAssignments {
		if !taskAssignment.IsActive {
			continue
		}
		if strings.EqualFold(taskAssignment.Task.Name, taskName) {
			taskMatches = append(taskMatches, taskAssignment)
		}
	}

	if len(taskMatches) == 0 {
		return ProjectTaskPair{}, fmt.Errorf("task %q was not found under project %q", taskName, project.Project.Name)
	}
	if len(taskMatches) > 1 {
		options := make([]string, 0, len(taskMatches))
		for _, match := range taskMatches {
			options = append(options, fmt.Sprintf("%s (#%d)", match.Task.Name, match.Task.ID))
		}
		return ProjectTaskPair{}, fmt.Errorf("task %q is ambiguous under project %q: %s", taskName, project.Project.Name, strings.Join(options, ", "))
	}

	match := taskMatches[0]
	return ProjectTaskPair{
		ProjectID:   project.Project.ID,
		ProjectName: project.Project.Name,
		TaskID:      match.Task.ID,
		TaskName:    match.Task.Name,
	}, nil
}

func formatDate(date time.Time) string {
	return date.Format("2006-01-02")
}

func fullName(user harvestapi.User) string {
	return strings.TrimSpace(strings.TrimSpace(user.FirstName) + " " + strings.TrimSpace(user.LastName))
}
