package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"cloudcostguard/backend/estimator"
	"cloudcostguard/internal/config"
	"cloudcostguard/internal/github"
	"github.com/spf13/cobra"
)

var region string

func init() {
	rootCmd.AddCommand(analyzeCmd)
	analyzeCmd.Flags().StringVar(&region, "region", "", "AWS region to use for pricing")
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
		resolvedRegion := "us-east-1" // Default region
		usageEstimates := estimator.UsageEstimates{}

		cfg, err := config.LoadConfig(".cloudcostguard.yml")
		if err == nil {
			repo = cfg.GitHub.Repo
			prNumberStr = fmt.Sprintf("%d", cfg.GitHub.PRNumber)
			if cfg.Region != "" {
				resolvedRegion = cfg.Region
			}
			usageEstimates = cfg.UsageEstimates
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("could not load config file: %w", err)
		}

		// Command-line flag overrides config file
		if region != "" {
			resolvedRegion = region
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
		planBytes, err := os.ReadFile(planPath)
		if err != nil {
			return fmt.Errorf("could not read plan file: %w", err)
		}

		// 1. Call the backend API
		backendURL := os.Getenv("CCG_BACKEND_URL")
		if backendURL == "" {
			backendURL = "http://localhost:8080"
		}

		requestBody, err := json.Marshal(map[string]interface{}{
			"plan":           json.RawMessage(planBytes),
			"usage_estimates": usageEstimates,
		})
		if err != nil {
			return fmt.Errorf("could not marshal request body: %w", err)
		}

		url := fmt.Sprintf("%s/estimate?region=%s", backendURL, resolvedRegion)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
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

		var result estimator.EstimationResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to decode backend response: %w", err)
		}

		// 2. Post the comment to GitHub
		comment := formatComment(result)
		if err := github.PostComment(repo, prNumberStr, githubToken, comment); err != nil {
			return fmt.Errorf("could not post comment to GitHub: %w", err)
		}

		fmt.Println("Successfully posted cost analysis to GitHub.")
		return nil
	},
}

func formatComment(result estimator.EstimationResponse) string {
	var builder strings.Builder
	builder.WriteString("## CloudCostGuard Analysis 🤖\n\n")
	builder.WriteString(fmt.Sprintf("Estimated Monthly Cost Impact: **$%.2f**\n\n", result.TotalMonthlyCost))

	if len(result.Resources) > 0 {
		builder.WriteString("| Resource | Monthly Cost | Details |\n")
		builder.WriteString("| :--- | :--- | :--- |\n")
		for _, resource := range result.Resources {
			builder.WriteString(fmt.Sprintf("| `%s` | `$%.2f` | %s |\n", resource.Address, resource.MonthlyCost, resource.CostBreakdown))
		}
	}

	return builder.String()
}
