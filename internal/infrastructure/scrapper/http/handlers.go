package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	inf "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure"
)

type Handler struct {
	logger inf.Logger
}

func NewHandler(logger inf.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST("/links", h.AddLink)
	h.logger.Info("method post /links added")
}

func (h *Handler) AddLink(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
