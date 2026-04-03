package models

type DeleteLink struct {
	ChatID int64  `json:"chatID"`
	URL    string `json:"url"`
}
