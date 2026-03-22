package stackoverflow

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

type StackOverflowClient struct {
	baseURL string
	client  *http.Client
}

func NewStackOverflowClient() *StackOverflowClient {
	return &StackOverflowClient{
		baseURL: "https://api.stackexchange.com/2.3",
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *StackOverflowClient) GetQuestionUpdate(questionID string) (ports.ResourceUpdate, error) {
	endpoint := c.baseURL + "/questions/" + questionID + "?site=stackoverflow"

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return ports.ResourceUpdate{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return ports.ResourceUpdate{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ports.ResourceUpdate{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var data stackOverflowQuestionResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return ports.ResourceUpdate{}, err
	}
	if len(data.Items) == 0 {
		return ports.ResourceUpdate{}, fmt.Errorf("question not found")
	}

	return ports.ResourceUpdate{
		LastUpdate: time.Unix(data.Items[0].LastActivityDate, 0),
	}, nil
}

func (c *StackOverflowClient) ParseStackOverflowURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if u.Host != "stackoverflow.com" {
		return "", fmt.Errorf("not stackoverflow url")
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 || parts[0] != "questions" {
		return "", fmt.Errorf("invalid stackoverflow question url")
	}

	return parts[1], nil
}
