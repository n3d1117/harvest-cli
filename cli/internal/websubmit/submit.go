package websubmit

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (c *Client) SubmitWeek(ctx context.Context, targetDate, now time.Time) (SubmitResult, error) {
	home, err := c.fetchTimesheetHome(ctx)
	if err != nil {
		return SubmitResult{}, err
	}
	if isSignInPage(home) {
		return SubmitResult{}, ErrUnauthenticated
	}
	if !home.PageData.ApprovalRequired {
		return SubmitResult{}, errors.New("timesheet approval is not enabled for this Harvest account")
	}

	dayURL := fmt.Sprintf("%s/time/day/%d/%d/%d/%d",
		strings.TrimRight(c.BaseURL, "/"),
		targetDate.Year(),
		int(targetDate.Month()),
		targetDate.Day(),
		home.PageData.OfUserID,
	)

	dayPage, err := c.getPage(ctx, dayURL)
	if err != nil {
		return SubmitResult{}, err
	}
	if isSignInPage(dayPage) {
		return SubmitResult{}, ErrUnauthenticated
	}
	if dayPage.CSRFToken == "" {
		return SubmitResult{}, errors.New("timesheet page did not include a CSRF token")
	}
	if dayPage.Timesheet.WeekOf == "" {
		return SubmitResult{}, errors.New("timesheet page did not include week data")
	}
	if dayPage.Timesheet.Approved {
		return SubmitResult{}, fmt.Errorf("week starting %s is already approved", dayPage.Timesheet.WeekOf)
	}

	weekStart, err := time.ParseInLocation("2006-01-02", dayPage.Timesheet.WeekOf, time.Local)
	if err != nil {
		return SubmitResult{}, fmt.Errorf("parse week start: %w", err)
	}
	weekEnd := weekStart.AddDate(0, 0, 6)

	values := url.Values{}
	values.Set("authenticity_token", dayPage.CSRFToken)
	values.Set("return_to", dayPage.URL.Path)
	values.Set("of_user", strconv.FormatInt(home.PageData.OfUserID, 10))
	values.Set("submitted_date", strconv.Itoa(now.In(time.Local).YearDay()))
	values.Set("submitted_date_year", strconv.Itoa(now.In(time.Local).Year()))
	values.Set("period_begin", strconv.Itoa(weekStart.YearDay()))
	values.Set("period_begin_year", strconv.Itoa(weekStart.Year()))
	values.Set("from_timesheet_beta", "true")
	values.Set("from_screen", "daily")

	resultPage, err := c.postForm(ctx, strings.TrimRight(c.BaseURL, "/")+"/daily/review", values, dayPage.URL.String())
	if err != nil {
		return SubmitResult{}, err
	}
	if isSignInPage(resultPage) {
		return SubmitResult{}, ErrUnauthenticated
	}
	if !resultPage.Timesheet.Submitted {
		return SubmitResult{}, errors.New("Harvest did not mark the week as submitted")
	}

	action := "submitted"
	if dayPage.Timesheet.Submitted {
		action = "resubmitted"
	}

	return SubmitResult{
		Action:          action,
		WeekStart:       weekStart.Format("2006-01-02"),
		WeekEnd:         weekEnd.Format("2006-01-02"),
		ReturnTo:        resultPage.URL.Path,
		SubmittedBefore: dayPage.Timesheet.Submitted,
		SubmittedAfter:  resultPage.Timesheet.Submitted,
	}, nil
}
