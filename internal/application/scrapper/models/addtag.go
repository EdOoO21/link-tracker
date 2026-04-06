package models

type AddTag struct {
	ChatID int64  `json:"chatID"`
	URL    string `json:"url"`
	Tag    string `json:"tag"`
}
