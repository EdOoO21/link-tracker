package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

const (
	requestTimeout  = 5 * time.Second
	minRepoPathPart = 2
	perPageLimit    = 10
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

func (c *Client) GetRepoUpdate(ctx context.Context, owner, repo string, since time.Time) (ports.ResourceUpdate, error) {
	endpoint := c.baseURL + "/repos/" + owner + "/" + repo + "/issues"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ports.ResourceUpdate{}, fmt.Errorf("create request: %w", err)
	}
	query := req.URL.Query()
	query.Set("state", "all")
	query.Set("sort", "created")
	query.Set("direction", "desc")
	query.Set("per_page", strconv.Itoa(perPageLimit))
	req.URL.RawQuery = query.Encode()
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.client.Do(req)
	if err != nil {
		return ports.ResourceUpdate{}, fmt.Errorf("send request: %w", err)
	}
	defer closeResponseBody(resp)

	if resp.StatusCode != http.StatusOK {
		return ports.ResourceUpdate{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var data githubIssuesResponse
	if decodeErr := json.NewDecoder(resp.Body).Decode(&data); decodeErr != nil {
		return ports.ResourceUpdate{}, fmt.Errorf("decode response: %w", decodeErr)
	}

	return data.toResourceUpdate(since), nil
}

func (c *Client) CanHandle(raw string) bool {
	_, _, err := c.ParseGitHubURL(raw)
	return err == nil
}

func (c *Client) CheckUpdate(ctx context.Context, raw string, since time.Time) (ports.ResourceUpdate, error) {
	owner, repo, err := c.ParseGitHubURL(raw)
	if err != nil {
		return ports.ResourceUpdate{}, fmt.Errorf("parse github url: %w", err)
	}

	return c.GetRepoUpdate(ctx, owner, repo, since)
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
