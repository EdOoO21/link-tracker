package scrapper

import (
	"context"
	"errors"
	"fmt"
	"slices"

	models "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/models"
	domain "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	ports "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

func (s *Scrapper) AddLink(ctx context.Context, link models.AddLink) error {
	s.logger.Info("add link requested", "chatID", link.ChatID, "url", link.URL, "tags", link.Tags)
	err := s.repo.TrackLink(ctx, link.ChatID, link.URL, link.Tags)
	if err != nil {
		switch {
		case errors.Is(err, ports.ErrChatNotFound):
			s.logger.Warn("add link failed: chat not found", "chatID", link.ChatID, "url", link.URL)
		case errors.Is(err, ports.ErrLinkAlreadyExists):
			s.logger.Warn("add link failed: link already exists", "chatID", link.ChatID, "url", link.URL)
		default:
			s.logger.Error("add link failed", "chatID", link.ChatID, "url", link.URL, "error", err)
		}
		return fmt.Errorf("track link in repository: %w", err)
	}

	s.logger.Info("link added", "chatID", link.ChatID, "url", link.URL, "tags", link.Tags)
	return nil
}

func (s *Scrapper) DeleteLink(ctx context.Context, link models.DeleteLink) error {
	s.logger.Info("delete link requested", "chatID", link.ChatID, "url", link.URL)
	err := s.repo.UnTrackLink(ctx, link.ChatID, link.URL)
	if err != nil {
		switch {
		case errors.Is(err, ports.ErrChatNotFound):
			s.logger.Warn("delete link failed: chat not found", "chatID", link.ChatID, "url", link.URL)
		case errors.Is(err, ports.ErrLinkNotFound):
			s.logger.Warn("delete link failed: link not found", "chatID", link.ChatID, "url", link.URL)
		default:
			s.logger.Error("delete link failed", "chatID", link.ChatID, "url", link.URL, "error", err)
		}
		return fmt.Errorf("untrack link in repository: %w", err)
	}

	s.logger.Info("link deleted", "chatID", link.ChatID, "url", link.URL)
	return nil
}

func (s *Scrapper) AddTag(ctx context.Context, tag models.AddTag) error {
	return s.mutateTag(
		ctx,
		"add",
		"added",
		tag.ChatID,
		tag.URL,
		tag.Tag,
		s.repo.AddTag,
		ports.ErrTagAlreadyExists,
		"tag already exists",
	)
}

func (s *Scrapper) GetTags(ctx context.Context, chatID int64, url string) ([]string, error) {
	s.logger.Info("get tags requested", "chatID", chatID, "url", url)
	tags, err := s.repo.GetTags(ctx, chatID, url)
	if err != nil {
		switch {
		case errors.Is(err, ports.ErrChatNotFound):
			s.logger.Warn("get tags failed: chat not found", "chatID", chatID, "url", url)
		case errors.Is(err, ports.ErrLinkNotFound):
			s.logger.Warn("get tags failed: link not found", "chatID", chatID, "url", url)
		default:
			s.logger.Error("get tags failed", "chatID", chatID, "url", url, "error", err)
		}
		return nil, fmt.Errorf("get tags in repository: %w", err)
	}

	slices.Sort(tags)
	s.logger.Info("tags loaded", "chatID", chatID, "url", url, "count", len(tags))
	return tags, nil
}

func (s *Scrapper) DeleteTag(ctx context.Context, tag models.DeleteTag) error {
	return s.mutateTag(
		ctx,
		"delete",
		"deleted",
		tag.ChatID,
		tag.URL,
		tag.Tag,
		s.repo.DeleteTag,
		ports.ErrTagNotFound,
		"tag not found",
	)
}

func (s *Scrapper) GetLinks(ctx context.Context, chatID int64, tags []string) ([]domain.Link, error) {
	s.logger.Info("get links requested", "chatID", chatID, "tags", tags)
	links, err := s.repo.ListLinks(ctx, chatID, tags)
	if err != nil {
		switch {
		case errors.Is(err, ports.ErrChatNotFound):
			s.logger.Warn("get links failed: chat not found", "chatID", chatID, "tags", tags)
		default:
			s.logger.Error("get links failed", "chatID", chatID, "tags", tags, "error", err)
		}
		return nil, fmt.Errorf("list links in repository: %w", err)
	}

	s.logger.Info("links loaded", "chatID", chatID, "tags", tags, "count", len(links))
	return links, nil
}

func (s *Scrapper) AddChat(ctx context.Context, chatID int64) error {
	s.logger.Info("add chat requested", "chatID", chatID)
	err := s.repo.AddChat(ctx, chatID)
	if err != nil {
		switch {
		case errors.Is(err, ports.ErrChatAlreadyExists):
			s.logger.Warn("add chat failed: chat already exists", "chatID", chatID)
		default:
			s.logger.Error("add chat failed", "chatID", chatID, "error", err)
		}
		return fmt.Errorf("add chat in repository: %w", err)
	}

	s.logger.Info("chat added", "chatID", chatID)
	return nil
}

func (s *Scrapper) DeleteChat(ctx context.Context, chatID int64) error {
	s.logger.Info("delete chat requested", "chatID", chatID)
	err := s.repo.DeleteChat(ctx, chatID)
	if err != nil {
		switch {
		case errors.Is(err, ports.ErrChatNotFound):
			s.logger.Warn("delete chat failed: chat not found", "chatID", chatID)
		default:
			s.logger.Error("delete chat failed", "chatID", chatID, "error", err)
		}
		return fmt.Errorf("delete chat in repository: %w", err)
	}

	s.logger.Info("chat deleted", "chatID", chatID)
	return nil
}

func (s *Scrapper) mutateTag(
	ctx context.Context,
	action string,
	successAction string,
	chatID int64,
	url string,
	tag string,
	operation func(context.Context, int64, string, string) error,
	expectedErr error,
	expectedWarn string,
) error {
	s.logger.Info(action+" tag requested", "chatID", chatID, "url", url, "tag", tag)
	err := operation(ctx, chatID, url, tag)
	if err != nil {
		switch {
		case errors.Is(err, ports.ErrChatNotFound):
			s.logger.Warn(action+" tag failed: chat not found", "chatID", chatID, "url", url, "tag", tag)
		case errors.Is(err, ports.ErrLinkNotFound):
			s.logger.Warn(action+" tag failed: link not found", "chatID", chatID, "url", url, "tag", tag)
		case errors.Is(err, expectedErr):
			s.logger.Warn(action+" tag failed: "+expectedWarn, "chatID", chatID, "url", url, "tag", tag)
		default:
			s.logger.Error(action+" tag failed", "chatID", chatID, "url", url, "tag", tag, "error", err)
		}
		return fmt.Errorf("%s tag in repository: %w", action, err)
	}

	s.logger.Info("tag "+successAction, "chatID", chatID, "url", url, "tag", tag)
	return nil
}
