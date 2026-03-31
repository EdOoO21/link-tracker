package repository

import (
	"sync"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

type UsersLinks struct {
	mu    sync.RWMutex
	links map[int64][]domain.Link
}

func NewRepo() *UsersLinks {
	return &UsersLinks{
		links: make(map[int64][]domain.Link),
	}
}

func (u *UsersLinks) TrackLink(chatID int64, url string, tags []string) bool {
	u.mu.Lock()
	defer u.mu.Unlock()

	if links, ok := u.links[chatID]; ok {
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
	u.links[chatID] = append(u.links[chatID], obj)

	return true
}

func (u *UsersLinks) UnTrackLink(chatID int64, url string) bool {
	u.mu.Lock()
	defer u.mu.Unlock()

	if _, ok := u.links[chatID]; ok {
		for i := range u.links[chatID] {
			if u.links[chatID][i].URL == url {
				u.links[chatID][i], u.links[chatID][len(u.links[chatID])-1] = u.links[chatID][len(u.links[chatID])-1], u.links[chatID][i]
				u.links[chatID] = u.links[chatID][:len(u.links[chatID])-1]
				return true
			}
		}
	}
	return false
}

func (u *UsersLinks) ListLinks(chatID int64, tags []string) []domain.Link {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if links, ok := u.links[chatID]; ok {
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

func (u *UsersLinks) IsPresent(chatID int64) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()

	_, ok := u.links[chatID]
	return ok
}

func (u *UsersLinks) IsLinkPresent(chatID int64, url string) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()

	for _, link := range u.links[chatID] {
		if link.URL == url {
			return true
		}
	}
	return false
}

func (u *UsersLinks) AddChat(chatID int64) bool {
	u.mu.Lock()
	defer u.mu.Unlock()

	if _, ok := u.links[chatID]; !ok {
		u.links[chatID] = make([]domain.Link, 0)
		return true
	}
	return false
}

func (u *UsersLinks) DeleteChat(chatID int64) bool {
	u.mu.Lock()
	defer u.mu.Unlock()

	if _, ok := u.links[chatID]; ok {
		delete(u.links, chatID)
		return true
	}
	return false
}

func (u *UsersLinks) ListAllLinks() map[int64][]domain.Link {
	u.mu.RLock()
	defer u.mu.RUnlock()

	res := make(map[int64][]domain.Link, len(u.links))
	for chatID, links := range u.links {
		copied := make([]domain.Link, len(links))
		copy(copied, links)
		res[chatID] = copied
	}
	return res
}

func (u *UsersLinks) UpdateLinksLastUpdate(chatID int64, url string, lastUpdate time.Time) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	for i := range u.links[chatID] {
		if u.links[chatID][i].URL == url {
			u.links[chatID][i].LastUpdate = lastUpdate
			return nil
		}
	}
	return ports.ErrLinkNotFound
}
