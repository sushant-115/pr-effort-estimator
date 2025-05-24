package metrics

import (
	"log"
	"pr-effort-estimator/api/github"
	"time"
)

// PrMetrics holds calculated metrics for a single PR
type PrMetrics struct {
	PR                *github.PrData
	TimeToFirstReview time.Duration
	TimeToMerge       time.Duration
	ReviewToMerge     time.Duration
}

// CalculateMetrics computes various time durations for a given PR.
func CalculateMetrics(pr *github.PrData) *PrMetrics {
	metrics := &PrMetrics{
		PR: pr,
	}

	if pr.FirstReviewedAt != nil {
		metrics.TimeToFirstReview = pr.FirstReviewedAt.Sub(pr.CreatedAt)
	}

	if pr.MergedAt != nil {
		metrics.TimeToMerge = pr.MergedAt.Sub(pr.CreatedAt)
		if pr.FirstReviewedAt != nil {
			metrics.ReviewToMerge = pr.MergedAt.Sub(*pr.FirstReviewedAt)
		}
	} else if pr.State == "closed" && pr.ClosedAt != nil {
		// If not merged but closed, we can still calculate time to close
		// This might indicate PRs that were abandoned or rejected.
		metrics.TimeToMerge = pr.ClosedAt.Sub(pr.CreatedAt) // Or a different metric like TimeToClose
	}

	return metrics
}

// AnalyzePrs calculates metrics for a slice of PRs and provides some aggregated stats.
func AnalyzePrs(prs []*github.PrData) {
	var totalTimeToFirstReview time.Duration
	var totalTimeToMerge time.Duration
	var reviewablePRs int // PRs that received a review
	var mergablePRs int   // PRs that were merged

	log.Printf("Analyzing %d pull requests...", len(prs))

	for _, pr := range prs {
		metrics := CalculateMetrics(pr)

		log.Printf("PR #%d: %s", pr.Number, pr.Title)
		log.Printf("  Size: +%d / -%d lines, %d files changed", pr.Additions, pr.Deletions, pr.ChangedFiles)
		log.Printf("  Created: %s", pr.CreatedAt.Format(time.RFC822))
		if metrics.TimeToFirstReview > 0 {
			log.Printf("  Time to First Review: %s", metrics.TimeToFirstReview)
			totalTimeToFirstReview += metrics.TimeToFirstReview
			reviewablePRs++
		} else {
			log.Printf("  Time to First Review: N/A (no reviews yet or found)")
		}

		if pr.MergedAt != nil {
			log.Printf("  Merged: %s", pr.MergedAt.Format(time.RFC822))
			log.Printf("  Time to Merge: %s", metrics.TimeToMerge)
			totalTimeToMerge += metrics.TimeToMerge
			mergablePRs++
		} else if pr.ClosedAt != nil {
			log.Printf("  Closed (not merged): %s (Time to close: %s)", pr.ClosedAt.Format(time.RFC822), pr.ClosedAt.Sub(pr.CreatedAt))
		} else {
			log.Printf("  Current State: %s", pr.State)
		}
		log.Printf("------------------------------------------")
	}

	log.Printf("\n--- Aggregated Statistics ---")
	if reviewablePRs > 0 {
		log.Printf("Average Time to First Review (for %d PRs): %s", reviewablePRs, totalTimeToFirstReview/time.Duration(reviewablePRs))
	} else {
		log.Printf("No PRs with reviews found to calculate Average Time to First Review.")
	}

	if mergablePRs > 0 {
		log.Printf("Average Time to Merge (for %d PRs): %s", mergablePRs, totalTimeToMerge/time.Duration(mergablePRs))
	} else {
		log.Printf("No merged PRs found to calculate Average Time to Merge.")
	}
}
