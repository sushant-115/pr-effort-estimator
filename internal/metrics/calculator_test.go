package metrics_test

import (
	"pr-effort-estimator/api/github"
	"pr-effort-estimator/internal/metrics"
	"testing"
	"time"
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

// TestAnalyzePrs is more of an integration test for logging and aggregation.
// We'll just ensure it doesn't panic and prints some output.
// For detailed metric calculation, individual CalculateMetrics tests are sufficient.
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
