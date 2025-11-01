// Package main is the entry point for the CloudCostGuard CLI.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cloudcostguard",
	Short: "A CLI for providing pre-commit cloud cost estimation.",
	Long:  `CloudCostGuard analyzes Infrastructure-as-Code changes and posts the estimated cost impact to your pull requests.`,
}

// main is the entry point for the CloudCostGuard CLI.
// It executes the root command and handles any errors.
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
