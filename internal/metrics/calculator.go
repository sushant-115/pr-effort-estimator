package metrics

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/sushant-115/pr-effort-estimator/api/github"

	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"
)

// PrMetrics holds calculated metrics for a single Pull Request.
type PrMetrics struct {
	Number            int
	Title             string
	TimeToFirstReview time.Duration
	TimeToMerge       time.Duration
	ReviewToMerge     time.Duration
	Additions         int
	Deletions         int
	ChangedFiles      int
	State             string
}

// NormalDistributionEstimates holds percentile estimates for a given metric.
type NormalDistributionEstimates struct {
	Mean        time.Duration
	StdDev      time.Duration
	P50         time.Duration // 50th percentile (median)
	P80         time.Duration // 80th percentile
	P90         time.Duration // 90th percentile
	P95         time.Duration // 95th percentile
	SampleCount int
}

// CalculateMetrics computes various time-based metrics for a single PR.
func CalculateMetrics(pr *github.PrData) *PrMetrics {
	metrics := &PrMetrics{
		Number:       pr.Number,
		Title:        pr.Title,
		Additions:    pr.Additions,
		Deletions:    pr.Deletions,
		ChangedFiles: pr.ChangedFiles,
		State:        pr.State,
	}

	// Calculate TimeToFirstReview
	if pr.FirstReviewedAt != nil {
		metrics.TimeToFirstReview = pr.FirstReviewedAt.Sub(pr.CreatedAt)
	}

	// Calculate TimeToMerge or TimeToClose
	if pr.MergedAt != nil {
		metrics.TimeToMerge = pr.MergedAt.Sub(pr.CreatedAt)
		if pr.FirstReviewedAt != nil {
			metrics.ReviewToMerge = pr.MergedAt.Sub(*pr.FirstReviewedAt)
		}
	} else if pr.ClosedAt != nil && pr.State == "closed" {
		// If closed but not merged, TimeToMerge becomes TimeToClose
		metrics.TimeToMerge = pr.ClosedAt.Sub(pr.CreatedAt)
		// ReviewToMerge is not applicable for unmerged PRs, keep as 0
	}

	return metrics
}

