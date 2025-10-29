package config

import (
	"os"
	"gopkg.in/yaml.v2"
)

// Config represents the structure of the .cloudcostguard.yml file.
type Config struct {
	// GitHub contains the GitHub-related configuration.
	GitHub struct {
		// Repo is the GitHub repository in the format "owner/repo".
		Repo     string `yaml:"repo"`
		// PRNumber is the pull request number.
		PRNumber int    `yaml:"pr_number"`
	} `yaml:"github"`
	// Region is the AWS region to use for pricing.
	Region string `yaml:"region"`
}

// LoadConfig loads the configuration from the specified path.
//
// Parameters:
//   path: The path to the configuration file.
//
// Returns:
//   A pointer to the loaded Config object, or an error if the file could not be read or parsed.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
