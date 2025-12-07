package http

import (
	"encoding/json"
	"log"
	"net/http"
)

func EncodeResponse(w http.ResponseWriter, message string, statusCode int, errCode ErrorCode) {
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(ErrorResponse{
		Code:    errCode,
		Message: message,
	})
	if err != nil {
		log.Println("EncodeResponse error: %w", err)
	}
}

func SwaggerErrorHandlerFunc(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if statusCode == http.StatusBadRequest {
		EncodeResponse(w, message, statusCode, BadRequest)
		return
	}
	EncodeResponse(w, message, http.StatusInternalServerError, InternalServerError)
	return
}

func RequestErrorHandlerFunc(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	EncodeResponse(w, err.Error(), http.StatusBadRequest, BadRequest)
}

func ResponseErrorHandlerFunc(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	EncodeResponse(w, err.Error(), http.StatusInternalServerError, BadRequest)
}
