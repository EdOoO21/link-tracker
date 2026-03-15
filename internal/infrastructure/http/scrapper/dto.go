package scrapper

import "time"

type AddLinkRequest struct {
	ChatID int64
	URL    string   `json:"url"`
	Tags   []string `json:"tags"`
}

type DeleteLinkRequest struct {
	ChatID int64
	URL    string `json:"url"`
}

type GetLinkResponse struct {
	ChatID     int64     `json:"chatID"`
	URL        string    `json:"url"`
	Tags       []string  `json:"tags"`
	LastUpdate time.Time `json:"lastupdate"`
}
