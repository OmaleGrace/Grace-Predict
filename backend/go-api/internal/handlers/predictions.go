package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/OmaleGrace/Grace-Predict/internal/auth"
	"github.com/OmaleGrace/Grace-Predict/internal/models"
	"github.com/google/uuid"
)

type PredictionHandler struct {
	MLServiceURL string
	Models       *models.Repository
}

func NewPredictionHandler(
	mlServiceURL string,
	modelRepo *models.Repository,
) *PredictionHandler {
	return &PredictionHandler{
		MLServiceURL: mlServiceURL,
		Models:       modelRepo,
	}
}

func (h *PredictionHandler) TestPrediction(w http.ResponseWriter, r *http.Request) {
	response, err := http.Get(h.MLServiceURL + "/predict/test")
	if err != nil {
		http.Error(w, "ML service unavailable", http.StatusBadGateway)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		http.Error(w, "ML service returned an error", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var prediction map[string]interface{}

	if err := json.NewDecoder(response.Body).Decode(&prediction); err != nil {
		http.Error(w, "Invalid response from ML service", http.StatusBadGateway)
		return
	}

	json.NewEncoder(w).Encode(prediction)
}

func (h *PredictionHandler) InspectDataset(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "CSV file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Filename == "" {
		http.Error(w, "invalid filename", http.StatusBadRequest)
		return
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", header.Filename)
	if err != nil {
		http.Error(w, "failed to prepare file", http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(part, file); err != nil {
		http.Error(w, "failed to read uploaded file", http.StatusBadRequest)
		return
	}

	if err := writer.Close(); err != nil {
		http.Error(w, "failed to prepare request", http.StatusInternalServerError)
		return
	}

	response, err := http.Post(
		h.MLServiceURL+"/datasets/inspect",
		writer.FormDataContentType(),
		&body,
	)
	if err != nil {
		http.Error(w, "ML service unavailable", http.StatusBadGateway)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		message, _ := io.ReadAll(response.Body)
		http.Error(
			w,
			fmt.Sprintf("ML service error: %s", string(message)),
			http.StatusBadGateway,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, response.Body)
}

type PredictRequest struct {
	ModelID   string                 `json:"model_id"`
	InputData map[string]interface{} `json:"input_data"`
}

func (h *PredictionHandler) Predict(w http.ResponseWriter, r *http.Request) {
	var req PredictRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	modelID, err := uuid.Parse(req.ModelID)
	if err != nil {
		http.Error(w, "invalid model_id", http.StatusBadRequest)
		return
	}

	if len(req.InputData) == 0 {
		http.Error(w, "input_data is required", http.StatusBadRequest)
		return
	}

	// Verify that the model exists in PostgreSQL BEFORE
	// asking the ML service to make the prediction.
	exists, err := h.Models.Exists(r.Context(), modelID)
	if err != nil {
		http.Error(w, "failed to verify model", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "model not found", http.StatusNotFound)
		return
	}

	payload, err := json.Marshal(req)
	if err != nil {
		http.Error(w, "failed to prepare prediction request", http.StatusInternalServerError)
		return
	}

	response, err := http.Post(
		h.MLServiceURL+"/models/predict",
		"application/json",
		bytes.NewReader(payload),
	)
	if err != nil {
		http.Error(w, "ML prediction service unavailable", http.StatusBadGateway)
		return
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		http.Error(w, "failed to read ML response", http.StatusBadGateway)
		return
	}

	if response.StatusCode != http.StatusOK {
		http.Error(
			w,
			fmt.Sprintf("ML prediction failed: %s", string(responseBody)),
			http.StatusBadGateway,
		)
		return
	}

	var prediction map[string]interface{}

	if err := json.Unmarshal(responseBody, &prediction); err != nil {
		http.Error(w, "invalid ML prediction response", http.StatusBadGateway)
		return
	}

	userIDString, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok || userIDString == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		http.Error(w, "invalid user identity", http.StatusUnauthorized)
		return
	}

	predictionJSON, err := json.Marshal(prediction)
	if err != nil {
		http.Error(w, "failed to encode prediction", http.StatusInternalServerError)
		return
	}

	inputJSON, err := json.Marshal(req.InputData)
	if err != nil {
		http.Error(w, "failed to encode input data", http.StatusInternalServerError)
		return
	}

	var confidence *float64

	if value, ok := prediction["confidence"].(float64); ok {
		confidence = &value
	}

	_, err = h.Models.SavePredictionHistory(
		r.Context(),
		models.SavePredictionHistoryParams{
			UserID:         userID,
			ModelID:        modelID,
			PredictionType: predictionType(prediction),
			InputData:      inputJSON,
			Prediction:     predictionJSON,
			Confidence:     confidence,
		},
	)

	if err != nil {
		http.Error(
			w,
			fmt.Sprintf("prediction succeeded but history save failed: %v", err),
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prediction)
}

func predictionType(prediction map[string]interface{}) string {
	if taskType, ok := prediction["task_type"].(string); ok {
		return taskType
	}

	return "prediction"
}
