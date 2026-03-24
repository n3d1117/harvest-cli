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

func (c *Client) PreviewSubmitWeek(ctx context.Context, targetDate time.Time) (SubmitResult, error) {
	plan, err := c.prepareSubmitWeek(ctx, targetDate)
	if err != nil {
		return SubmitResult{}, err
	}

	action := "would_submit"
	if plan.SubmittedBefore {
		action = "would_resubmit"
	}

	return SubmitResult{
		Action:          action,
		WeekStart:       plan.WeekStart.Format("2006-01-02"),
		WeekEnd:         plan.WeekEnd.Format("2006-01-02"),
		ReturnTo:        plan.ReturnTo,
		SubmittedBefore: plan.SubmittedBefore,
		SubmittedAfter:  plan.SubmittedBefore,
	}, nil
}

func (c *Client) SubmitWeek(ctx context.Context, targetDate, now time.Time) (SubmitResult, error) {
	plan, err := c.prepareSubmitWeek(ctx, targetDate)
	if err != nil {
		return SubmitResult{}, err
	}

	values := url.Values{}
	values.Set("authenticity_token", plan.CSRFToken)
	values.Set("return_to", plan.ReturnTo)
	values.Set("of_user", strconv.FormatInt(plan.OfUserID, 10))
	values.Set("submitted_date", strconv.Itoa(now.In(time.Local).YearDay()))
	values.Set("submitted_date_year", strconv.Itoa(now.In(time.Local).Year()))
	values.Set("period_begin", strconv.Itoa(plan.WeekStart.YearDay()))
	values.Set("period_begin_year", strconv.Itoa(plan.WeekStart.Year()))
	values.Set("from_timesheet_beta", "true")
	values.Set("from_screen", "daily")

	resultPage, err := c.postForm(ctx, strings.TrimRight(c.BaseURL, "/")+"/daily/review", values, strings.TrimRight(c.BaseURL, "/")+plan.ReturnTo)
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
	if plan.SubmittedBefore {
		action = "resubmitted"
	}

	return SubmitResult{
		Action:          action,
		WeekStart:       plan.WeekStart.Format("2006-01-02"),
		WeekEnd:         plan.WeekEnd.Format("2006-01-02"),
		ReturnTo:        resultPage.URL.Path,
		SubmittedBefore: plan.SubmittedBefore,
		SubmittedAfter:  resultPage.Timesheet.Submitted,
	}, nil
}

func (c *Client) prepareSubmitWeek(ctx context.Context, targetDate time.Time) (submitWeekPlan, error) {
	home, err := c.fetchTimesheetHome(ctx)
	if err != nil {
		return submitWeekPlan{}, err
	}
	if isSignInPage(home) {
		return submitWeekPlan{}, ErrUnauthenticated
	}
	if !home.PageData.ApprovalRequired {
		return submitWeekPlan{}, errors.New("timesheet approval is not enabled for this Harvest account")
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
		return submitWeekPlan{}, err
	}
	if isSignInPage(dayPage) {
		return submitWeekPlan{}, ErrUnauthenticated
	}
	if dayPage.CSRFToken == "" {
		return submitWeekPlan{}, errors.New("timesheet page did not include a CSRF token")
	}
	if dayPage.Timesheet.WeekOf == "" {
		return submitWeekPlan{}, errors.New("timesheet page did not include week data")
	}
	if dayPage.Timesheet.Approved {
		return submitWeekPlan{}, fmt.Errorf("week starting %s is already approved", dayPage.Timesheet.WeekOf)
	}

	weekStart, err := time.ParseInLocation("2006-01-02", dayPage.Timesheet.WeekOf, time.Local)
	if err != nil {
		return submitWeekPlan{}, fmt.Errorf("parse week start: %w", err)
	}

	return submitWeekPlan{
		CSRFToken:       dayPage.CSRFToken,
		OfUserID:        home.PageData.OfUserID,
		ReturnTo:        dayPage.URL.Path,
		SubmittedBefore: dayPage.Timesheet.Submitted,
		WeekStart:       weekStart,
		WeekEnd:         weekStart.AddDate(0, 0, 6),
	}, nil
}
