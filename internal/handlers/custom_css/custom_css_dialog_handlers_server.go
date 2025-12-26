//go:build server

package custom_css

import (
	"encoding/json"
	"log"
	"net/http"

	"MrRSS/internal/handlers/core"
)

// HandleUploadCSSDialog is not available in server mode.
func HandleUploadCSSDialog(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	log.Printf("File dialog operations are not available in server mode")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": "File dialog operations are not available in server mode. Use web upload instead.",
	})
}
