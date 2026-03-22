package scrapper

import (
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/models"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
)

func ToAddLink(req AddLinkRequest, chatID int64) models.AddLink {
	return models.AddLink{
		ChatID: chatID,
		URL:    req.URL,
		Tags:   req.Tags,
	}
}

func ToDeleteLink(req DeleteLinkRequest, chatID int64) models.DeleteLink {
	return models.DeleteLink{
		ChatID: chatID,
		URL:    req.URL,
	}
}

func ToGetLinkResponse(link domain.Link) LinkResponse {
	tags := make([]string, 0, len(link.Tags))
	for k := range link.Tags {
		tags = append(tags, k)
	}

	return LinkResponse{
		URL:        link.URL,
		Tags:       tags,
		LastUpdate: link.LastUpdate,
	}
}
