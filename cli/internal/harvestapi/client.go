package harvestapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const DefaultBaseURL = "https://api.harvestapp.com/api/v2"

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	AccountID  string
	Token      string
	UserAgent  string
}

type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type Project struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"code,omitempty"`
}

type Task struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type TaskAssignment struct {
	ID       int64 `json:"id"`
	IsActive bool  `json:"is_active"`
	Task     Task  `json:"task"`
}

type ProjectAssignment struct {
	ID              int64            `json:"id"`
	IsActive        bool             `json:"is_active"`
	Project         Project          `json:"project"`
	TaskAssignments []TaskAssignment `json:"task_assignments"`
}

type TimeEntry struct {
	ID        int64   `json:"id"`
	SpentDate string  `json:"spent_date"`
	Hours     float64 `json:"hours"`
	Notes     string  `json:"notes,omitempty"`
	Project   Project `json:"project"`
	Task      Task    `json:"task"`
}

type CreateTimeEntryInput struct {
	ProjectID int64
	TaskID    int64
	SpentDate string
	Hours     float64
	Notes     string
}

type projectAssignmentsPage struct {
	ProjectAssignments []ProjectAssignment `json:"project_assignments"`
	NextPage           int                 `json:"next_page"`
}

type timeEntriesPage struct {
	TimeEntries []TimeEntry `json:"time_entries"`
	NextPage    int         `json:"next_page"`
}

func New(accountID, token string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		BaseURL:    DefaultBaseURL,
		HTTPClient: httpClient,
		AccountID:  accountID,
		Token:      token,
		UserAgent:  "harvest-cli",
	}
}

func (c *Client) Me(ctx context.Context) (User, error) {
	var user User
	if err := c.doJSON(ctx, http.MethodGet, "/users/me", nil, nil, &user); err != nil {
		return User{}, err
	}
	return user, nil
}

func (c *Client) ProjectAssignments(ctx context.Context) ([]ProjectAssignment, error) {
	var all []ProjectAssignment
	page := 1

	for {
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("per_page", "200")

		var payload projectAssignmentsPage
		if err := c.doJSON(ctx, http.MethodGet, "/users/me/project_assignments", query, nil, &payload); err != nil {
			return nil, err
		}

		all = append(all, payload.ProjectAssignments...)
		if payload.NextPage == 0 {
			return all, nil
		}
		page = payload.NextPage
	}
}

func (c *Client) CreateTimeEntry(ctx context.Context, input CreateTimeEntryInput) (TimeEntry, error) {
	payload := struct {
		ProjectID int64   `json:"project_id"`
		TaskID    int64   `json:"task_id"`
		SpentDate string  `json:"spent_date"`
		Hours     float64 `json:"hours"`
		Notes     string  `json:"notes,omitempty"`
	}{
		ProjectID: input.ProjectID,
		TaskID:    input.TaskID,
		SpentDate: input.SpentDate,
		Hours:     input.Hours,
		Notes:     input.Notes,
	}

	var entry TimeEntry
	if err := c.doJSON(ctx, http.MethodPost, "/time_entries", nil, payload, &entry); err != nil {
		return TimeEntry{}, err
	}
	return entry, nil
}

func (c *Client) TimeEntries(ctx context.Context, fromDate, toDate string) ([]TimeEntry, error) {
	var all []TimeEntry
	page := 1

	for {
		query := url.Values{}
		query.Set("from", fromDate)
		query.Set("to", toDate)
		query.Set("page", strconv.Itoa(page))
		query.Set("per_page", "200")

		var payload timeEntriesPage
		if err := c.doJSON(ctx, http.MethodGet, "/time_entries", query, nil, &payload); err != nil {
			return nil, err
		}

		all = append(all, payload.TimeEntries...)
		if payload.NextPage == 0 {
			return all, nil
		}
		page = payload.NextPage
	}
}

func (c *Client) doJSON(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	endpoint := strings.TrimRight(c.BaseURL, "/") + path
	requestURL, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("parse request url: %w", err)
	}
	if query != nil {
		requestURL.RawQuery = query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		buffer := &bytes.Buffer{}
		if err := json.NewEncoder(buffer).Encode(body); err != nil {
			return fmt.Errorf("encode request body: %w", err)
		}
		bodyReader = buffer
	}

	request, err := http.NewRequestWithContext(ctx, method, requestURL.String(), bodyReader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+c.Token)
	request.Header.Set("Harvest-Account-Id", c.AccountID)
	request.Header.Set("User-Agent", c.UserAgent)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := c.HTTPClient.Do(request)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return decodeAPIError(response)
	}

	if out == nil {
		io.Copy(io.Discard, response.Body)
		return nil
	}

	if err := json.NewDecoder(response.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func decodeAPIError(response *http.Response) error {
	payload, _ := io.ReadAll(response.Body)

	var fields map[string]any
	if err := json.Unmarshal(payload, &fields); err == nil {
		for _, key := range []string{"error", "message", "description"} {
			if value, ok := fields[key].(string); ok && strings.TrimSpace(value) != "" {
				return fmt.Errorf("%s: %s", response.Status, value)
			}
		}
	}

	message := strings.TrimSpace(string(payload))
	if message == "" {
		message = response.Status
	}
	return fmt.Errorf("%s: %s", response.Status, message)
}
