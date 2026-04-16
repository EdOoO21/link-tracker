package stackoverflow

type stackOverflowQuestionResponse struct {
	Items []stackOverflowQuestion `json:"items"`
}

type stackOverflowQuestion struct {
	Title string `json:"title"`
}

type stackOverflowAnswersResponse struct {
	Items []stackOverflowAnswer `json:"items"`
}

type stackOverflowCommentsResponse struct {
	Items []stackOverflowComment `json:"items"`
}

type stackOverflowAnswer struct {
	Body         string             `json:"body"`
	CreationDate int64              `json:"creation_date"`
	Owner        stackOverflowOwner `json:"owner"`
}

type stackOverflowComment struct {
	Body         string             `json:"body"`
	CreationDate int64              `json:"creation_date"`
	Owner        stackOverflowOwner `json:"owner"`
}

type stackOverflowOwner struct {
	DisplayName string `json:"display_name"`
}
