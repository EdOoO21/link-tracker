package scrapper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	ports "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

type Client struct {
	sourceURL string
	client    *http.Client
	logger    ports.Logger
}

func NewClient(logger ports.Logger, sourceURL string) *Client {
	return &Client{
		sourceURL: sourceURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		logger: logger,
	}
}

func (c *Client) AddChat(chatID int64) error {
	endpoint := c.sourceURL + "/tg-chat/" + strconv.FormatInt(chatID, 10)
	resp, err := c.doRequest(http.MethodPost, endpoint, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("addchat request send: %w", err)
	}
	defer c.closeResp(resp)

	if resp.StatusCode == http.StatusConflict {
		return fmt.Errorf("chat already exists: %w", ports.ErrChatAlreadyExists)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	c.logger.Info("chat added successfully", "chatID", chatID)
	return nil
}

func (c *Client) DeleteChat(chatID int64) error {
	endpoint := c.sourceURL + "/tg-chat/" + strconv.FormatInt(chatID, 10)
	resp, err := c.doRequest(http.MethodDelete, endpoint, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("deletechat request send: %w", err)
	}
	defer c.closeResp(resp)

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("chat not found: %w", ports.ErrChatNotFound)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	c.logger.Info("chat deleted successfully", "chatID", chatID)
	return nil
}

func (c *Client) TrackLink(chatID int64, url string, tags []string) error {
	reqBody := TrackLinkRequest{
		URL:  url,
		Tags: tags,
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}
	endpoint := c.sourceURL + "/links"
	resp, err := c.doRequest(http.MethodPost, endpoint, &chatID, bytes.NewReader(data), nil)
	if err != nil {
		return fmt.Errorf("tracklink request send: %w", err)
	}
	defer c.closeResp(resp)

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("chat not found: %w", ports.ErrChatNotFound)
	}

	if resp.StatusCode == http.StatusConflict {
		return fmt.Errorf("link already exists: %w", ports.ErrLinkAlreadyExists)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	c.logger.Info("link added to track successfully", "chatID", chatID)
	return nil
}

func (c *Client) UnTrackLink(chatID int64, url string) error {
	reqBody := UnTrackLinkRequest{
		URL: url,
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}
	endpoint := c.sourceURL + "/links"
	resp, err := c.doRequest(http.MethodDelete, endpoint, &chatID, bytes.NewReader(data), nil)
	if err != nil {
		return fmt.Errorf("untracklink request send: %w", err)
	}
	defer c.closeResp(resp)

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("chat not found: %w", ports.ErrChatNotFound)
	}

	if resp.StatusCode == http.StatusConflict {
		return fmt.Errorf("link not found: %w", ports.ErrLinkNotFound)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	c.logger.Info("link deleted successfully", "chatID", chatID)
	return nil
}

func (c *Client) ListLinks(chatID int64, tags []string) ([]domain.Link, error) {
	endpoint := c.sourceURL + "/links"
	query := url.Values{}
	for _, tag := range tags {
		query.Add("tag", tag)
	}
	resp, err := c.doRequest(http.MethodGet, endpoint, &chatID, nil, query)
	if err != nil {
		return nil, fmt.Errorf("list links request send: %w", err)
	}
	defer c.closeResp(resp)

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("chat not found: %w", ports.ErrChatNotFound)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	data := &ListLinksResponse{}
	if err = json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	c.logger.Info("link list got successfully", "chatID", chatID)
	return ToDomainLinks(data.Links), nil
}

func (c *Client) doRequest(method, path string, chatID *int64, body io.Reader, query url.Values) (*http.Response, error) {
	if query != nil {
		encoded := query.Encode()
		if encoded != "" {
			path += "?" + encoded
		}
	}

	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}

	if chatID != nil {
		req.Header.Set("Tg-Chat-Id", strconv.FormatInt(*chatID, 10))
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) closeResp(resp *http.Response) {
	if closeErr := resp.Body.Close(); closeErr != nil {
		c.logger.Error("failed to close response body", "error", closeErr)
	}
}
