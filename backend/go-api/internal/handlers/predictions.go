package handlers

import (
	"encoding/json"
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