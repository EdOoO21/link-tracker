package scrapper

import "time"

type TrackLinkRequest struct {
	URL  string   `json:"url"`
	Tags []string `json:"tags"`
}

type UnTrackLinkRequest struct {
	URL string `json:"url"`
}

type LinkResponse struct {
	URL        string    `json:"url"`
	Tags       []string  `json:"tags"`
	LastUpdate time.Time `json:"lastupdate"`
}

type ListLinksResponse struct {
	Links []LinkResponse `json:"links"`
	Size  int            `json:"size"`
}
