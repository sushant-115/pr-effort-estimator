package config_test

import (
	"os"
	"pr-effort-estimator/pkg/config"
	"testing"
)

func TestLoadGitHubConfig_Success(t *testing.T) {
	// Set up mock environment variables
	os.Setenv("GITHUB_TOKEN", "test_token")
	os.Setenv("GITHUB_OWNER", "test_owner")
	os.Setenv("GITHUB_REPO", "test_repo")
	defer func() {
		// Clean up environment variables after test
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITHUB_OWNER")
		os.Unsetenv("GITHUB_REPO")
	}()

	cfg, err := config.LoadGitHubConfig()
	if err != nil {
		t.Fatalf("LoadGitHubConfig failed unexpectedly: %v", err)
	}

	if cfg.Token != "test_token" {
		t.Errorf("Expected token 'test_token', got '%s'", cfg.Token)
	}
	if cfg.Owner != "test_owner" {
		t.Errorf("Expected owner 'test_owner', got '%s'", cfg.Owner)
	}
	if cfg.Repo != "test_repo" {
		t.Errorf("Expected repo 'test_repo', got '%s'", cfg.Repo)
	}
}

func TestLoadGitHubConfig_MissingToken(t *testing.T) {
	// Set up mock environment variables (missing token)
	os.Unsetenv("GITHUB_TOKEN") // Ensure it's unset
	os.Setenv("GITHUB_OWNER", "test_owner")
	os.Setenv("GITHUB_REPO", "test_repo")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITHUB_OWNER")
		os.Unsetenv("GITHUB_REPO")
	}()

	_, err := config.LoadGitHubConfig()
	if err == nil {
		t.Fatal("Expected an error when GITHUB_TOKEN is missing, but got none")
	}
	expectedErr := "GITHUB_TOKEN environment variable not set"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestLoadGitHubConfig_MissingOwner(t *testing.T) {
	os.Setenv("GITHUB_TOKEN", "test_token")
	os.Unsetenv("GITHUB_OWNER") // Ensure it's unset
	os.Setenv("GITHUB_REPO", "test_repo")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITHUB_OWNER")
		os.Unsetenv("GITHUB_REPO")
	}()

	_, err := config.LoadGitHubConfig()
	if err == nil {
		t.Fatal("Expected an error when GITHUB_OWNER is missing, but got none")
	}
	expectedErr := "GITHUB_OWNER environment variable not set"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestLoadGitHubConfig_MissingRepo(t *testing.T) {
	os.Setenv("GITHUB_TOKEN", "test_token")
	os.Setenv("GITHUB_OWNER", "test_owner")
	os.Unsetenv("GITHUB_REPO") // Ensure it's unset
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GITHUB_OWNER")
		os.Unsetenv("GITHUB_REPO")
	}()

	_, err := config.LoadGitHubConfig()
	if err == nil {
		t.Fatal("Expected an error when GITHUB_REPO is missing, but got none")
	}
	expectedErr := "GITHUB_REPO environment variable not set"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}
