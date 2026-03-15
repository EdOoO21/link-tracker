package scrapper

import (
	"net/http"

	service "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/interfaces"
	logger "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

type Handler struct {
	scrapperService service.ScrapperService
	logger          logger.Logger
}

func NewHandler(scrapperService service.ScrapperService, logger logger.Logger) *Handler {
	return &Handler{
		scrapperService: scrapperService,
		logger:          logger,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/links", h.handleLinks)
	mux.HandleFunc("/tg-chat", h.handleTgChatID)
}

func (h *Handler) handleLinks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.PostLinks(w, r)
	case http.MethodGet:
		h.GetLinks(w, r)
	case http.MethodDelete:
		h.DeleteLinks(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleTgChatID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.PostTgChatID(w, r)
	case http.MethodDelete:
		h.DeleteTgChatID(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
