package github_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sushant-115/github.com/sushant-115/pr-effort-estimator/api/github"
	"github.com/sushant-115/github.com/sushant-115/pr-effort-estimator/pkg/config"

	gh "github.com/google/go-github/v63/github"
)

// setupMockGitHubServer creates a mock GitHub API server for testing.
// It returns the test server and a cleanup function.
func setupMockGitHubServer(t *testing.T) (*httptest.Server, func()) {
	mux := http.NewServeMux()

	// Mock endpoint for listing pull requests
	mux.HandleFunc("/repos/test_owner/test_repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Query().Get("state") != "closed" {
			http.Error(w, "Invalid state query parameter", http.StatusBadRequest)
			return
		}

		// Simulate a list of PRs
		prs := []*gh.PullRequest{
			{
				Number:    gh.Int(1),
				Title:     gh.String("Test PR 1"),
				State:     gh.String("closed"),
				CreatedAt: &gh.Timestamp{time.Now().Add(-48 * time.Hour)},
				User:      &gh.User{Login: gh.String("user1")},
			},
			{
				Number:    gh.Int(2),
				Title:     gh.String("Test PR 2"),
				State:     gh.String("closed"),
				CreatedAt: &gh.Timestamp{time.Now().Add(-72 * time.Hour)},
				User:      &gh.User{Login: gh.String("user2")},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(prs)
	})

	// Mock endpoint for getting a specific pull request
	mux.HandleFunc("/repos/test_owner/test_repo/pulls/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		prNum := r.URL.Path[len("/repos/test_owner/test_repo/pulls/"):]
		switch prNum {
		case "1":
			pr := &gh.PullRequest{
				Number:       gh.Int(1),
				Title:        gh.String("Test PR 1"),
				State:        gh.String("closed"),
				CreatedAt:    &gh.Timestamp{time.Now().Add(-48 * time.Hour)},
				MergedAt:     &gh.Timestamp{time.Now().Add(-24 * time.Hour)},
				Additions:    gh.Int(100),
				Deletions:    gh.Int(50),
				ChangedFiles: gh.Int(5),
				User:         &gh.User{Login: gh.String("user1")},
				Labels: []*gh.Label{
					{Name: gh.String("bug")},
					{Name: gh.String("feature")},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(pr)
		case "2":
			pr := &gh.PullRequest{
				Number:       gh.Int(2),
				Title:        gh.String("Test PR 2"),
				State:        gh.String("closed"),
				CreatedAt:    &gh.Timestamp{(time.Now().Add(-72 * time.Hour))},
				ClosedAt:     &gh.Timestamp{time.Now().Add(-12 * time.Hour)}, // Closed but not merged
				Additions:    gh.Int(20),
				Deletions:    gh.Int(10),
				ChangedFiles: gh.Int(2),
				User:         &gh.User{Login: gh.String("user2")},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(pr)
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	})

	// Mock endpoint for listing pull request reviews
	mux.HandleFunc("/repos/test_owner/test_repo/pulls/1/reviews", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		reviews := []*gh.PullRequestReview{
			{
				SubmittedAt: &gh.Timestamp{time.Now().Add(-40 * time.Hour)}, // First review for PR}1
				State:       gh.String("COMMENTED"),
			},
			{
				SubmittedAt: &gh.Timestamp{time.Now().Add(-30 * time.Hour)},
				State:       gh.String("APPROVED"),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(reviews)
	})

	mux.HandleFunc("/repos/test_owner/test_repo/pulls/2/reviews", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Simulate no reviews for PR 2
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, "[]")
	})

	server := httptest.NewServer(mux)
	return server, func() { server.Close() }
}

func TestGetPullRequests(t *testing.T) {
	// server, cleanup := setupMockGitHubServer(t)
	// defer cleanup()

	// Create a client with mock config
	cfg := &config.GitHubConfig{
		Token: "dummy_token", // Token is not used by mock server, but required by client
		Owner: "test_owner",
		Repo:  "test_repo",
	}
	client := github.NewClient(cfg)

	// Manually set the base URL for the underlying go-github client
	// This is crucial because go-github client uses its own internal http client.
	// We need to ensure the client used by the github.Client points to our mock server.
	//client.GhClient().BaseURL = server.URL + "/"

	ctx := context.Background()
	prs, err := client.GetPullRequests(ctx, "closed", 10)
	if err != nil {
		t.Fatalf("GetPullRequests failed: %v", err)
	}

	if len(prs) != 2 {
		t.Fatalf("Expected 2 pull requests, got %d", len(prs))
	}

	// Verify PR 1 data
	pr1 := prs[0]
	if pr1.Number != 1 {
		t.Errorf("Expected PR number 1, got %d", pr1.Number)
	}
	if pr1.Title != "Test PR 1" {
		t.Errorf("Expected PR title 'Test PR 1', got '%s'", pr1.Title)
	}
	if pr1.State != "closed" {
		t.Errorf("Expected PR state 'closed', got '%s'", pr1.State)
	}
	if pr1.Author != "user1" {
		t.Errorf("Expected PR author 'user1', got '%s'", pr1.Author)
	}
	if pr1.MergedAt == nil {
		t.Errorf("Expected PR 1 to be merged, but MergedAt is nil")
	}
	if pr1.Additions != 100 {
		t.Errorf("Expected PR 1 additions 100, got %d", pr1.Additions)
	}
	if pr1.Deletions != 50 {
		t.Errorf("Expected PR 1 deletions 50, got %d", pr1.Deletions)
	}
	if pr1.ChangedFiles != 5 {
		t.Errorf("Expected PR 1 changed files 5, got %d", pr1.ChangedFiles)
	}
	if pr1.FirstReviewedAt == nil {
		t.Errorf("Expected PR 1 to have a first review time, but it's nil")
	}
	if len(pr1.Labels) != 2 || pr1.Labels[0] != "bug" || pr1.Labels[1] != "feature" {
		t.Errorf("Expected PR 1 labels [bug feature], got %v", pr1.Labels)
	}

	// Verify PR 2 data
	pr2 := prs[1]
	if pr2.Number != 2 {
		t.Errorf("Expected PR number 2, got %d", pr2.Number)
	}
	if pr2.MergedAt != nil {
		t.Errorf("Expected PR 2 not to be merged, but MergedAt is not nil")
	}
	if pr2.ClosedAt == nil {
		t.Errorf("Expected PR 2 to be closed, but ClosedAt is nil")
	}
	if pr2.FirstReviewedAt != nil {
		t.Errorf("Expected PR 2 not to have a first review time, but it's %v", pr2.FirstReviewedAt)
	}
}

// Helper to expose the internal ghClient for testing purposes
//

// NOTE: Add this method to your github/client.go file
// func (c *Client) GhClient() *gh.Client {
// 	return c.ghClient
// }
