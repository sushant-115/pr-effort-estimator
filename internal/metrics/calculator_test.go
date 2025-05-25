package metrics_test

import (
	"math"
	"testing"
	"time"

	"github.com/sushant-115/pr-effort-estimator/api/github"
	"github.com/sushant-115/pr-effort-estimator/internal/metrics"
)

func TestCalculateMetrics_MergedPRWithReview(t *testing.T) {
	now := time.Now()
	pr := &github.PrData{
		Number:          1,
		CreatedAt:       now.Add(-48 * time.Hour),
		FirstReviewedAt: &[]time.Time{now.Add(-36 * time.Hour)}[0], // 12 hours after creation
		MergedAt:        &[]time.Time{now.Add(-24 * time.Hour)}[0], // 24 hours after first review
		State:           "merged",
		Additions:       100,
		Deletions:       50,
		ChangedFiles:    5,
	}

	m := metrics.CalculateMetrics(pr)

	expectedTimeToFirstReview := 12 * time.Hour
	expectedTimeToMerge := 24 * time.Hour
	expectedReviewToMerge := 12 * time.Hour

	if m.TimeToFirstReview != expectedTimeToFirstReview {
		t.Errorf("Expected TimeToFirstReview %v, got %v", expectedTimeToFirstReview, m.TimeToFirstReview)
	}
	if m.TimeToMerge != expectedTimeToMerge {
		t.Errorf("Expected TimeToMerge %v, got %v", expectedTimeToMerge, m.TimeToMerge)
	}
	if m.ReviewToMerge != expectedReviewToMerge {
		t.Errorf("Expected ReviewToMerge %v, got %v", expectedReviewToMerge, m.ReviewToMerge)
	}
}

func TestCalculateMetrics_ClosedPRNoMerge(t *testing.T) {
	now := time.Now()
	pr := &github.PrData{
		Number:          2,
		CreatedAt:       now.Add(-72 * time.Hour),
		FirstReviewedAt: &[]time.Time{now.Add(-60 * time.Hour)}[0],
		ClosedAt:        &[]time.Time{now.Add(-24 * time.Hour)}[0], // Closed 48 hours after first review
		State:           "closed",
		Additions:       20,
		Deletions:       10,
		ChangedFiles:    2,
	}

	m := metrics.CalculateMetrics(pr)

	expectedTimeToFirstReview := 12 * time.Hour
	expectedTimeToMerge := 48 * time.Hour  // TimeToMerge is actually TimeToClose in this case
	expectedReviewToMerge := 0 * time.Hour // Should be 0 as it's not merged

	if m.TimeToFirstReview != expectedTimeToFirstReview {
		t.Errorf("Expected TimeToFirstReview %v, got %v", expectedTimeToFirstReview, m.TimeToFirstReview)
	}
	if m.TimeToMerge != expectedTimeToMerge {
		t.Errorf("Expected TimeToMerge (TimeToClose) %v, got %v", expectedTimeToMerge, m.TimeToMerge)
	}
	if m.ReviewToMerge != expectedReviewToMerge {
		t.Errorf("Expected ReviewToMerge %v for non-merged PR, got %v", expectedReviewToMerge, m.ReviewToMerge)
	}
}

func TestCalculateMetrics_OpenPRNoReviewNoMerge(t *testing.T) {
	now := time.Now()
	pr := &github.PrData{
		Number:    3,
		CreatedAt: now.Add(-24 * time.Hour),
		State:     "open",
		Additions: 50,
		Deletions: 5,
	}

	m := metrics.CalculateMetrics(pr)

	if m.TimeToFirstReview != 0 {
		t.Errorf("Expected TimeToFirstReview 0, got %v", m.TimeToFirstReview)
	}
	if m.TimeToMerge != 0 {
		t.Errorf("Expected TimeToMerge 0, got %v", m.TimeToMerge)
	}
	if m.ReviewToMerge != 0 {
		t.Errorf("Expected ReviewToMerge 0, got %v", m.ReviewToMerge)
	}
}

func TestCalculateMetrics_MergedPRNoReview(t *testing.T) {
	now := time.Now()
	pr := &github.PrData{
		Number:    4,
		CreatedAt: now.Add(-48 * time.Hour),
		MergedAt:  &[]time.Time{now.Add(-24 * time.Hour)}[0],
		State:     "merged",
		Additions: 10,
		Deletions: 2,
	}

	m := metrics.CalculateMetrics(pr)

	expectedTimeToMerge := 24 * time.Hour

	if m.TimeToFirstReview != 0 {
		t.Errorf("Expected TimeToFirstReview 0, got %v", m.TimeToFirstReview)
	}
	if m.TimeToMerge != expectedTimeToMerge {
		t.Errorf("Expected TimeToMerge %v, got %v", expectedTimeToMerge, m.TimeToMerge)
	}
	if m.ReviewToMerge != 0 {
		t.Errorf("Expected ReviewToMerge 0, got %v", m.ReviewToMerge)
	}
}

