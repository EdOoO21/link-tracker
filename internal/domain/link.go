package domain

import "time"

type Link struct {
	ChatID     int64               `json:"chatID"`
	URL        string              `json:"url"`
	Tags       map[string]struct{} `json:"tags"`
	LastUpdate time.Time           `json:"lastupdate"`
}
