package config

import (
	"fmt"
	"os"
)

type GitHubConfig struct {
	Token      string
	Owner      string
	Repo       string
	BaseBranch string // Optional: for filtering PRs
}

func LoadGitHubConfig() (*GitHubConfig, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	owner := os.Getenv("GITHUB_OWNER")
	if owner == "" {
		return nil, fmt.Errorf("GITHUB_OWNER environment variable not set")
	}

	repo := os.Getenv("GITHUB_REPO")
	if repo == "" {
		return nil, fmt.Errorf("GITHUB_REPO environment variable not set")
	}

	return &GitHubConfig{
		Token: token,
		Owner: owner,
		Repo:  repo,
	}, nil
}
