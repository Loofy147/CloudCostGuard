package main

import (
	"fmt"
	"os"
	"strconv"

	"cloudcostguard/internal/config"
	"cloudcostguard/internal/estimator"
	"cloudcostguard/internal/github"
	"cloudcostguard/internal/pricing"
	"cloudcostguard/internal/terraform"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(analyzeCmd)
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze [PLAN_JSON_PATH] [GITHUB_REPO] [PR_NUMBER]",
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

		// Load from config file first
		cfg, err := config.LoadConfig(".cloudcostguard.yml")
		if err == nil {
			repo = cfg.GitHub.Repo
			prNumberStr = strconv.Itoa(cfg.GitHub.PRNumber)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("could not load config file: %w", err)
		}

		// Override with CLI arguments if provided
		if len(args) == 3 {
			repo = args[1]
			prNumberStr = args[2]
		}

		// Final check
		if repo == "" || prNumberStr == "" {
			return fmt.Errorf("github repo and PR number must be provided via config file or CLI arguments")
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

		plan, err := terraform.ParsePlan(planFile)
		if err != nil {
			return fmt.Errorf("could not parse plan file: %w", err)
		}

		fmt.Println("Fetching latest AWS pricing data...")
		priceList := pricing.NewPriceList()
		if os.Getenv("CCG_TEST_MODE") == "true" {
			if err := priceList.LoadFromFile("../../testdata/aws_pricing.json"); err != nil {
				return fmt.Errorf("could not load mock pricing data: %w", err)
			}
		} else {
			urls := []string{
				"https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonEC2/current/index.json",
				"https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonRDS/current/index.json",
				"https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonS3/current/index.json",
				"https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonElasticLoadBalancing/current/index.json",
			}
			for _, url := range urls {
				if err := priceList.LoadFromURL(url); err != nil {
					fmt.Printf("Warning: could not load pricing data from %s: %v\n", url, err)
				}
			}
		}

		cost, err := estimator.Estimate(plan, priceList)
		if err != nil {
			return fmt.Errorf("could not estimate cost: %w", err)
		}

		comment := fmt.Sprintf("## CloudCostGuard Analysis 🤖\n\nEstimated Monthly Cost Impact: **$%.2f**", cost)
		if err := github.PostComment(repo, prNumberStr, githubToken, comment); err != nil {
			return fmt.Errorf("could not post comment to GitHub: %w", err)
		}

		fmt.Println("Successfully posted cost analysis to GitHub.")
		return nil
	},
}
