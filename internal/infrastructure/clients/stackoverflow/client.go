package stackoverflow

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
	requestTimeout       = 5 * time.Second
	minQuestionPathParts = 2
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewStackOverflowClient() *Client {
	return &Client{
		baseURL: "https://api.stackexchange.com/2.3",
		client:  &http.Client{Timeout: requestTimeout},
	}
}

func (c *Client) GetQuestionUpdate(ctx context.Context, questionID string) (ports.ResourceUpdate, error) {
	endpoint := c.baseURL + "/questions/" + questionID + "?site=stackoverflow"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ports.ResourceUpdate{}, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return ports.ResourceUpdate{}, fmt.Errorf("send request: %w", err)
	}
	defer closeResponseBody(resp)

	if resp.StatusCode != http.StatusOK {
		return ports.ResourceUpdate{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var data stackOverflowQuestionResponse
	if decodeErr := json.NewDecoder(resp.Body).Decode(&data); decodeErr != nil {
		return ports.ResourceUpdate{}, fmt.Errorf("decode response: %w", decodeErr)
	}
	if len(data.Items) == 0 {
		return ports.ResourceUpdate{}, errors.New("question not found")
	}

	return ports.ResourceUpdate{
		LastUpdate: time.Unix(data.Items[0].LastActivityDate, 0),
	}, nil
}

func (c *Client) ParseStackOverflowURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}
	if u.Host != "stackoverflow.com" {
		return "", errors.New("not stackoverflow url")
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < minQuestionPathParts || parts[0] != "questions" {
		return "", errors.New("invalid stackoverflow question url")
	}

	return parts[1], nil
}

func closeResponseBody(resp *http.Response) {
	_ = resp.Body.Close()
}
