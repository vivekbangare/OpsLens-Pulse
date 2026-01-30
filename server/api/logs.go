package api

import (
	"net/http"
	"os"
)

// GetLogs fetches logs from a file
func GetLogs(path string, lines int) string {
	b, _ := os.ReadFile(path)
	return string(b)
}

// LogsHandler is an HTTP handler for fetching logs
func LogsHandler(w http.ResponseWriter, r *http.Request) {
	// Example: you can get the log path from query params
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/var/log/app.log" // default path
	}

	content := GetLogs(path, 100) // 100 lines (not implemented yet)
	w.Write([]byte(content))
}
