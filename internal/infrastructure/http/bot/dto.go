package bot

type PostUpdateRequest struct {
	URL         string  `json:"url"`
	ChatIDS     []int64 `json:"tgChatIds"`
	Description string  `json:"description"`
}
