package ports

import "errors"

var (
	ErrChatNotFound      = errors.New("chat not found")
	ErrChatAlreadyExists = errors.New("chat already exists")
	ErrLinkAlreadyExists = errors.New("link already exists")
	ErrLinkNotFound      = errors.New("link not found")
	ErrTagAlreadyExists  = errors.New("tag already exists")
	ErrTagNotFound       = errors.New("tag not found")
)
