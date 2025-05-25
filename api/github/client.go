package github

import (
	"context"
	"log"
	"time"

	gh "github.com/google/go-github/v63/github"
	"golang.org/x/oauth2"

	"github.com/sushant-115/pr-effort-estimator/pkg/config"
)

type Client struct {
	ghClient *gh.Client
	config   *config.GitHubConfig
}

func NewClient(cfg *config.GitHubConfig) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		ghClient: gh.NewClient(tc),
		config:   cfg,
	}
}

// GetPullRequests fetches a list of pull requests for the configured repository.
// It can be filtered by state (e.g., "closed", "all").
func (c *Client) GetPullRequests(ctx context.Context, state string, perPage int) ([]*PrData, error) {
	opts := &gh.PullRequestListOptions{
		State: state,
		ListOptions: gh.ListOptions{
			PerPage: perPage,
		},
	}

	var allPrs []*PrData
	for {
		prs, resp, err := c.ghClient.PullRequests.List(ctx, c.config.Owner, c.config.Repo, opts)
		if err != nil {
			return nil, err
		}

		for _, pr := range prs {
			// Fetch detailed PR to get additions/deletions/changed files
			detailedPR, _, err := c.ghClient.PullRequests.Get(ctx, c.config.Owner, c.config.Repo, pr.GetNumber())
			if err != nil {
				log.Printf("Warning: Could not fetch detailed PR #%d: %v", pr.GetNumber(), err)
				continue
			}

			// Fetch reviews to find the first review time
			reviews, _, err := c.ghClient.PullRequests.ListReviews(ctx, c.config.Owner, c.config.Repo, pr.GetNumber(), nil)
			var firstReviewedAt *time.Time
			if err == nil && len(reviews) > 0 {
				// Sort reviews by creation time to find the first
				earliestReviewTime := reviews[0].GetSubmittedAt()
				for _, review := range reviews {
					if review.GetSubmittedAt().Before(earliestReviewTime.Time) {
						earliestReviewTime = review.GetSubmittedAt()
					}
				}
				firstReviewedAt = &earliestReviewTime.Time
			} else if err != nil {
				log.Printf("Warning: Could not fetch reviews for PR #%d: %v", pr.GetNumber(), err)
			}
			mergedAt := detailedPR.GetMergedAt().Time
			closedAt := detailedPR.GetClosedAt().Time
			prData := &PrData{
				Number:          detailedPR.GetNumber(),
				Title:           detailedPR.GetTitle(),
				State:           detailedPR.GetState(),
				Author:          detailedPR.GetUser().GetLogin(),
				CreatedAt:       detailedPR.GetCreatedAt().Time,
				MergedAt:        &mergedAt,
				ClosedAt:        &closedAt,
				Additions:       detailedPR.GetAdditions(),
				Deletions:       detailedPR.GetDeletions(),
				ChangedFiles:    detailedPR.GetChangedFiles(),
				FirstReviewedAt: firstReviewedAt,
			}
			for _, label := range detailedPR.Labels {
				prData.Labels = append(prData.Labels, label.GetName())
			}
			allPrs = append(allPrs, prData)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allPrs, nil
}
