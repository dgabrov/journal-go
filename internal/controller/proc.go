package controller

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

const cookieName = "jou12"

type ErrNoAuth struct{}

func (e *ErrNoAuth) Error() string {
	return "no authentication token found"
}

type Success struct {
	Success string `json:"success"`
}

func NewSuccess() *Success {
	return &Success{Success: "success"}
}

type ErrorResponse struct {
	Message string   `json:"message"`
	Items   []string `json:"items"`
}

func writeJsonResponse(writer http.ResponseWriter, payload any, err error) {
	status := http.StatusOK
	processPayload := payload

	if err != nil {
		processPayload = ErrorResponse{
			Message: err.Error(),
			Items:   make([]string, 0),
		}

		status = http.StatusBadRequest
	}

	header := writer.Header()
	header.Set("Content-Type", "application/json")
	header.Set("Cache-Control", "no-cache")

	writer.WriteHeader(status)

	// write the actual object here
	bts, err := json.Marshal(processPayload)
	if err != nil {
		slog.Error("error marshalling payload")
	}

	_, err = writer.Write(bts)
	if err != nil {
		slog.Error("error writing payload")
	}
}

func parseJsonRequest(r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to parse request body: %w", err)
	}
	return nil
}

func getToken(r *http.Request) (string, error) {
	// Check Authorization header first
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		lower := strings.ToLower(authHeader)
		if strings.HasPrefix(lower, "bearer ") {
			return authHeader[7:], nil
		}

		return authHeader, nil
	}

	// Check for cookie if Authorization header not found
	cookie, err := r.Cookie(cookieName)
	if err == nil {
		return cookie.Value, nil
	}

	// No token found
	return "", &ErrNoAuth{}
}
