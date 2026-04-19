package models

type FailedLinksReport struct {
	ChatID int64    `json:"tgChatId"`
	URLs   []string `json:"urls"`
}
