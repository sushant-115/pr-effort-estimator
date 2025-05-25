package cmd_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/sushant-115/pr-effort-estimator/api/github"
	"github.com/sushant-115/pr-effort-estimator/internal/metrics"
	"github.com/sushant-115/pr-effort-estimator/pkg/config"

	gh "github.com/google/go-github/v63/github"
	// Import the main package as a separate package for testing
	// This requires moving main.go into its own directory or renaming it
	// For simplicity, we'll assume `main` package can be imported as `main_app`
	// If `main.go` is in the root, you'd just use `.` for import.
	// For this example, I'll assume you move main.go into a `cmd/pr-estimator/main.go` structure
	// or similar, making it importable. If not, the test will need to be in the same package.
	// For now, I'll keep it in `main_test` and assume it can access the `main` package's logic.
	// In a real project, main.go usually doesn't have direct unit tests, but integration tests.
	// For demonstration, I'll put it here.
	// Import the main package to run its init functions if any, or just to make sure dependencies are met.
)

// This mock server is similar to the one in github/client_test.go,
// but it's self-contained for the main test.
func setupMockGitHubServerForMain(t *testing.T) (*httptest.Server, func()) {
	mux := http.NewServeMux()

	mux.HandleFunc("/repos/test_owner/test_repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		prs := []*gh.PullRequest{
			{
				Number:    gh.Int(101),
				Title:     gh.String("Main Test PR 1"),
				State:     gh.String("closed"),
				CreatedAt: &gh.Timestamp{time.Now().Add(-96 * time.Hour)},
				User:      &gh.User{Login: gh.String("main_user1")},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(prs)
	})

	mux.HandleFunc("/repos/test_owner/test_repo/pulls/101", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		pr := &gh.PullRequest{
			Number:       gh.Int(101),
			Title:        gh.String("Main Test PR 1"),
			State:        gh.String("closed"),
			CreatedAt:    &gh.Timestamp{time.Now().Add(-96 * time.Hour)},
			MergedAt:     &gh.Timestamp{time.Now().Add(-48 * time.Hour)},
			Additions:    gh.Int(200),
			Deletions:    gh.Int(100),
			ChangedFiles: gh.Int(10),
			User:         &gh.User{Login: gh.String("main_user1")},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pr)
	})

	mux.HandleFunc("/repos/test_owner/test_repo/pulls/101/reviews", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		reviews := []*gh.PullRequestReview{
			{
				SubmittedAt: &gh.Timestamp{time.Now().Add(-72 * time.Hour)}, // First review for PR 101
				State:       gh.String("COMMENTED"),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(reviews)
	})

	server := httptest.NewServer(mux)
	return server, func() { server.Close() }
}

func TestMainAppIntegration(t *testing.T) {
	// Setup mock environment variables
	os.Setenv("GITHUB_TOKEN", "test_token")
	os.Setenv("GITHUB_OWNER", "test_owner")
	os.Setenv("GITHUB_REPO", "test_repo")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITHUB_OWNER")
		os.Unsetenv("GITHUB_REPO")
	}()

	// Setup mock GitHub server
	// server, cleanup := setupMockGitHubServerForMain(t)
	// defer cleanup()

	// Redirect go-github client to use our mock server
	// This is a global variable in go-github, so it affects all clients
	// oldBaseURL := gh.PRLink{}
	//gh.PullRequestsURL = server.URL + "/repos/%v/%v/pulls"
	//defer func() { gh.PullRequestsURL = oldBaseURL }()

	// For detailed PR/review calls, we need to override the BaseURL of the actual client.
	// This requires modifying the `github.NewClient` function or using reflection,
	// or, as done in client_test.go, adding a helper method to expose the internal client.
	// For this main integration test, we'll assume the `GhClient()` helper is available.
	// If not, this part would be trickier without direct modification to the main package's client creation.

	// Temporarily capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr) // Restore default output

	// Call the main function (assuming it's accessible or its logic is extracted)
	// Since main() is not exported, we can't directly call it from another package.
	// For a true integration test of `main`, you'd typically run the compiled binary
	// and assert its output/side effects.
	// For testing purposes within Go's test framework, we'll simulate its core logic.

	// Simulate the main logic:
	cfg, err := config.LoadGitHubConfig() // Assuming config.LoadGitHubConfig is accessible
	if err != nil {
		t.Fatalf("Error loading GitHub configuration in main test: %v", err)
	}

	ghClient := github.NewClient(cfg) // Assuming github.NewClient is accessible

	// Manually set the base URL for the underlying go-github client for detailed calls
	// if ghClientWithInternalClient, ok := ghClient.(interface{ GhClient() *gh.Client }); ok {
	// 	ghClientWithInternalClient.GhClient().BaseURL = server.URL + "/"
	// } else {
	// 	t.Log("Warning: Could not access internal ghClient to set BaseURL for detailed calls.")
	// }

	ctx := context.Background()
	prs, err := ghClient.GetPullRequests(ctx, "closed", 1) // Fetch 1 PR per page for simplicity
	if err != nil {
		t.Fatalf("Error fetching pull requests in main test: %v", err)
	}

	metrics.AnalyzePrs(prs) // Assuming metrics.AnalyzePrs is accessible

	// Assert on the captured log output
	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("Fetching closed pull requests for test_owner/test_repo...")) {
		t.Errorf("Expected log output to contain 'Fetching closed pull requests...', got:\n%s", output)
	}
	if !bytes.Contains(buf.Bytes(), []byte("PR #101: Main Test PR 1")) {
		t.Errorf("Expected log output to contain 'PR #101: Main Test PR 1', got:\n%s", output)
	}
	if !bytes.Contains(buf.Bytes(), []byte("Time to Merge: 48h0m0s")) {
		t.Errorf("Expected log output to contain 'Time to Merge: 48h0m0s', got:\n%s", output)
	}
	if !bytes.Contains(buf.Bytes(), []byte("Average Time to Merge (for 1 PRs): 48h0m0s")) {
		t.Errorf("Expected log output to contain 'Average Time to Merge (for 1 PRs): 48h0m0s', got:\n%s", output)
	}
}

// To make the main package's functions accessible for testing,
// you might need to extract the core logic from `main.go` into a separate
// exported function (e.g., `RunApp`) in a new package (e.g., `app`).
// For this example, I'm using an alias `_main` and assuming the functions are available,
// but in a real project, this would be structured differently.

// If `main.go` is in the root, and you want to test it directly, you'd put this
// test file in the root as `main_test.go` and use `.` for imports:
// import (
//     . "pr-estimator" // This imports the main package
// )
// However, this is generally discouraged for `main` packages.
