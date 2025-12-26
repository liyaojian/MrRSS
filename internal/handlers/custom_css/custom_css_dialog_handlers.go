//go:build !server

package custom_css

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/utils"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// HandleUploadCSSDialog opens a file dialog to select CSS file for upload.
func HandleUploadCSSDialog(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if h.App == nil {
		log.Printf("File dialog not available")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "File dialog not available",
		})
		return
	}

	// Type assert to *application.App to access Dialog
	app, ok := h.App.(*application.App)
	if !ok {
		log.Printf("File dialog not available: app is not *application.App type")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "File dialog not available",
		})
		return
	}

	filePath, err := app.Dialog.OpenFileWithOptions(&application.OpenFileDialogOptions{
		Title: "Select CSS File",
		Filters: []application.FileFilter{
			{
				DisplayName: "CSS Files (*.css)",
				Pattern:     "*.css",
			},
			{
				DisplayName: "All Files (*)",
				Pattern:     "*",
			},
		},
		CanChooseFiles:       true,
		AllowsOtherFileTypes: true,
	}).PromptForSingleSelection()
	if err != nil {
		log.Printf("Error opening file dialog: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to open file dialog",
		})
		return
	}

	if filePath == "" {
		// User cancelled the dialog
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "cancelled"})
		return
	}

	// Read the selected file
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening selected file: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to open selected file",
		})
		return
	}
	defer file.Close()

	// Get file info for validation
	fileInfo, err := file.Stat()
	if err != nil {
		log.Printf("Error getting file info: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to get file info",
		})
		return
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".css" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Only CSS files are allowed",
		})
		return
	}

	// Validate file size (max 1MB)
	if fileInfo.Size() > 1<<20 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "CSS file is too large (max 1MB)",
		})
		return
	}

	// Get data directory
	dataDir, err := utils.GetDataDir()
	if err != nil {
		log.Printf("Error getting data directory: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to get data directory",
		})
		return
	}

	// Save CSS file
	cssFilePath := filepath.Join(dataDir, customCSSFileName)
	destFile, err := os.Create(cssFilePath)
	if err != nil {
		log.Printf("Error creating CSS file: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to save CSS file",
		})
		return
	}
	defer destFile.Close()

	// Copy file content
	written, err := io.Copy(destFile, file)
	if err != nil {
		log.Printf("Error writing CSS file: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to write CSS file",
		})
		return
	}

	log.Printf("CSS file uploaded via dialog: %s (%d bytes)", filePath, written)

	// Update setting in database
	if err := h.DB.SetSetting("custom_css_file", customCSSFileName); err != nil {
		log.Printf("Error saving custom_css_file setting: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to update settings",
		})
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "CSS file uploaded successfully",
	})
}
