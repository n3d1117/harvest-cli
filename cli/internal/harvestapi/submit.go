package harvestapi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type WeeklySummaryWeek struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Submitted bool   `json:"submitted"`
	Approved  bool   `json:"approved"`
}

type SubmitWeekInput struct {
	UserID     int64
	TargetDate time.Time
	WeekStart  time.Time
}

type SubmitResult struct {
	Action          string `json:"action"`
	WeekStart       string `json:"week_start"`
	WeekEnd         string `json:"week_end"`
	ReturnTo        string `json:"return_to"`
	SubmittedBefore bool   `json:"submitted_before"`
	SubmittedAfter  bool   `json:"submitted_after"`
}

type weeklySummaryResponse struct {
	Weeks []WeeklySummaryWeek `json:"weeks"`
}

func (c *Client) WeeklySummary(ctx context.Context, targetDate time.Time) (WeeklySummaryWeek, error) {
	query := url.Values{}
	date := targetDate.In(time.Local).Format("2006-01-02")
	query.Set("start_date", date)
	query.Set("end_date", date)

	var payload weeklySummaryResponse
	if err := c.doJSON(ctx, http.MethodGet, "/day_entries/weekly_summary", query, nil, &payload); err != nil {
		return WeeklySummaryWeek{}, err
	}
	if len(payload.Weeks) == 0 {
		return WeeklySummaryWeek{}, fmt.Errorf("weekly summary did not include week data")
	}

	return payload.Weeks[0], nil
}

func (c *Client) SubmitWeekForApproval(ctx context.Context, input SubmitWeekInput) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fields := map[string]string{
		"of_user":             strconv.FormatInt(input.UserID, 10),
		"submitted_date":      input.TargetDate.In(time.Local).Format("2006-01-02"),
		"submitted_date_year": strconv.Itoa(input.TargetDate.In(time.Local).Year()),
		"period_begin":        strconv.Itoa(input.WeekStart.In(time.Local).YearDay()),
		"period_begin_year":   strconv.Itoa(input.WeekStart.In(time.Local).Year()),
	}
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return fmt.Errorf("encode submit form: %w", err)
		}
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close submit form: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.submitBaseURL()+"/daily/submit", body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+c.Token)
	request.Header.Set("Harvest-Account-Id", c.AccountID)
	request.Header.Set("User-Agent", c.UserAgent)
	request.Header.Set("Accept", "*/*")
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("X-Harvest-Webview-Client", "true")
	request.Header.Set("X-Local-Timezone-Offset", strconv.Itoa(localTimezoneOffsetMinutes(input.TargetDate.In(time.Local))))

	client := *c.HTTPClient
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK, http.StatusNoContent, http.StatusFound, http.StatusSeeOther:
		_, _ = io.Copy(io.Discard, response.Body)
		return nil
	default:
		return decodeAPIError(response)
	}
}

func SubmitReturnTo(targetDate time.Time) string {
	date := targetDate.In(time.Local)
	return fmt.Sprintf("/time?day=%d&month=%d&year=%d", date.Day(), int(date.Month()), date.Year())
}

func (c *Client) submitBaseURL() string {
	baseURL := strings.TrimRight(c.BaseURL, "/")
	return strings.TrimSuffix(baseURL, "/api/v2")
}

func localTimezoneOffsetMinutes(value time.Time) int {
	_, offsetSeconds := value.Zone()
	return -(offsetSeconds / 60)
}
