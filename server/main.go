package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
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
	"opslense-pulse/server/auth"
	"opslense-pulse/server/config"
	"opslense-pulse/server/db"
	"opslense-pulse/server/middleware"
	"opslense-pulse/server/store"
	"opslense-pulse/shared"
)

const ServerVersion = "1.0.0"

// -------------------------------------------
// Helpers
// -------------------------------------------

func printHelp() {
	fmt.Print(`
OpsLens-Pulse Server

Usage:
  opslens-pulse-server [options]

Options:
  --config <path>     Path to config file
  --version           Show version
  --help              Show help
`)
}

func generateAPIKey() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic("crypto rand failed")
	}
	return "opl_" + hex.EncodeToString(b)
}

func hashAPIKey(rawKey string) string {
	sum := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(sum[:])
}

func limitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10MB limit
		next.ServeHTTP(w, r)
	})
}

// Bootstrap API key in POSTGRES (not ClickHouse)
func bootstrapAPIKeyIfNeeded(s *store.PostgresStore, db *sql.DB) error {

	count, err := s.CountAPIKeys()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	// Get default tenant UUID
	var tenantID string
	err = db.QueryRow(`
        SELECT id FROM tenants WHERE slug = 'default-tenant'
    `).Scan(&tenantID)
	if err != nil {
		return err
	}

	rawKey := generateAPIKey()
	hash := hashAPIKey(rawKey)

	err = s.InsertAPIKey(shared.APIKey{
		TenantID: tenantID,
		KeyID:    uuid.NewString(),
		KeyHash:  hash,
		Name:     "bootstrap-admin",
	})
	if err != nil {
		return err
	}

	fmt.Println("========================================")
	fmt.Println(" OpsLens Pulse – Bootstrap API Key")
	fmt.Println("========================================")
	fmt.Println(" THIS KEY IS SHOWN ONLY ONCE")
	fmt.Println(" API KEY:")
	fmt.Println(" ", rawKey)
	fmt.Println("========================================")

	return nil
}

// -------------------------------------------
// Main
// -------------------------------------------

func main() {

	shared.InitLogger("server")
	log.Println("🚀 OpsLens-Pulse Server starting...")

	// ---------------- CLI ----------------
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

	// ---------------- Load Config ----------------
	cfg, path, created, err := config.LoadOrCreate(configPath)
	if err != nil {
		log.Fatal(err)
	}
	if created {
		log.Printf("📄 Config created at: %s\n", path)
	} else {
		log.Printf("📄 Config loaded from: %s\n", path)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	// ---------------- Connect ClickHouse ----------------
	chcfg := config.LoadClickHouse()
	chStore, err := store.NewClickHouseStore(chcfg.DSN())
	if err != nil {
		log.Fatalf("❌ ClickHouse connection FAILED: %v", err)
	}
	log.Println("✅ Connected to ClickHouse")

	// ---------------- Connect Postgres ----------------
	pg, err := db.NewPostgres()
	if err != nil {
		log.Fatalf("❌ Postgres connection FAILED: %v", err)
	}
	log.Println("✅ Connected to Postgres")

	// Postgres store for API keys
	pgStore := store.NewPostgresStore(pg)

	// ---------------- Bootstrap API Key ----------------
	if err := bootstrapAPIKeyIfNeeded(pgStore, pg); err != nil {
		log.Fatalf("Failed to bootstrap API key: %v", err)
	}

	// ---------------- JWT Manager ----------------
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}
	jwtManager := auth.NewJWTManager(jwtSecret)

	// ---------------- Router ----------------
	mux := http.NewServeMux()

	// -------- Agent APIs (API KEY AUTH via Postgres) --------
	mux.Handle("/api/metrics",
		middleware.AgentAuth(pgStore)(api.MetricsHandler(chStore)))

	mux.Handle("/api/heartbeat",
		middleware.AgentAuth(pgStore)(api.HeartbeatHandler(chStore)))

	mux.Handle("/api/logs",
		middleware.AgentAuth(pgStore)(api.LogsHandler(chStore)))

	mux.Handle("/api/container/metrics",
		middleware.AgentAuth(pgStore)(api.ContainerMetricsHandler(chStore)))

	mux.Handle("/api/tenants",
		middleware.UserAuth(jwtManager, pg)(
			api.MyTenantsHandler(pg),
		),
	)
	// -------- User APIs (JWT AUTH) --------
	mux.Handle("/api/hosts",
		middleware.UserAuth(jwtManager, pg)(
			middleware.RequirePermission("hosts.read")(
				api.HostsHandler(chStore),
			),
		),
	)

	mux.Handle("/api/hosts/summary",
		middleware.UserAuth(jwtManager, pg)(
			middleware.RequirePermission("hosts.read")(
				api.HostSummaryHandler(chStore),
			),
		),
	)

	mux.Handle("/api/logs/fetch",
		middleware.UserAuth(jwtManager, pg)(
			middleware.RequirePermission("logs.read")(
				api.FetchLogsHandler(chStore),
			),
		),
	)

	mux.Handle("/api/container/logs",
		middleware.UserAuth(jwtManager, pg)(
			middleware.RequirePermission("logs.read")(
				api.ContainerLogsHandler(chStore),
			),
		),
	)

	// -------- Public --------
	mux.Handle("/api/login", api.LoginHandler(pg, jwtManager))

	// ---------------- Serve React ----------------
	uiPath := "./ui/dist"

	mux.Handle("/assets/",
		http.StripPrefix("/assets/",
			http.FileServer(http.Dir(filepath.Join(uiPath, "assets"))),
		),
	)

	mux.Handle("/vite.svg",
		http.FileServer(http.Dir(uiPath)),
	)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Cache-Control", "no-cache")

		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		path := filepath.Join(uiPath, r.URL.Path)
		if _, err := os.Stat(path); err == nil {
			http.ServeFile(w, r, path)
			return
		}

		http.ServeFile(w, r, filepath.Join(uiPath, "index.html"))
	})

	// ---------------- HTTP Server ----------------
	addr := fmt.Sprintf(":%d", cfg.ListenPort)

	srv := &http.Server{
		Addr:           addr,
		Handler:        limitBody(mux),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		log.Println("Server listening on", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// ---------------- Graceful Shutdown ----------------
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = srv.Shutdown(ctx)

	log.Println("Server stopped")
}
