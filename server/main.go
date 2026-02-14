package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"

	"opslense-pulse/server/api"
	"opslense-pulse/server/config"
	"opslense-pulse/server/store"
	"opslense-pulse/shared"
)

const ServerVersion = "1.0.0"

// -------------------------------------------
// Helpers
// -------------------------------------------

// Print CLI help
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

// Generate a random API key
func generateAPIKey() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return "opl_" + hex.EncodeToString(b)
}

// Hash API key (SHA256)
func hashAPIKey(rawKey string) string {
	sum := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(sum[:])
}

// Bootstrap API key if DB is empty
func bootstrapAPIKeyIfNeeded(s store.Store) error {
	count, err := s.CountAPIKeys()
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	rawKey := generateAPIKey()
	hash := hashAPIKey(rawKey)

	err = s.InsertAPIKey(shared.APIKey{
		AccountID:   "default",
		KeyID:       uuid.NewString(),
		KeyHash:     hash,
		Name:        "bootstrap-admin",
		IsActive:    1,
		IsBootstrap: 1,
		CreatedAt:   time.Now(),
	})
	if err != nil {
		return err
	}

	fmt.Println("========================================")
	fmt.Println(" OpsLens Pulse – Bootstrap API Key")
	fmt.Println("========================================")
	fmt.Println(" THIS KEY IS SHOWN ONLY ONCE")
	fmt.Println()
	fmt.Println(" API KEY:")
	fmt.Println(" ", rawKey)
	fmt.Println()
	fmt.Println(" Store it securely. It cannot be recovered.")
	fmt.Println("========================================")

	return nil
}

// Auth wrapper for API handlers
func makeAuthHandler(st store.Store, handler func(store.Store) http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		rawKey := strings.TrimPrefix(header, "Bearer ")
		ok, err := st.ValidateAPIKey(rawKey)
		if err != nil || !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "account_id", "default")
		handler(st)(w, r.WithContext(ctx))
	}
}

// -------------------------------------------
// Main
// -------------------------------------------

func main() {
	shared.InitLogger("server")
	log.Println("🚀 OpsLens-Pulse Server starting...")

	// CLI flags
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

	// Load server config
	cfg, path, created, err := config.LoadOrCreate(configPath)
	if err != nil {
		log.Fatal(err)
	}
	if created {
		log.Printf("📄 Server config created at: %s\n", path)
	} else {
		log.Printf("📄 Server config loaded from: %s\n", path)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid server config: %v", err)
	}

	// -------------------------------
	// Connect to ClickHouse or fallback to MemoryStore
	// -------------------------------
	var st store.Store
	chcfg := config.LoadClickHouse()
	chStore, err := store.NewClickHouseStore(chcfg.DSN())
	if err != nil {
		//log.Println("ClickHouse unavailable, falling back to MemoryStore:", err)
		// st = store.NewMemoryStore()
		log.Fatalf("❌ ClickHouse connection FAILED: %v", err)
	} else {
		log.Println("✅ Connected to ClickHouse")
		st = chStore
	}

	// Bootstrap API key
	if err := bootstrapAPIKeyIfNeeded(st); err != nil {
		log.Fatalf("Failed to bootstrap API key: %v", err)
	}

	// -------------------------------
	// Serve React (Vite production build)
	// -------------------------------

	uiPath := "./ui/dist"

	// Serve static assets directly
	http.Handle("/assets/",
		http.StripPrefix("/assets/",
			http.FileServer(http.Dir(filepath.Join(uiPath, "assets"))),
		),
	)

	// Serve favicon / vite.svg if needed
	http.Handle("/vite.svg",
		http.FileServer(http.Dir(uiPath)),
	)

	// SPA fallback for everything else (but not /api)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		// If this is an API route, let registered handlers process it
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		http.ServeFile(w, r, filepath.Join(uiPath, "index.html"))
	})

	// -------------------------------
	// API routes - WITH AUTHENTICATION
	// -------------------------------
	http.HandleFunc("/api/metrics", makeAuthHandler(st, api.MetricsHandler))
	http.HandleFunc("/api/heartbeat", makeAuthHandler(st, api.HeartbeatHandler))
	http.HandleFunc("/api/hosts", makeAuthHandler(st, api.HostsHandler))
	http.HandleFunc("/api/hosts/summary", makeAuthHandler(st, api.HostSummaryHandler))
	http.HandleFunc("/api/logs", makeAuthHandler(st, api.LogsHandler))
	http.HandleFunc("/api/logs/fetch", makeAuthHandler(st, api.FetchLogsHandler))
	http.HandleFunc("/api/container/metrics", makeAuthHandler(st, api.ContainerMetricsHandler))
	http.HandleFunc("/api/container/logs", makeAuthHandler(st, api.ContainerLogsHandler))
	http.HandleFunc("/api/logs/sources", makeAuthHandler(st, api.LogSourcesHandler))

	// -------------------------------
	// Start server
	// -------------------------------
	addr := fmt.Sprintf(":%d", cfg.ListenPort)
	log.Println("Server listening on", addr)

	srv := &http.Server{
		Addr:           addr,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// -------------------------------
	// Graceful shutdown
	// -------------------------------
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("Server stopped")
}
