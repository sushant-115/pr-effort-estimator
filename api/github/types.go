package github

import "time"

// PrData represents simplified pull request information
type PrData struct {
	Number          int
	Title           string
	State           string // e.g., "open", "closed", "merged"
	Author          string
	CreatedAt       time.Time
	MergedAt        *time.Time // Pointer as it can be nil if not merged
	ClosedAt        *time.Time // Pointer as it can be nil if not closed
	Additions       int
	Deletions       int
	ChangedFiles    int
	FirstReviewedAt *time.Time // Timestamp of the first review
	Labels          []string   // Labels applied to the PR
}
