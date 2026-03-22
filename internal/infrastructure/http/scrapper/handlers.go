package scrapper

import (
	"encoding/json"
	"errors"
	"net/http"

	ports "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

func (h *Handler) PostLinks(w http.ResponseWriter, r *http.Request, chatID int64) {
	defer h.closeBody(r)
	var req AddLinkRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if err = h.scrapperService.AddLink(ToAddLink(req, chatID)); err != nil {
		writeServiceError(w, err, "failed to add link")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteLinks(w http.ResponseWriter, r *http.Request, chatID int64) {
	defer h.closeBody(r)
	var req DeleteLinkRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if err = h.scrapperService.DeleteLink(ToDeleteLink(req, chatID)); err != nil {
		writeServiceError(w, err, "failed to delete link")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetLinks(w http.ResponseWriter, r *http.Request, chatID int64) {
	query := r.URL.Query()
	tags := query["tag"]

	links, err := h.scrapperService.GetLinks(chatID, tags)
	if err != nil {
		writeServiceError(w, err, "failed to get links")
		return
	}

	resp := &ListLinksResponse{
		Links: make([]LinkResponse, 0, len(links)),
		Size:  len(links),
	}

	for i := range links {
		resp.Links = append(resp.Links, ToGetLinkResponse(links[i]))
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

}

func (h *Handler) PostTgChatID(w http.ResponseWriter, r *http.Request, chatID int64) {
	if err := h.scrapperService.AddChat(chatID); err != nil {
		writeServiceError(w, err, "failed to add chat")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteTgChatID(w http.ResponseWriter, r *http.Request, chatID int64) {
	if err := h.scrapperService.DeleteChat(chatID); err != nil {
		writeServiceError(w, err, "failed to delete chat")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) closeBody(r *http.Request) {
	if err := r.Body.Close(); err != nil {
		h.logger.Error("failed to close request body", "error", err)
	}
}

func writeServiceError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, ports.ErrChatNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, ports.ErrChatAlreadyExists):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, ports.ErrLinkAlreadyExists):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, ports.ErrLinkNotFound):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		http.Error(w, fallback, http.StatusInternalServerError)
	}
}
