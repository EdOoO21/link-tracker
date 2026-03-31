package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

type Client struct {
	sourceURL string
	client    *http.Client
	logger    ports.Logger
}

const requestTimeout = 5 * time.Second

func NewClient(logger ports.Logger, sourceURL string) *Client {
	return &Client{
		sourceURL: sourceURL,
		client: &http.Client{
			Timeout: requestTimeout,
		},
		logger: logger,
	}
}

func (c *Client) SendUpdates(chatIDs []int64, url, description string) error {
	endpoint := c.sourceURL + "/updates"
	reqBody := &SendUpdatesRequest{
		URL:         url,
		Description: description,
		ChatIDs:     chatIDs,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("request create: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}

	defer c.closeResp(resp)

	if resp.StatusCode == http.StatusBadRequest {
		return errors.New("invalid request")
	}

	if resp.StatusCode == http.StatusInternalServerError {
		return errors.New("failed to send updates to some chats")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) closeResp(resp *http.Response) {
	if closeErr := resp.Body.Close(); closeErr != nil {
		c.logger.Error("failed to close response body", "error", closeErr)
	}
}
