package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"cloudcostguard/internal/config"
	"cloudcostguard/internal/github"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(analyzeCmd)
}

// analyzeCmd represents the analyze command, which is the main entry point for the CLI tool.
// It reads a Terraform plan, sends it to the backend for cost estimation, and posts the results
// as a comment on a GitHub pull request.
var analyzeCmd = &cobra.Command{
	Use:   "analyze [PLAN_JSON_PATH]",
	Short: "Analyzes a Terraform plan and posts the cost estimate to a GitHub PR.",
	Args:  cobra.RangeArgs(0, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		// --- Configuration resolution ---
		planPath := "plan.json"
		if len(args) > 0 {
			planPath = args[0]
		}

		repo := ""
		prNumberStr := ""

		cfg, err := config.LoadConfig(".cloudcostguard.yml")
		if err == nil {
			repo = cfg.GitHub.Repo
			prNumberStr = fmt.Sprintf("%d", cfg.GitHub.PRNumber)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("could not load config file: %w", err)
		}

		if len(args) == 3 {
			repo = args[1]
			prNumberStr = args[2]
		}

		if repo == "" || prNumberStr == "" {
			return fmt.Errorf("github repo and PR number must be provided")
		}

		githubToken := os.Getenv("GITHUB_TOKEN")
		if githubToken == "" {
			return fmt.Errorf("GITHUB_TOKEN environment variable not set")
		}

		// --- Execution ---
		planFile, err := os.Open(planPath)
		if err != nil {
			return fmt.Errorf("could not open plan file: %w", err)
		}
		defer planFile.Close()

		// 1. Call the backend API
		backendURL := os.Getenv("CCG_BACKEND_URL")
		if backendURL == "" {
			backendURL = "http://localhost:8080"
		}

		req, err := http.NewRequest("POST", backendURL+"/estimate", planFile)
		if err != nil {
			return fmt.Errorf("failed to create request to backend: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to call backend: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("backend returned an error: %s", resp.Status)
		}

		var result map[string]float64
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to decode backend response: %w", err)
		}
		cost := result["estimated_monthly_cost"]

		// 2. Post the comment to GitHub
		comment := fmt.Sprintf("## CloudCostGuard Analysis 🤖\n\nEstimated Monthly Cost Impact: **$%.2f**", cost)
		if err := github.PostComment(repo, prNumberStr, githubToken, comment); err != nil {
			return fmt.Errorf("could not post comment to GitHub: %w", err)
		}

		fmt.Println("Successfully posted cost analysis to GitHub.")
		return nil
	},
}
