package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

type GitHubClient struct {
	baseURL string
	client  *http.Client
}

func NewGitHubClient() *GitHubClient {
	return &GitHubClient{
		baseURL: "https://api.github.com",
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *GitHubClient) GetRepoUpdate(owner, repo string) (ports.ResourceUpdate, error) {
	endpoint := c.baseURL + "/repos/" + owner + "/" + repo

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return ports.ResourceUpdate{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.client.Do(req)
	if err != nil {
		return ports.ResourceUpdate{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ports.ResourceUpdate{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var data githubRepoResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return ports.ResourceUpdate{}, err
	}

	last := data.PushedAt
	if data.UpdatedAt.After(last) {
		last = data.UpdatedAt
	}

	return ports.ResourceUpdate{LastUpdate: last}, nil
}

func (c *GitHubClient) ParseGitHubURL(raw string) (owner, repo string, err error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", err
	}
	if u.Host != "github.com" {
		return "", "", fmt.Errorf("not github url")
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid github repo url")
	}

	return parts[0], parts[1], nil
}
