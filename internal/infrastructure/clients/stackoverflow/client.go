package stackoverflow

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

func (c *Client) GetQuestionUpdate(ctx context.Context, questionID string, since time.Time) (ports.ResourceUpdate, error) {
	title, err := c.fetchQuestionTitle(ctx, questionID)
	if err != nil {
		return ports.ResourceUpdate{}, err
	}

	answers, err := c.fetchAnswers(ctx, questionID)
	if err != nil {
		return ports.ResourceUpdate{}, err
	}

	comments, err := c.fetchComments(ctx, questionID)
	if err != nil {
		return ports.ResourceUpdate{}, err
	}

	return stackOverflowUpdatesResponse{
		Title:    title,
		Answers:  answers,
		Comments: comments,
	}.toResourceUpdate(since), nil
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

func (c *Client) fetchQuestionTitle(ctx context.Context, questionID string) (string, error) {
	endpoint := c.baseURL + "/questions/" + questionID + "?site=stackoverflow"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer closeResponseBody(resp)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var data stackOverflowQuestionResponse
	if decodeErr := json.NewDecoder(resp.Body).Decode(&data); decodeErr != nil {
		return "", fmt.Errorf("decode response: %w", decodeErr)
	}
	if len(data.Items) == 0 {
		return "", errors.New("question not found")
	}

	return data.Items[0].Title, nil
}

func (c *Client) fetchAnswers(ctx context.Context, questionID string) ([]stackOverflowAnswer, error) {
	data, err := fetchQuestionItems[stackOverflowAnswersResponse](ctx, c, questionID, "answers")
	if err != nil {
		return nil, err
	}

	return data.Items, nil
}

func (c *Client) fetchComments(ctx context.Context, questionID string) ([]stackOverflowComment, error) {
	data, err := fetchQuestionItems[stackOverflowCommentsResponse](ctx, c, questionID, "comments")
	if err != nil {
		return nil, err
	}

	return data.Items, nil
}

func fetchQuestionItems[T any](ctx context.Context, client *Client, questionID, resource string) (T, error) {
	var data T

	endpoint := client.baseURL + "/questions/" + questionID + "/" + resource
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return data, fmt.Errorf("create request: %w", err)
	}

	query := req.URL.Query()
	query.Set("site", "stackoverflow")
	query.Set("sort", "creation")
	query.Set("order", "desc")
	query.Set("pagesize", strconv.Itoa(itemsPerTypePageLimit))
	query.Set("filter", "withbody")
	req.URL.RawQuery = query.Encode()

	resp, err := client.client.Do(req)
	if err != nil {
		return data, fmt.Errorf("send request: %w", err)
	}
	defer closeResponseBody(resp)

	if resp.StatusCode != http.StatusOK {
		return data, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	if decodeErr := json.NewDecoder(resp.Body).Decode(&data); decodeErr != nil {
		return data, fmt.Errorf("decode response: %w", decodeErr)
	}

	return data, nil
}
