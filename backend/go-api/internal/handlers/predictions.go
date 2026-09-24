package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type PredictionHandler struct {
	MLServiceURL string
}

func NewPredictionHandler(mlServiceURL string) *PredictionHandler {
	return &PredictionHandler{
		MLServiceURL: mlServiceURL,
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

	if _, err := io.Copy(w, response.Body); err != nil {
		return
	}
}
