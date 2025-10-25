package config

import (
	"os"
	"gopkg.in/yaml.v2"
)

// Config represents the structure of the .cloudcostguard.yml file.
type Config struct {
	GitHub struct {
		Repo     string `yaml:"repo"`
		PRNumber int    `yaml:"pr_number"`
	} `yaml:"github"`
}

// LoadConfig loads the configuration from the specified path.
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
