package commands

import (
	"fmt"
	"time"

	"harvest/internal/config"
	"harvest/internal/harvestapi"
	"harvest/internal/output"
)

func (a *App) runSubmitWeek(args []string) error {
	fs := newFlagSet("submit week", submitWeekHelp, a.Stdout, a.Stderr)
	dateFlag := fs.String("date", "", "Date in YYYY-MM-DD or `today`")
	dryRun := fs.Bool("dry-run", false, "Validate and preview the submit without sending it")
	jsonOutput := fs.Bool("json", false, "Print JSON")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return &exitError{Code: 2, Message: "submit week does not accept positional arguments"}
	}

	targetDate, err := resolveDateInput(*dateFlag, a.Now())
	if err != nil {
		return &exitError{Code: 2, Message: err.Error()}
	}
	date, err := time.ParseInLocation("2006-01-02", targetDate, time.Local)
	if err != nil {
		return err
	}

	client, _, err := a.client(config.Values{})
	if err != nil {
		return err
	}

	before, err := client.WeeklySummary(a.context(), date)
	if err != nil {
		return err
	}
	if before.Approved {
		return fmt.Errorf("week starting %s is already approved", before.StartDate)
	}

	result := harvestapi.SubmitResult{
		Action:          "would_submit",
		WeekStart:       before.StartDate,
		WeekEnd:         before.EndDate,
		ReturnTo:        harvestapi.SubmitReturnTo(date),
		SubmittedBefore: before.Submitted,
		SubmittedAfter:  before.Submitted,
	}
	if before.Submitted {
		result.Action = "would_resubmit"
	}

	if !*dryRun {
		user, err := client.Me(a.context())
		if err != nil {
			return err
		}
		weekStart, err := time.ParseInLocation("2006-01-02", before.StartDate, time.Local)
		if err != nil {
			return fmt.Errorf("parse week start: %w", err)
		}
		if err := client.SubmitWeekForApproval(a.context(), harvestapi.SubmitWeekInput{
			UserID:     user.ID,
			TargetDate: date,
			WeekStart:  weekStart,
		}); err != nil {
			return err
		}

		after, err := client.WeeklySummary(a.context(), date)
		if err != nil {
			return err
		}
		if !after.Submitted {
			return fmt.Errorf("Harvest did not mark the week as submitted")
		}

		result.Action = "submitted"
		if before.Submitted {
			result.Action = "resubmitted"
		}
		result.WeekStart = after.StartDate
		result.WeekEnd = after.EndDate
		result.SubmittedAfter = after.Submitted
	}

	if *jsonOutput {
		if *dryRun {
			return output.JSON(a.Stdout, struct {
				OK     bool `json:"ok"`
				DryRun bool `json:"dry_run"`
				Result any  `json:"result"`
			}{
				OK:     true,
				DryRun: true,
				Result: struct {
					Action          string `json:"action"`
					WeekStart       string `json:"week_start"`
					WeekEnd         string `json:"week_end"`
					ReturnTo        string `json:"return_to"`
					SubmittedBefore bool   `json:"submitted_before"`
				}{
					Action:          result.Action,
					WeekStart:       result.WeekStart,
					WeekEnd:         result.WeekEnd,
					ReturnTo:        result.ReturnTo,
					SubmittedBefore: result.SubmittedBefore,
				},
			})
		}

		return output.JSON(a.Stdout, struct {
			OK     bool                    `json:"ok"`
			Result harvestapi.SubmitResult `json:"result"`
		}{
			OK:     true,
			Result: result,
		})
	}

	verb := map[string]string{
		"submitted":      "Submitted",
		"resubmitted":    "Resubmitted",
		"would_submit":   "Dry run: would submit",
		"would_resubmit": "Dry run: would resubmit",
	}[result.Action]
	if verb == "" && *dryRun {
		verb = "Dry run: would submit"
	}
	if verb == "" {
		verb = "Submitted"
	}
	fmt.Fprintf(a.Stdout, "%s week %s to %s for approval.\n", verb, result.WeekStart, result.WeekEnd)
	return nil
}
