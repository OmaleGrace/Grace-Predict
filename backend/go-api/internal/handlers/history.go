package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/OmaleGrace/Grace-Predict/internal/auth"
	"github.com/OmaleGrace/Grace-Predict/internal/models"
	"github.com/google/uuid"
)

type HistoryHandler struct {
	Models *models.Repository
}

func NewHistoryHandler(modelsRepo *models.Repository) *HistoryHandler {
	return &HistoryHandler{
		Models: modelsRepo,
	}
}

func (h *HistoryHandler) List(w http.ResponseWriter, r *http.Request) {
	userIDString, ok := r.Context().Value(auth.UserIDKey).(string)

	if !ok || userIDString == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDString)

	if err != nil {
		http.Error(w, "invalid user id", http.StatusUnauthorized)
		return
	}

	limit := 20
	offset := 0

	if value := r.URL.Query().Get("limit"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			limit = parsed
		}
	}

	if value := r.URL.Query().Get("offset"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			offset = parsed
		}
	}

	history, err := h.Models.GetPredictionHistory(
		r.Context(),
		userID,
		limit,
		offset,
	)

	if err != nil {
		http.Error(w, "failed to load prediction history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"history": history,
		"limit":   limit,
		"offset":  offset,
	})
}