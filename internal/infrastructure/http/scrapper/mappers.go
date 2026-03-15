package scrapper

import (
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/models"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
)

func ToAddLink(req AddLinkRequest) models.AddLink {
	return models.AddLink{
		ChatID: req.ChatID,
		URL:    req.URL,
		Tags:   req.Tags,
	}
}

func ToDeleteLink(req DeleteLinkRequest) models.DeleteLink {
	return models.DeleteLink{
		ChatID: req.ChatID,
		URL:    req.URL,
	}
}

func ToGetLinkResponse(link domain.Link) GetLinkResponse {
	tags := make([]string, 0, len(link.Tags))
	for k := range link.Tags {
		tags = append(tags, k)
	}

	return GetLinkResponse{
		ChatID:     link.ChatID,
		URL:        link.URL,
		Tags:       tags,
		LastUpdate: link.LastUpdate,
	}
}
