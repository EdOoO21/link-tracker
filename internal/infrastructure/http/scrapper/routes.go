package scrapper

import (
	"net/http"
	"strconv"
	"strings"

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
	mux.HandleFunc("/tg-chat/", h.handleTgChatID)
}

func (h *Handler) handleLinks(w http.ResponseWriter, r *http.Request) {
	chatID, err := strconv.ParseInt(r.Header.Get("Tg-Chat-Id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid header value: Tg-Chat-Id", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPost:
		h.PostLinks(w, r, chatID)
	case http.MethodGet:
		h.GetLinks(w, r, chatID)
	case http.MethodDelete:
		h.DeleteLinks(w, r, chatID)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleTgChatID(w http.ResponseWriter, r *http.Request) {
	idPart := strings.TrimPrefix(r.URL.Path, "/tg-chat/")
	if idPart == "" || strings.Contains(idPart, "/") {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	chatID, err := strconv.ParseInt(idPart, 10, 64)
	if err != nil {
		http.Error(w, "invalid chat id", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPost:
		h.PostTgChatID(w, r, chatID)
	case http.MethodDelete:
		h.DeleteTgChatID(w, r, chatID)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
