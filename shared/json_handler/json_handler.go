package json_handler

import (
	"encoding/json"
	"net/http"

	"github.com/joolshouston/pismo-technical-test/shared/model"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
	}
}

func WriteError(w http.ResponseWriter, error *model.ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(error.Status)

	if err := json.NewEncoder(w).Encode(error); err != nil {
		http.Error(w, `{"error":"failed to encode error response"}`, http.StatusInternalServerError)
	}
}
