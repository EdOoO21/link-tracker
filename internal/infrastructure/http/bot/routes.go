package bot

import (
	"net/http"

	service "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/interfaces"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/logger"
)

type Handler struct {
	service service.BotService
	logger  *logger.Logger
}

func NewHandler(logger *logger.Logger, service service.BotService) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/updates", h.handleUpdate)
}

func (h *Handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.PostUpdates(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
