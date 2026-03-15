package scrapper

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	appscrapper "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper"
)

func (h *Handler) PostLinks(w http.ResponseWriter, r *http.Request) {
	defer h.closeBody(r)
	var req AddLinkRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	chatID, err := strconv.ParseInt(r.Header.Get("Tg-Chat-Id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid header value: Tg-Chat-Id", http.StatusBadRequest)
		return
	}

	req.ChatID = chatID
	if err = h.scrapperService.AddLink(ToAddLink(req)); err != nil {
		writeServiceError(w, err, "failed to add link")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetLinks(w http.ResponseWriter, r *http.Request) {
	chatID, err := strconv.ParseInt(r.Header.Get("Tg-Chat-Id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid header value: Tg-Chat-Id", http.StatusBadRequest)
		return
	}

	query := r.URL.Query()
	tags := query["tag"]

	links, err := h.scrapperService.GetLinks(chatID, tags)
	if err != nil {
		writeServiceError(w, err, "failed to get links")
		return
	}

	resp := make([]GetLinkResponse, 0, len(links))
	for i := range links {
		resp = append(resp, ToGetLinkResponse(links[i]))
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

}

func (h *Handler) DeleteLinks(w http.ResponseWriter, r *http.Request) {
	defer h.closeBody(r)
	var req DeleteLinkRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	chatID, err := strconv.ParseInt(r.Header.Get("Tg-Chat-Id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid header value: Tg-Chat-Id", http.StatusBadRequest)
		return
	}

	req.ChatID = chatID
	if err = h.scrapperService.DeleteLink(ToDeleteLink(req)); err != nil {
		writeServiceError(w, err, "failed to delete link")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) PostTgChatID(w http.ResponseWriter, r *http.Request) {
	pathListed := strings.Split(strings.TrimPrefix(strings.Trim(r.URL.Path, "/"), "tg-chat/"), "/")

	if len(pathListed) != 1 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	chatID, err := strconv.ParseInt(pathListed[0], 10, 64)
	if err != nil {
		http.Error(w, "invalid header value: Tg-Chat-Id", http.StatusBadRequest)
		return
	}
	if err = h.scrapperService.AddChat(chatID); err != nil {
		writeServiceError(w, err, "failed to add chat")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteTgChatID(w http.ResponseWriter, r *http.Request) {
	pathListed := strings.Split(strings.TrimPrefix(strings.Trim(r.URL.Path, "/"), "tg-chat/"), "/")

	if len(pathListed) != 1 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	chatID, err := strconv.ParseInt(pathListed[0], 10, 64)
	if err != nil {
		http.Error(w, "invalid header value: Tg-Chat-Id", http.StatusBadRequest)
		return
	}

	if err = h.scrapperService.DeleteChat(chatID); err != nil {
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
	case errors.Is(err, appscrapper.ErrChatNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, appscrapper.ErrChatAlreadyExists):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, appscrapper.ErrLinkAlreadyExists):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, appscrapper.ErrLinkNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, fallback, http.StatusInternalServerError)
	}
}
