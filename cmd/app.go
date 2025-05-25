package cmd

import (
	"context"
	"log"

	"github.com/sushant-115/pr-effort-estimator/api/github"
	"github.com/sushant-115/pr-effort-estimator/internal/metrics"
	"github.com/sushant-115/pr-effort-estimator/pkg/config"
)

func Run() {
	cfg, err := config.LoadGitHubConfig()
	if err != nil {
		log.Fatalf("Error loading GitHub configuration: %v", err)
	}

	ghClient := github.NewClient(cfg)
	ctx := context.Background()

	// Fetch all closed pull requests for historical analysis
	log.Printf("Fetching closed pull requests for %s/%s...", cfg.Owner, cfg.Repo)
	prs, err := ghClient.GetPullRequests(ctx, "closed", 100) // Fetch 100 PRs per page
	if err != nil {
		log.Fatalf("Error fetching pull requests: %v", err)
	}

	metrics.AnalyzePrs(prs)

	// You could extend this to fetch "open" PRs and try to estimate their review time
	// based on historical data. This would involve more advanced statistical modeling.
}
