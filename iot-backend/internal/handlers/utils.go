// internal/handlers/utils.go
package handlers

import (
	"encoding/json"
	"net/http"

	"iot-backend/internal/models"
)

// sendJSON envoie une réponse JSON nettoyée
func sendJSON(w http.ResponseWriter, data interface{}, status int) {
	cleaned := cleanAllStrings(data)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(cleaned)
}

// sendError envoie une réponse d'erreur standardisée
func sendError(w http.ResponseWriter, message string, status int) {
	response := models.ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	}
	sendJSON(w, response, status)
}
