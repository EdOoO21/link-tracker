package bot

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) PostUpdates(w http.ResponseWriter, r *http.Request) {
	defer h.closeBody(r)
	var req PostUpdateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if len(req.ChatIDS) <= 0 {
		http.Error(w, "chatIDs is empty in json body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "url is empty in json body", http.StatusBadRequest)
		return
	}

	if err = h.service.SendUpdateMessages(ToSendUpdates(req)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) closeBody(r *http.Request) {
	if err := r.Body.Close(); err != nil {
		h.logger.Error("failed to close request body", "error", err)
	}
}
