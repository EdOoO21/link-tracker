package scrapper

import "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"

func ToDomainLinks(data []LinkResponse) []domain.Link {
	list := make([]domain.Link, 0, len(data))
	for _, linkResp := range data {
		tags := make(map[string]struct{})
		for _, v := range linkResp.Tags {
			tags[v] = struct{}{}
		}
		linkDomain := domain.Link{
			URL:        linkResp.URL,
			LastUpdate: linkResp.LastUpdate,
			Tags:       tags,
		}
		list = append(list, linkDomain)
	}
	return list
}
