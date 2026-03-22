package bot

type SendUpdatesRequest struct {
	URL         string  `json:"url"`
	ChatIDS     []int64 `json:"tgChatIds"`
	Description string  `json:"description"`
}
