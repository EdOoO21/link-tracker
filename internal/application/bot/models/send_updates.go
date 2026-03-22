package models

type SendUpdates struct {
	URL         string  `json:"url"`
	ChatIDS     []int64 `json:"tgChatIds"`
	Description string  `json:"description"`
}
