package ports

import "time"

type ResourceUpdate struct {
	HasUpdate  bool
	LastUpdate time.Time
	Message    string
}
