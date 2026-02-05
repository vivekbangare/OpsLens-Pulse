package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"strings"

	"opslense-pulse/server/api"
	"opslense-pulse/server/config"
	"opslense-pulse/shared"
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
			w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}
		next(w, r)
	}
}

func main() {
	shared.InitLogger("server")
	log.Println("🚀 OpsLens-Pulse Server starting...")

	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to config file")
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

	// Load config first
	cfg, path, created, err := config.LoadOrCreate(configPath)
	if err != nil {
		log.Fatal(err)
	}
	if created {
		log.Printf("📄 Server config created at: %s\n", path)
	} else {
		log.Printf("📄 Server config loaded from: %s\n", path)
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid server config: %v", err)
	}

	token := cfg.Token
	if token == "" {
		log.Fatal("Server token must be set in config file")
	}
	api.SetAuthToken(token)

	// 1. Serve static assets (CSS, JS)
	fs := http.FileServer(http.Dir("./server/ui/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 2. Serve HTML pages and hide .html in URLs
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		} else if !strings.HasSuffix(path, ".html") {
			path = path + ".html"
		}
		http.ServeFile(w, r, "./server/ui/static"+path)
	})

	// API endpoints with auth
	http.HandleFunc("/api/heartbeat", authMiddleware(api.HeartbeatHandler))
	http.HandleFunc("/api/logs", authMiddleware(api.LogsHandler))
	http.HandleFunc("/api/metrics", authMiddleware(api.MetricsHandler))
	http.HandleFunc("/api/hosts", authMiddleware(api.HostsHandler))

	addr := fmt.Sprintf(":%d", cfg.ListenPort)
	log.Println("Server listening on", addr)

	srv := &http.Server{Addr: addr}

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
	log.Println("Server stopped")
}
