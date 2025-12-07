package http

import (
	"encoding/json"
	"net/http"
)

func SwaggerErrorHandlerFunc(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(ErrorResponse{
		Code:    "test",
		Message: message,
	})
}

func RequestErrorHandlerFunc(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Code:    "test1",
		Message: err.Error(),
	})
}

func ResponseErrorHandlerFunc(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Code:    "test",
		Message: err.Error(),
	})
}
