package httputil

import (
	"encoding/json"
	"net/http"

	"github.com/Athooh/social-network/pkg/logger"
)

// SendJSON is a helper function to send JSON responses
func SendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// SendError is a helper function to send error responses
func SendError(w http.ResponseWriter, status int, message string, isWarning bool) {
	if isWarning {
		logger.Warn("%s (status: %d)", message, status)
	} else {
		logger.Error("%s (status: %d)", message, status)
	}
	SendJSON(w, status, map[string]string{"error": message})
}