func TestAnalyzePrs(t *testing.T) {
	now := time.Now()
	prs := []*github.PrData{
		{
			Number:          1,
			Title:           "PR 1",
			CreatedAt:       now.Add(-48 * time.Hour),
			FirstReviewedAt: &[]time.Time{now.Add(-36 * time.Hour)}[0],
			MergedAt:        &[]time.Time{now.Add(-24 * time.Hour)}[0],
			State:           "merged",
			Additions:       100,
			Deletions:       50,
			ChangedFiles:    5,
		},
		{
			Number:    2,
			Title:     "PR 2",
			CreatedAt: now.Add(-24 * time.Hour),
			State:     "open",
			Additions: 10,
			Deletions: 2,
		},
	}

	// This test primarily checks if the function runs without panicking
	// and produces some output. Capturing log output for assertion is
	// more complex and often not necessary for simple logging functions.
	metrics.AnalyzePrs(prs)
}

func TestEstimateTimesUsingNormalDistribution(t *testing.T) {
	// Sample data in hours
	// 24h, 48h, 36h, 60h, 30h
	prMetrics := []*metrics.PrMetrics{
		{Number: 1, TimeToMerge: 24 * time.Hour, State: "merged"},
		{Number: 2, TimeToMerge: 48 * time.Hour, State: "merged"},
		{Number: 3, TimeToMerge: 36 * time.Hour, State: "merged"},
		{Number: 4, TimeToMerge: 60 * time.Hour, State: "merged"},
		{Number: 5, TimeToMerge: 30 * time.Hour, State: "merged"},
		{Number: 6, TimeToMerge: 0 * time.Hour, State: "open"}, // Should be ignored
	}

	// Selector for TimeToMerge
	selector := func(m *metrics.PrMetrics) time.Duration {
		return m.TimeToMerge
	}

	estimates := metrics.EstimateTimesUsingNormalDistribution(prMetrics, selector, "Test Metric")

	// Convert expected values to hours for easier comparison
	expectedMeanHours := (24.0 + 48.0 + 36.0 + 60.0 + 30.0) / 5.0 // 39.6
	// Expected standard deviation needs to be calculated by hand or a tool for precise comparison
	// For a simple test, we can check if it's within a reasonable range.
	// Using a calculator for StDev of [24, 48, 36, 60, 30] is approx 13.91
	expectedStdDevHours := 13.91 // Approximate

	// Quantiles:
	// For a normal distribution with mean 39.6 and stddev 13.91:
	// P50 (Median) is close to Mean
	// P80, P90, P95 will be higher than the mean

	// Allow for small floating point deviations
	tolerance := 1 * time.Minute

	if estimates.SampleCount != 5 {
		t.Errorf("Expected sample count 5, got %d", estimates.SampleCount)
	}
	if math.Abs(estimates.Mean.Hours()-expectedMeanHours) > tolerance.Hours() {
		t.Errorf("Expected Mean %v, got %v", time.Duration(expectedMeanHours*float64(time.Hour)), estimates.Mean)
	}
	// For StdDev and Percentiles, check if they are within a reasonable range
	// As exact values depend on `gonum`'s implementation, and floating point math.
	if estimates.StdDev < 10*time.Hour || estimates.StdDev > 20*time.Hour { // Rough range
		t.Errorf("Expected StdDev to be around 13.91h, got %v", estimates.StdDev)
	}
	if estimates.P50 < 35*time.Hour || estimates.P50 > 45*time.Hour { // Roughly around the mean
		t.Errorf("Expected P50 to be around 39.6h, got %v", estimates.P50)
	}
	if estimates.P80 < 50*time.Hour || estimates.P80 > 60*time.Hour { // Should be > mean
		t.Errorf("Expected P80 to be greater than mean, got %v", estimates.P80)
	}
	if estimates.P90 < 55*time.Hour || estimates.P90 > 70*time.Hour { // Should be > P80
		t.Errorf("Expected P90 to be greater than P80, got %v", estimates.P90)
	}
	if estimates.P95 < 60*time.Hour || estimates.P95 > 80*time.Hour { // Should be > P90
		t.Errorf("Expected P95 to be greater than P90, got %v", estimates.P95)
	}
}

func TestEstimateTimesUsingNormalDistribution_InsufficientData(t *testing.T) {
	prMetrics := []*metrics.PrMetrics{
		{Number: 1, TimeToMerge: 24 * time.Hour, State: "merged"},
	}
	selector := func(m *metrics.PrMetrics) time.Duration {
		return m.TimeToMerge
	}

	estimates := metrics.EstimateTimesUsingNormalDistribution(prMetrics, selector, "Insufficient Data Test")
	if estimates.SampleCount != 0 {
		t.Errorf("Expected 0 sample count for insufficient data, got %d", estimates.SampleCount)
	}
	// Other fields should also be zero-valued
	if estimates.Mean != 0 || estimates.StdDev != 0 || estimates.P50 != 0 {
		t.Errorf("Expected all estimates to be zero-valued for insufficient data, got %+v", estimates)
	}
}

func TestEstimateTimesUsingNormalDistribution_NoValidData(t *testing.T) {
	prMetrics := []*metrics.PrMetrics{
		{Number: 1, TimeToMerge: 0 * time.Hour, State: "open"},
		{Number: 2, TimeToMerge: 0 * time.Hour, State: "open"},
	}
	selector := func(m *metrics.PrMetrics) time.Duration {
		return m.TimeToMerge
	}

	estimates := metrics.EstimateTimesUsingNormalDistribution(prMetrics, selector, "No Valid Data Test")
	if estimates.SampleCount != 0 {
		t.Errorf("Expected 0 sample count for no valid data, got %d", estimates.SampleCount)
	}
}
