package domain

import "time"

type Link struct {
	URL        string              `json:"url"`
	Tags       map[string]struct{} `json:"tags"`
	LastUpdate time.Time           `json:"lastupdate"`
}
