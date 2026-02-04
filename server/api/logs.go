package api

import (
	"net/http"
	"os"
	"path/filepath"
)

const logDir = "/var/log/opslens-pulse"

// LogsHandler is an HTTP handler for fetching logs
func LogsHandler(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")
	if file == "" {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}

	// Prevent directory traversal attacks
	path := filepath.Join(logDir, filepath.Base(file))

	content, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "cannot read log", http.StatusInternalServerError)
		return
	}

	w.Write(content)
}
