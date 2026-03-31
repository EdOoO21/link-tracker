package bot

import "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/models"

func ToSendUpdates(req PostUpdateRequest) models.SendUpdates {
	return models.SendUpdates{
		URL:         req.URL,
		ChatIDs:     req.ChatIDs,
		Description: req.Description,
	}
}
