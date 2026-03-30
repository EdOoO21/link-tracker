package scrappergrpc

import (
	"context"
	"fmt"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	ports "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
	pc "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client struct {
	client pc.ScrapperServiceClient
	logger ports.Logger
}

func NewGRPCScrapperClient(logger ports.Logger, conn grpc.ClientConnInterface) *Client {
	return &Client{
		client: pc.NewScrapperServiceClient(conn),
		logger: logger,
	}
}

func (c *Client) AddChat(chatID int64) error {
	ctx := context.Background()
	_, err := c.client.AddChat(ctx, &pc.ChatRequest{ChatId: chatID})
	if err != nil {
		return mapRPCError("add chat", err)
	}
	c.logger.Info("chat added successfully", "chatID", chatID)
	return nil
}

func (c *Client) DeleteChat(chatID int64) error {
	ctx := context.Background()
	_, err := c.client.DeleteChat(ctx, &pc.ChatRequest{ChatId: chatID})
	if err != nil {
		return mapRPCError("delete chat", err)
	}
	c.logger.Info("chat deleted successfully", "chatID", chatID)
	return nil
}

func (c *Client) TrackLink(chatID int64, url string, tags []string) error {
	ctx := context.Background()
	_, err := c.client.AddLink(ctx, &pc.AddLinkRequest{ChatId: chatID, Url: url, Tags: tags})
	if err != nil {
		return mapRPCError("track link", err)
	}
	c.logger.Info("link added to track successfully", "chatID", chatID)
	return nil
}

func (c *Client) UnTrackLink(chatID int64, url string) error {
	ctx := context.Background()
	_, err := c.client.DeleteLink(ctx, &pc.DeleteLinkRequest{ChatId: chatID, Url: url})
	if err != nil {
		return mapRPCError("untrack link", err)
	}
	c.logger.Info("link deleted successfully", "chatID", chatID)
	return nil
}

func (c *Client) ListLinks(chatID int64, tags []string) ([]domain.Link, error) {
	ctx := context.Background()
	resp, err := c.client.ListLinks(ctx, &pc.ListLinksRequest{ChatId: chatID, Tags: tags})
	if err != nil {
		return nil, mapRPCError("list links", err)
	}

	links := make([]domain.Link, 0, len(resp.GetLinks()))
	for _, link := range resp.GetLinks() {
		tagSet := make(map[string]struct{}, len(link.GetTags()))
		for _, tag := range link.GetTags() {
			tagSet[tag] = struct{}{}
		}

		lastUpdate := time.Time{}
		if ts := link.GetLastUpdate(); ts != nil {
			lastUpdate = ts.AsTime()
		}

		links = append(links, domain.Link{
			URL:        link.GetUrl(),
			Tags:       tagSet,
			LastUpdate: lastUpdate,
		})
	}

	c.logger.Info("link list got successfully", "chatID", chatID)
	return links, nil
}

func mapRPCError(action string, err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("%s: %w", action, err)
	}

	switch st.Code() {
	case codes.NotFound:
		switch st.Message() {
		case ports.ErrChatNotFound.Error():
			return fmt.Errorf("%s: %w", action, ports.ErrChatNotFound)
		case ports.ErrLinkNotFound.Error():
			return fmt.Errorf("%s: %w", action, ports.ErrLinkNotFound)
		}
	case codes.AlreadyExists:
		switch st.Message() {
		case ports.ErrChatAlreadyExists.Error():
			return fmt.Errorf("%s: %w", action, ports.ErrChatAlreadyExists)
		case ports.ErrLinkAlreadyExists.Error():
			return fmt.Errorf("%s: %w", action, ports.ErrLinkAlreadyExists)
		}
	}

	return fmt.Errorf("%s: %s", action, st.Message())
}
