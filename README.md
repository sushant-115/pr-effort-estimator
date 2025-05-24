# **Pull Request Time/Effort Estimator**

A tool written in Go to analyze and estimate the review and merge times of Pull Requests (PRs) from GitHub repositories. By leveraging historical PR data, this tool helps teams understand their code review bottlenecks and improve efficiency.

## **Table of Contents**

* [Features](#bookmark=id.cf5356b9ooef)  
* [Factors Influencing PR Review Time](#bookmark=id.9xe2kgw8zc9m)  
* [Project Structure](#bookmark=id.y6t4ka14c85e)  
* [Getting Started](#bookmark=id.a9k4nkaeglmv)  
  * [Prerequisites](#bookmark=id.zctleeyu391a)  
  * [Installation](#bookmark=id.snqekw7g3hnb)  
  * [Configuration](#bookmark=id.62lhudly5uye)  
  * [Running the Tool](#bookmark=id.l1lbjqtw491c)  
* [How to Run Tests](#bookmark=id.2bz0k0jiwqtg)  
* [Future Enhancements](#bookmark=id.wazr4uz7e53k)  
* [Contributing](#bookmark=id.abu865k8u5i6)  
* [License](#bookmark=id.y5465kf2lv7j)

## **Features**

* **GitHub PR Data Fetching:** Connects to the GitHub API to retrieve detailed information about pull requests, including creation, merge, and close times, as well as additions, deletions, and changed files.  
* **Review Time Calculation:** Calculates key metrics such as:  
  * Time to First Review (Time from PR creation to the first review comment/approval).  
  * Time to Merge (Total time from PR creation to its merge).  
  * Review to Merge Time (Time from the first review to merge).  
* **Basic Analytics:** Provides aggregated statistics like average time to first review and average time to merge for historical PRs.  
* **Modular Design:** Structured into distinct packages for configuration, GitHub interaction, and metric calculation, promoting maintainability and testability.

## **Factors Influencing PR Review Time**

This tool considers and helps analyze the impact of the following factors on pull request review time:

* **Pull Request Size:** Larger PRs typically take longer to review. The tool extracts Additions, Deletions, and ChangedFiles.  
* **Complexity:** Inferred from ChangedFiles. More advanced analysis would require deeper code parsing.  
* **Team Experience:** Reflected in the observed average review and merge times.  
* **Context and Labels:** PR labels are fetched and can be used for further analysis (though not explicitly used for estimation in the current version).  
* **Reviewer Availability:** While not directly measured, its impact is embedded in the historical review times.

## **Project Structure**

The project follows a clean, modular Go project structure:

```pr-effort-estimator/  
├── main.go                 # Entry point of the application  
├── cmd/                    # Contains the core application logic  
│   └── app.go              # The main application execution flow  
│   └── app_test.go         # Integration tests for the core application logic  
├── api/github/             # Handles all interactions with the GitHub API  
│   └── client.go           # GitHub API client and logic  
│   └── types.go            # Data structures for GitHub PR data  
│   └── client_test.go      # Unit tests for the GitHub client  
├── interal/metrics/        # Contains logic for calculating PR-related metrics  
│   └── calculator.go       # Functions to compute time durations and aggregate stats  
│   └── calculator_test.go  # Unit tests for metric calculations  
├── pkg/config/             # Manages application configuration  
│   └── config.go           # Loads GitHub token, owner, and repository details  
│   └── config_test.go      # Unit tests for configuration loading  
└── README.md               # This README file
```

## **Sample Output**

```
2025/05/24 20:39:44 PR #4: Core controller (State: closed)
2025/05/24 20:39:44   Time to First Review: N/A (No reviews or PR still open)
2025/05/24 20:39:44   Time to Close: 24h25m46s
2025/05/24 20:39:44   Size: +43941 / -166, Files: 507
2025/05/24 20:39:44 ---
2025/05/24 20:39:44 PR #3: Project renamed, making it generic (State: closed)
2025/05/24 20:39:44   Time to First Review: N/A (No reviews or PR still open)
2025/05/24 20:39:44   Time to Close: 4m49s
2025/05/24 20:39:44   Size: +5607 / -698, Files: 76
2025/05/24 20:39:44 ---
2025/05/24 20:39:44 PR #2: custom apis (State: closed)
2025/05/24 20:39:44   Time to First Review: N/A (No reviews or PR still open)
2025/05/24 20:39:44   Time to Close : 5m26s
2025/05/24 20:39:44   Size: +540423 / -2445, Files: 1515
2025/05/24 20:39:44 ---
2025/05/24 20:39:44 PR #1: Init project (State: closed)
2025/05/24 20:39:44   Time to First Review: N/A (No reviews or PR still open)
2025/05/24 20:39:44   Time to Close: 22m50s
2025/05/24 20:39:44   Size: +2804 / -14, Files: 542
2025/05/24 20:39:44 ---
2025/05/24 20:39:44 
--- Aggregated Metrics (Simple Average) ---
2025/05/24 20:39:44 No merged PRs to calculate average time to merge.
2025/05/24 20:39:44 
--- Normal Distribution Based Estimates ---
Estimated Time to First Review (based on 23 PRs):
  Mean: 86h48m26.434782608s, StdDev: 159h20m22.686840149s
  50th Percentile (Median): 86h48m26.434782608s
  80th Percentile: 220h54m39.468086424s
  90th Percentile: 291h0m33.487134501s
  95th Percentile: 348h53m51.791733276s

Estimated Time to Merge (based on 77 merged PRs):
  Mean: 76h30m16.493506493s, StdDev: 197h33m57.49083277s
  50th Percentile (Median): 76h30m16.493506493s
  80th Percentile: 242h46m49.067904474s
  90th Percentile: 329h41m44.013357244s
  95th Percentile: 401h28m18.05992664s
```

## **Getting Started**

Follow these steps to set up and run the Pull Request Time Estimator.

### **Prerequisites**

* **Go (1.16 or higher):** Ensure you have Go installed on your system. You can download it from [golang.org](https://golang.org/dl/).  
* **GitHub Personal Access Token:** You need a GitHub Personal Access Token (PAT) with repo scope to access private repositories or public repositories with higher rate limits.  
  * Go to GitHub \-\> Settings \-\> Developer settings \-\> Personal access tokens \-\> Tokens (classic) \-\> Generate new token.  
  * Give it a descriptive name (e.g., pr-estimator-token).  
  * Select the repo scope.  
  * Copy the generated token immediately, as you won't be able to see it again.

### **Installation**

1. **Clone the repository:**  
   git clone https://github.com/sushant-115/pr-effort-estimator.git \# Replace with your repo URL  
   cd pr-effort-estimator

2. **Initialize Go Module and Download Dependencies:**  
   go mod tidy

### **Configuration**

The tool uses environment variables for GitHub authentication and repository details.

Set the following environment variables:

* GITHUB\_TOKEN: Your GitHub Personal Access Token.  
* GITHUB\_OWNER: The username or organization that owns the repository (e.g., octocat).  
* GITHUB\_REPO: The name of the repository (e.g., Spoon-Knife).

**Example (Linux/macOS):**

export GITHUB\_TOKEN="ghp\_YOUR\_ACTUAL\_GITHUB\_PATH"  
export GITHUB\_OWNER="your-github-username-or-org"  
export GITHUB\_REPO="your-repository-name"

**Example (Windows \- Command Prompt):**

set GITHUB\_TOKEN="ghp\_YOUR\_ACTUAL\_GITHUB\_PAT"  
set GITHUB\_OWNER="your-github-username-or-org"  
set GITHUB\_REPO="your-repository-name"

**Example (Windows \- PowerShell):**

$env:GITHUB\_TOKEN="ghp\_YOUR\_ACTUAL\_GITHUB\_PAT"  
$env:GITHUB\_OWNER="your-github-username-or-org"  
$env:GITHUB\_REPO="your-repository-name"

### **Running the Tool**

After setting up the environment variables, run the application from the project root:

go run main.go

The tool will fetch closed pull requests for the configured repository and print analysis results (average time to first review, average time to merge) to the console.

## **How to Run Tests**

To ensure the reliability of the tool, you can run the provided unit and integration tests.

From the project root directory, execute:

go test ./...

This command will run all tests in all subdirectories.

## **Future Enhancements**

* **Data Persistence:** Store historical PR data and calculated metrics in a database (e.g., SQLite, PostgreSQL) for more complex analysis and trend tracking.  
* **Predictive Modeling:** Implement statistical or machine learning models to estimate review times for *new* open pull requests based on their size, complexity, and historical data.  
* **GitHub Actions Integration:** Create a GitHub Action to automatically calculate and comment estimated review times on new PRs.  
* **Web Interface:** Develop a simple web UI using a Go web framework (e.g., Gin, Echo) to visualize trends and provide interactive reports.  
* **More Granular Metrics:** Track additional metrics such as the number of review rounds, re-request for reviews, and time spent in different review states.  
* **Advanced Complexity Analysis:** Integrate tools to analyze code complexity (e.g., cyclomatic complexity) to provide a more accurate measure of PR complexity.  
* **Reviewer Workload Awareness:** Integrate with other systems (e.g., Slack, calendar) to understand reviewer availability and workload, influencing estimation.

## **Contributing**

Contributions are welcome\! If you'd like to contribute, please follow these steps:

1. **Fork the repository.**  
2. **Create a new branch** for your feature or bug fix: git checkout \-b feature/your-feature-name or bugfix/fix-description.  
3. **Implement your changes** and write relevant tests.  
4. **Ensure all tests pass** (go test ./...).  
5. **Commit your changes** with a clear and concise message.  
6. **Push your branch** to your forked repository.  
7. **Open a Pull Request** to the main branch of the original repository.

Please ensure your code adheres to Go best practices and includes comprehensive tests.

### **Contributors**

* **\[Your Name/GitHub Handle Here\]** \- Initial development and core logic.  
* *(Add other contributors here as the project grows)*

## **License**

This project is licensed under the MIT License \- see the LICENSE file for details (if you plan to include one).
