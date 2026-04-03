package github

import "time"

type githubRepoResponse struct {
	PushedAt  time.Time `json:"pushed_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
