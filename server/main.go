package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"opslense-pulse/server/api"
	"opslense-pulse/server/config"
)

const ServerVersion = "1.0.0"

func printHelp() {
	fmt.Print(`
OpsLens-Pulse Server

Usage:
  opslens-pulse-server [options]

Options:
  --config <path>     Path to config file
  --version           Show version
  --help              Show help

Defaults:
  Linux:   /etc/opslens-pulse/server-config.yaml
  Windows: C:\ProgramData\OpsLens-Pulse\server-config.yaml

Env:
  OPS_SERVER_CONFIG
`)
}

// authMiddleware ensures requests have the correct token
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		expected := "Bearer " + api.GetAuthToken()
		if authHeader != expected {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return
		}
		next(w, r)
	}
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "Config path")
	showHelp := flag.Bool("help", false, "Help")
	showVersion := flag.Bool("version", false, "Version")
	flag.Parse()

	if *showHelp {
		printHelp()
		return
	}
	if *showVersion {
		fmt.Println(ServerVersion)
		return
	}

	cfg, err := config.LoadOrCreate(configPath)
	if err != nil {
		log.Fatal(err)
	}

	// Allow override from environment variable
	token := os.Getenv("SERVER_TOKEN")
	if token == "" {
		token = cfg.Token
	}
	api.SetAuthToken(token)

	// Serve static UI files
	http.Handle("/", http.FileServer(http.Dir("./ui/static")))

	// API endpoints with auth
	http.HandleFunc("/api/heartbeat", authMiddleware(api.HeartbeatHandler))
	http.HandleFunc("/api/logs", authMiddleware(api.LogsHandler))
	http.HandleFunc("/api/metrics", authMiddleware(api.MetricsHandler))
	http.HandleFunc("/api/hosts", authMiddleware(api.HostsHandler))

	addr := fmt.Sprintf(":%d", cfg.ListenPort)
	log.Println("Server listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
