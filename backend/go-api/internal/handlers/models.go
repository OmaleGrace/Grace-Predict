package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/OmaleGrace/Grace-Predict/internal/auth"
	"github.com/OmaleGrace/Grace-Predict/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TrainModelRequest struct {
	DatasetID    string `json:"dataset_id"`
	TargetColumn string `json:"target_column"`
	Name         string `json:"name"`
	ModelType    string `json:"model_type"`
}

type ModelHandler struct {
	Models       *models.Repository
	MLServiceURL string
}

func NewModelHandler(
	modelsRepo *models.Repository,
	mlServiceURL string,
) *ModelHandler {
	return &ModelHandler{
		Models:       modelsRepo,
		MLServiceURL: mlServiceURL,
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

func (h *ModelHandler) Train(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req TrainModelRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	datasetID, err := uuid.Parse(req.DatasetID)
	if err != nil {
		http.Error(w, "invalid dataset_id", http.StatusBadRequest)
		return
	}

	if req.TargetColumn == "" {
		http.Error(w, "target_column is required", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		req.Name = "Untitled Model"
	}

	if req.ModelType == "" {
		req.ModelType = "random_forest"
	}

	var (
		filePath string
	)

	err = h.Models.DB.QueryRow(
		r.Context(),
		`
        SELECT file_path
        FROM datasets
        WHERE id = $1
          AND user_id = $2
        `,
		datasetID,
		userID,
	).Scan(&filePath)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "dataset not found", http.StatusNotFound)
			return
		}

		http.Error(
			w,
			"failed to load dataset: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	modelID := uuid.New()

	payload := map[string]interface{}{
		"file_path":     filePath,
		"target_column": req.TargetColumn,
		"model_id":      modelID.String(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(
			w,
			"failed to create training request",
			http.StatusInternalServerError,
		)
		return
	}

	mlResponse, err := http.Post(
		h.MLServiceURL+"/models/train",
		"application/json",
		bytes.NewReader(payloadBytes),
	)

	if err != nil {
		http.Error(
			w,
			"ML training service unavailable",
			http.StatusBadGateway,
		)
		return
	}

	defer mlResponse.Body.Close()

	responseBody, err := io.ReadAll(mlResponse.Body)
	if err != nil {
		http.Error(
			w,
			"failed to read ML response",
			http.StatusBadGateway,
		)
		return
	}

	if mlResponse.StatusCode != http.StatusOK {
		http.Error(
			w,
			fmt.Sprintf(
				"ML training failed: %s",
				string(responseBody),
			),
			http.StatusBadGateway,
		)
		return
	}

	var trainingResult struct {
		ModelID      string                 `json:"model_id"`
		TaskType     string                 `json:"task_type"`
		TargetColumn string                 `json:"target_column"`
		Features     []string               `json:"features"`
		Rows         int                    `json:"rows"`
		TrainingRows int                    `json:"training_rows"`
		TestRows     int                    `json:"test_rows"`
		Metrics      map[string]interface{} `json:"metrics"`
	}

	if err := json.Unmarshal(
		responseBody,
		&trainingResult,
	); err != nil {
		http.Error(
			w,
			"invalid ML training response",
			http.StatusBadGateway,
		)
		return
	}

	returnedModelID, err := uuid.Parse(trainingResult.ModelID)
	if err != nil {
		http.Error(
			w,
			"ML returned invalid model id",
			http.StatusBadGateway,
		)
		return
	}

	if returnedModelID != modelID {
		http.Error(
			w,
			"ML model id mismatch",
			http.StatusBadGateway,
		)
		return
	}

	parameters := map[string]interface{}{
		"features":      trainingResult.Features,
		"target_column": trainingResult.TargetColumn,
		"training_rows": trainingResult.TrainingRows,
		"test_rows":     trainingResult.TestRows,
	}

	_, err = h.Models.Create(
		r.Context(),
		models.CreateModelParams{
			ID:         modelID,
			DatasetID:  &datasetID,
			Name:       req.Name,
			ModelType:  req.ModelType,
			TaskType:   trainingResult.TaskType,
			Metrics:    trainingResult.Metrics,
			Parameters: parameters,
		},
	)

	if err != nil {
		http.Error(
			w,
			"model trained but failed to save metadata: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"model_id":      modelID,
		"dataset_id":    datasetID,
		"name":          req.Name,
		"model_type":    req.ModelType,
		"task_type":     trainingResult.TaskType,
		"target_column": trainingResult.TargetColumn,
		"features":      trainingResult.Features,
		"rows":          trainingResult.Rows,
		"training_rows": trainingResult.TrainingRows,
		"test_rows":     trainingResult.TestRows,
		"metrics":       trainingResult.Metrics,
	})
}
