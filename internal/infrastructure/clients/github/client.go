package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

const (
	requestTimeout  = 5 * time.Second
	minRepoPathPart = 2
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewGitHubClient() *Client {
	return &Client{
		baseURL: "https://api.github.com",
		client:  &http.Client{Timeout: requestTimeout},
	}
}

func (c *Client) GetRepoUpdate(owner, repo string) (ports.ResourceUpdate, error) {
	endpoint := c.baseURL + "/repos/" + owner + "/" + repo

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, endpoint, nil)
	if err != nil {
		return ports.ResourceUpdate{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.client.Do(req)
	if err != nil {
		return ports.ResourceUpdate{}, fmt.Errorf("send request: %w", err)
	}
	defer closeResponseBody(resp)

	if resp.StatusCode != http.StatusOK {
		return ports.ResourceUpdate{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var data githubRepoResponse
	if decodeErr := json.NewDecoder(resp.Body).Decode(&data); decodeErr != nil {
		return ports.ResourceUpdate{}, fmt.Errorf("decode response: %w", decodeErr)
	}

	last := data.PushedAt
	if data.UpdatedAt.After(last) {
		last = data.UpdatedAt
	}

	return ports.ResourceUpdate{LastUpdate: last}, nil
}

func (c *Client) ParseGitHubURL(raw string) (owner, repo string, err error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", fmt.Errorf("parse url: %w", err)
	}
	if u.Host != "github.com" {
		return "", "", errors.New("not github url")
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < minRepoPathPart {
		return "", "", errors.New("invalid github repo url")
	}

	return parts[0], parts[1], nil
}

func closeResponseBody(resp *http.Response) {
	_ = resp.Body.Close()
}
