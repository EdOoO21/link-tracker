package bot

type PostUpdateRequest struct {
	URL         string  `json:"url"`
	ChatIDs     []int64 `json:"tgChatIds"`
	Description string  `json:"description"`
}
