package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Run("loads a valid config file", func(t *testing.T) {
		configYAML := `
github:
  repo: "my-org/my-repo"
  pr_number: 123
`
		tmpfile, err := os.CreateTemp("", "config-*.yml")
		assert.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.WriteString(configYAML)
		assert.NoError(t, err)
		tmpfile.Close()

		config, err := LoadConfig(tmpfile.Name())
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "my-org/my-repo", config.GitHub.Repo)
		assert.Equal(t, 123, config.GitHub.PRNumber)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := LoadConfig("non-existent-file.yml")
		assert.Error(t, err)
	})

	t.Run("returns error for invalid yaml", func(t *testing.T) {
		configYAML := `invalid-yaml`
		tmpfile, err := os.CreateTemp("", "config-*.yml")
		assert.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.WriteString(configYAML)
		assert.NoError(t, err)
		tmpfile.Close()

		_, err = LoadConfig(tmpfile.Name())
		assert.Error(t, err)
	})
}
