package models

type AddLink struct {
	ChatID int64    `json:"chatID"`
	URL    string   `json:"url"`
	Tags   []string `json:"tags"`
}
