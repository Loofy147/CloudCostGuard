package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cloudcostguard",
	Short: "A CLI for providing pre-commit cloud cost estimation.",
	Long:  `CloudCostGuard analyzes Infrastructure-as-Code changes and posts the estimated cost impact to your pull requests.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
