package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type commentBody struct {
	Body string `json:"body"`
}

// PostComment posts a comment to a GitHub pull request.
//
// Parameters:
//   repo: The GitHub repository in the format "owner/repo".
//   prNumber: The pull request number.
//   token: The GitHub API token.
//   comment: The comment to post.
//
// Returns:
//   An error if the comment could not be posted, nil otherwise.
func PostComment(repo, prNumber, token, comment string) error {
	apiURL := os.Getenv("GITHUB_API_URL")
	if apiURL == "" {
		apiURL = "https://api.github.com"
	}
	url := fmt.Sprintf("%s/repos/%s/issues/%s/comments", apiURL, repo, prNumber)

	body, err := json.Marshal(commentBody{Body: comment})
	if err != nil {
		return fmt.Errorf("could not marshal comment body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to post comment: %s", resp.Status)
	}

	return nil
}
