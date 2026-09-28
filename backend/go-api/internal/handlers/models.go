package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/OmaleGrace/Grace-Predict/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ModelHandler struct {
	Models *models.Repository
}

func NewModelHandler(
	modelsRepo *models.Repository,
) *ModelHandler {
	return &ModelHandler{
		Models: modelsRepo,
	}
}

type CreateModelRequest struct {
	ID         *string                `json:"id"`
	DatasetID  *string                `json:"dataset_id"`
	Name       string                 `json:"name"`
	ModelType  string                 `json:"model_type"`
	TaskType   string                 `json:"task_type"`
	Metrics    map[string]interface{} `json:"metrics"`
	Parameters map[string]interface{} `json:"parameters"`
}

func (h *ModelHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req CreateModelRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	modelID := uuid.New()

	if req.ID != nil && *req.ID != "" {
		parsedID, err := uuid.Parse(*req.ID)
		if err != nil {
			http.Error(
				w,
				"invalid model id",
				http.StatusBadRequest,
			)
			return
		}

		modelID = parsedID
	}

	var datasetID *uuid.UUID

	if req.DatasetID != nil && *req.DatasetID != "" {
		parsedDatasetID, err := uuid.Parse(*req.DatasetID)
		if err != nil {
			http.Error(
				w,
				"invalid dataset_id",
				http.StatusBadRequest,
			)
			return
		}

		datasetID = &parsedDatasetID
	}

	if req.Name == "" {
		req.Name = "Untitled Model"
	}

	if req.ModelType == "" {
		req.ModelType = "random_forest"
	}

	if req.TaskType == "" {
		http.Error(
			w,
			"task_type is required",
			http.StatusBadRequest,
		)
		return
	}

	result, err := h.Models.Create(
		r.Context(),
		models.CreateModelParams{
			ID:         modelID,
			DatasetID:  datasetID,
			Name:       req.Name,
			ModelType:  req.ModelType,
			TaskType:   req.TaskType,
			Metrics:    req.Metrics,
			Parameters: req.Parameters,
		},
	)

	if err != nil {
		http.Error(
			w,
			"failed to load models: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      result,
		"message": "model saved",
	})
}

func (h *ModelHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	result, err := h.Models.List(r.Context())

	if err != nil {
	http.Error(
		w,
		"failed to load models: "+err.Error(),
		http.StatusInternalServerError,
	)
	return
}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"models": result,
	})
}

func (h *ModelHandler) Get(
	w http.ResponseWriter,
	r *http.Request,
) {
	idString := r.PathValue("id")

	id, err := uuid.Parse(idString)

	if err != nil {
		http.Error(
			w,
			"invalid model id",
			http.StatusBadRequest,
		)
		return
	}

	result, err := h.Models.GetByID(
		r.Context(),
		id,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(
			w,
			"model not found",
			http.StatusNotFound,
		)
		return
	}

	if err != nil {
		http.Error(
			w,
			"failed to load model",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(result)
}
