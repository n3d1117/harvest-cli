package websubmit

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	titlePattern     = regexp.MustCompile(`(?is)<title>([^<]+)</title>`)
	metaTokenPattern = regexp.MustCompile(`name="csrf-token"\s+content="([^"]+)"`)
	formTokenPattern = regexp.MustCompile(`name="authenticity_token"\s+value="([^"]+)"`)
)

type pageData struct {
	OfUserID    int64 `json:"of_user_id"`
	CurrentUser struct {
		ID       int64  `json:"id"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
	} `json:"current_user"`
	ApprovalRequired bool `json:"approval_required"`
}

type timesheetData struct {
	WeekOf    string `json:"week_of"`
	Submitted bool   `json:"submitted"`
	Approved  bool   `json:"approved"`
}

type page struct {
	Title     string
	URL       *url.URL
	HTML      string
	CSRFToken string
	PageData  pageData
	Timesheet timesheetData
}

func parsePage(body []byte, requestURL *url.URL) (page, error) {
	html := string(body)
	result := page{
		Title:     strings.TrimSpace(firstMatch(titlePattern, html)),
		URL:       requestURL,
		HTML:      html,
		CSRFToken: firstMatch(metaTokenPattern, html),
	}

	if raw := extractJSONIsland(html, "page_data-data-island"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &result.PageData); err != nil {
			return page{}, fmt.Errorf("decode page data: %w", err)
		}
	}
	if raw := extractJSONIsland(html, "timesheet-data-island"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &result.Timesheet); err != nil {
			return page{}, fmt.Errorf("decode timesheet data: %w", err)
		}
	}

	return result, nil
}

func extractJSONIsland(html, id string) string {
	startMarker := `<script id="` + id + `" type="application/json">`
	start := strings.Index(html, startMarker)
	if start == -1 {
		return ""
	}
	start += len(startMarker)
	end := strings.Index(html[start:], "</script>")
	if end == -1 {
		return ""
	}
	return html[start : start+end]
}

func firstMatch(pattern *regexp.Regexp, input string) string {
	matches := pattern.FindStringSubmatch(input)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

func isSignInPage(page page) bool {
	if strings.EqualFold(strings.TrimSpace(page.Title), "Sign in - Harvest") {
		return true
	}
	return strings.Contains(page.HTML, `<form action="/sessions"`)
}