// AnalyzePrs iterates through a slice of PrData, calculates metrics, and prints them.
func AnalyzePrs(prs []*github.PrData) {
	var allMetrics []*PrMetrics
	var totalTimeToMerge time.Duration
	var mergedPrCount int

	log.Println("\n--- Individual PR Analysis ---")
	for _, pr := range prs {
		metrics := CalculateMetrics(pr)
		allMetrics = append(allMetrics, metrics)

		log.Printf("PR #%d: %s (State: %s)", metrics.Number, metrics.Title, metrics.State)
		if metrics.TimeToFirstReview > 0 {
			log.Printf("  Time to First Review: %v", metrics.TimeToFirstReview)
		} else {
			log.Println("  Time to First Review: N/A (No reviews or PR still open)")
		}

		if metrics.State == "merged" {
			log.Printf("  Time to Merge: %v", metrics.TimeToMerge)
			if metrics.ReviewToMerge > 0 {
				log.Printf("  Review to Merge: %v", metrics.ReviewToMerge)
			}
			log.Printf("  Size: +%d / -%d, Files: %d", metrics.Additions, metrics.Deletions, metrics.ChangedFiles)
			totalTimeToMerge += metrics.TimeToMerge
			mergedPrCount++
		} else if metrics.State == "closed" {
			log.Printf("  Time to Close (unmerged): %v", metrics.TimeToMerge)
			log.Printf("  Size: +%d / -%d, Files: %d", metrics.Additions, metrics.Deletions, metrics.ChangedFiles)
		} else { // open
			log.Printf("  Current Age: %v", time.Since(pr.CreatedAt))
			log.Printf("  Size: +%d / -%d, Files: %d", metrics.Additions, metrics.Deletions, metrics.ChangedFiles)
		}
		log.Println("---")
	}

	log.Println("\n--- Aggregated Metrics (Simple Average) ---")
	if mergedPrCount > 0 {
		avgTimeToMerge := totalTimeToMerge / time.Duration(mergedPrCount)
		log.Printf("Average Time to Merge (for %d merged PRs): %v\n", mergedPrCount, avgTimeToMerge)
	} else {
		log.Println("No merged PRs to calculate average time to merge.")
	}

	log.Println("\n--- Normal Distribution Based Estimates ---")
	estimateTimeToFirstReview := EstimateTimesUsingNormalDistribution(allMetrics, func(m *PrMetrics) time.Duration {
		return m.TimeToFirstReview
	}, "Time to First Review")

	estimateTimeToMerge := EstimateTimesUsingNormalDistribution(allMetrics, func(m *PrMetrics) time.Duration {
		return m.TimeToMerge
	}, "Time to Merge (Merged PRs)")

	if estimateTimeToFirstReview.SampleCount > 0 {
		fmt.Printf("Estimated Time to First Review (based on %d PRs):\n", estimateTimeToFirstReview.SampleCount)
		fmt.Printf("  Mean: %v, StdDev: %v\n", estimateTimeToFirstReview.Mean, estimateTimeToFirstReview.StdDev)
		fmt.Printf("  50th Percentile (Median): %v\n", estimateTimeToFirstReview.P50)
		fmt.Printf("  80th Percentile: %v\n", estimateTimeToFirstReview.P80)
		fmt.Printf("  90th Percentile: %v\n", estimateTimeToFirstReview.P90)
		fmt.Printf("  95th Percentile: %v\n", estimateTimeToFirstReview.P95)
	} else {
		log.Println("Not enough data to estimate Time to First Review using normal distribution.")
	}

	if estimateTimeToMerge.SampleCount > 0 {
		fmt.Printf("\nEstimated Time to Merge (based on %d merged PRs):\n", estimateTimeToMerge.SampleCount)
		fmt.Printf("  Mean: %v, StdDev: %v\n", estimateTimeToMerge.Mean, estimateTimeToMerge.StdDev)
		fmt.Printf("  50th Percentile (Median): %v\n", estimateTimeToMerge.P50)
		fmt.Printf("  80th Percentile: %v\n", estimateTimeToMerge.P80)
		fmt.Printf("  90th Percentile: %v\n", estimateTimeToMerge.P90)
		fmt.Printf("  95th Percentile: %v\n", estimateTimeToMerge.P95)
	} else {
		log.Println("Not enough data to estimate Time to Merge using normal distribution.")
	}
}

// EstimateTimesUsingNormalDistribution calculates normal distribution-based estimates
// for a given time metric from a slice of PrMetrics.
// It takes a selector function to pick the duration from each PrMetrics object.
func EstimateTimesUsingNormalDistribution(metrics []*PrMetrics, selector func(*PrMetrics) time.Duration, metricName string) NormalDistributionEstimates {
	durations := []float64{}
	for _, m := range metrics {
		val := selector(m)
		if val > 0 { // Only include valid, non-zero durations for calculation
			// Convert duration to hours (or any consistent unit) for stat calculations
			durations = append(durations, val.Hours())
		}
	}

	if len(durations) < 2 { // Need at least 2 data points for std dev
		log.Printf("Warning: Not enough data points (%d) to calculate %s normal distribution. Skipping.", len(durations), metricName)
		return NormalDistributionEstimates{}
	}

	mean := stat.Mean(durations, nil)
	stdDev := stat.StdDev(durations, nil)

	// Create a normal distribution
	norm := distuv.Normal{
		Mu:    mean,
		Sigma: stdDev,
	}

	estimates := NormalDistributionEstimates{
		Mean:        time.Duration(mean * float64(time.Hour)),
		StdDev:      time.Duration(stdDev * float64(time.Hour)),
		SampleCount: len(durations),
	}

	// Calculate percentiles using the Quantile (inverse CDF) function
	estimates.P50 = time.Duration(norm.Quantile(0.50) * float64(time.Hour))
	estimates.P80 = time.Duration(norm.Quantile(0.80) * float64(time.Hour))
	estimates.P90 = time.Duration(norm.Quantile(0.90) * float64(time.Hour))
	estimates.P95 = time.Duration(norm.Quantile(0.95) * float64(time.Hour))

	// Handle potential NaNs if stdDev is 0 (only one data point or all same)
	if math.IsNaN(estimates.P50.Hours()) {
		estimates.P50 = estimates.Mean
		estimates.P80 = estimates.Mean
		estimates.P90 = estimates.Mean
		estimates.P95 = estimates.Mean
	}

	return estimates
}
