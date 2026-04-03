package models

import "time"

type TrackedLink struct {
	LinkID     int64
	URL        string
	LastUpdate time.Time
	ChatIDs    []int64
}
