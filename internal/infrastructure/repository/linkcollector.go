package repository

import (
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
)

type UsersLinks map[int64][]domain.Link

func NewRepo() UsersLinks {
	return make(map[int64][]domain.Link)
}

func (u UsersLinks) TrackLink(chatID int64, url string, tags []string) bool {
	if links, ok := u[chatID]; ok {
		for _, link := range links {
			if link.URL == url {
				return false
			}
		}
	}
	obj := domain.Link{
		URL:        url,
		Tags:       make(map[string]struct{}),
		LastUpdate: time.Now(),
	}
	for _, tag := range tags {
		obj.Tags[tag] = struct{}{}
	}
	u[chatID] = append(u[chatID], obj)

	return true
}

func (u UsersLinks) UnTrackLink(chatID int64, url string) bool {
	if _, ok := u[chatID]; ok {
		for i := range u[chatID] {
			if u[chatID][i].URL == url {
				u[chatID][i], u[chatID][len(u[chatID])-1] = u[chatID][len(u[chatID])-1], u[chatID][i]
				u[chatID] = u[chatID][:len(u[chatID])-1]
				return true
			}
		}
	}
	return false
}

func (u UsersLinks) ListLinks(chatID int64, tags []string) []domain.Link {
	if links, ok := u[chatID]; ok {
		res := make([]domain.Link, 0)

		for _, link := range links {
			flag := false
			for _, tag := range tags {
				if _, ok = link.Tags[tag]; !ok {
					flag = true
					break
				}
			}
			if !flag {
				res = append(res, link)
			}
		}

		return res
	}

	return nil
}

func (u UsersLinks) IsPresent(chatID int64) bool {
	_, ok := u[chatID]
	return ok
}

func (u UsersLinks) AddChat(chatID int64) bool {
	if _, ok := u[chatID]; !ok {
		u[chatID] = make([]domain.Link, 0)
		return true
	}
	return false
}

func (u UsersLinks) DeleteChat(chatID int64) bool {
	if _, ok := u[chatID]; ok {
		delete(u, chatID)
		return true
	}
	return false
}
