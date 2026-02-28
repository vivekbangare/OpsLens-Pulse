package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"opslense-pulse/server/api"
	"opslense-pulse/server/auth"
	"opslense-pulse/server/config"
	"opslense-pulse/server/db"
	"opslense-pulse/server/handlers"
	"opslense-pulse/server/logger"
	"opslense-pulse/server/middleware"
	"opslense-pulse/server/store"
	"opslense-pulse/server/utils"
	"opslense-pulse/shared"
)

const ServerVersion = "1.0.0"

func main() {

	// ---------------- Logger ----------------
	if err := logger.Init(); err != nil {
		panic(err)
	}
	defer logger.Sync()

	logger.Log.Info("server starting",
		zap.String("version", ServerVersion),
	)

	// ---------------- CLI ----------------
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to config file")
	showVersion := flag.Bool("version", false, "Version")
	flag.Parse()

	if *showVersion {
		fmt.Println(ServerVersion)
		return
	}

	// ---------------- Load Config ----------------
	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Log.Fatal("config load failed", zap.Error(err))
	}

	logger.Log.Info("config loaded",
		zap.Int("listen_port", cfg.Server.ListenPort),
	)

	chDB, err := db.NewClickHouse(cfg.ClickHouse)
	if err != nil {
		logger.Log.Fatal("clickhouse connection failed", zap.Error(err))
	}

	chStore := store.NewClickHouseStore(chDB)
	chStore.StartHealthMonitor()
	logger.Log.Info("connected to clickhouse")

	// ---------------- Postgres ----------------
	pg, err := db.NewPostgres(cfg.Postgres)
	if err != nil {
		logger.Log.Fatal("postgres connection failed", zap.Error(err))
	}

	logger.Log.Info("connected to postgres")

	// ---------------- Health Check ----------------
	if err := checkDBHealth("clickhouse", chStore.DB()); err != nil {
		logger.Log.Fatal("clickhouse unhealthy")
	}
	if err := checkDBHealth("postgres", pg); err != nil {
		logger.Log.Fatal("postgres unhealthy")
	}

	logger.Log.Info("startup summary",
		zap.String("version", ServerVersion),
		zap.Int("listen_port", cfg.Server.ListenPort),
		zap.String("postgres_host", cfg.Postgres.Host),
		zap.String("clickhouse_host", cfg.ClickHouse.Host),
		zap.Int("pg_max_open", cfg.Postgres.MaxOpenConns),
		zap.Int("ch_max_open", cfg.ClickHouse.MaxOpenConns),
		zap.Int("rate_limit_requests", cfg.Server.RateLimit.Requests),
		zap.Int("rate_limit_window_sec", cfg.Server.RateLimit.WindowSec),
	)

	// ---------------- Bootstrap API Key ----------------
	pgStore := store.NewPostgresStore(pg)
	if err := bootstrapAPIKeyIfNeeded(pgStore, pg); err != nil {
		logger.Log.Fatal("bootstrap api key failed", zap.Error(err))
	}

	// ---------------- JWT ----------------
	jwtManager := auth.NewJWTManager(
		cfg.Security.JWTSecret,
		cfg.Security.JWTIssuer,
		cfg.Security.JWTAudience,
	)

	// ---------------- Router ----------------
	mux := http.NewServeMux()

	// -------- Agent APIs --------
	mux.Handle("/api/metrics",
		middleware.AgentAuth(pgStore)(api.MetricsHandler(chStore)))

	mux.Handle("/api/heartbeat",
		middleware.AgentAuth(pgStore)(api.HeartbeatHandler(chStore)))

	mux.Handle("/api/logs",
		middleware.AgentAuth(pgStore)(api.LogsHandler(chStore)))

	mux.Handle("/api/container/metrics",
		middleware.AgentAuth(pgStore)(api.ContainerMetricsHandler(chStore)))

	// -------- Audit APIs --------
	mux.Handle("/api/audit",
		middleware.UserAuth(jwtManager, pg)(
			middleware.RequirePermission("audit.read", pgStore)(
				api.AuditLogsHandler(pgStore),
			),
		),
	)

	// -------- User APIs --------
	mux.Handle("/api/tenants",
		middleware.UserAuth(jwtManager, pg)(
			api.MyTenantsHandler(pg),
		),
	)

	mux.Handle("/api/hosts",
		middleware.UserAuth(jwtManager, pg)(
			middleware.RequirePermission("hosts.read", pgStore)(
				api.HostsHandler(chStore),
			),
		),
	)

	mux.Handle("/api/hosts/summary",
		middleware.UserAuth(jwtManager, pg)(
			middleware.RequirePermission("hosts.read", pgStore)(
				api.HostSummaryHandler(chStore),
			),
		),
	)

	mux.Handle("/api/logs/fetch",
		middleware.UserAuth(jwtManager, pg)(
			middleware.RequirePermission("logs.read", pgStore)(
				api.FetchLogsHandler(chStore),
			),
		),
	)

	mux.Handle("/api/container/logs",
		middleware.UserAuth(jwtManager, pg)(
			middleware.RequirePermission("logs.read", pgStore)(
				api.ContainerLogsHandler(chStore),
			),
		),
	)

	logHandler := &handlers.Handler{Store: chStore}

	mux.Handle("/api/logs/search",
		middleware.UserAuth(jwtManager, pg)(
			middleware.RequirePermission("logs.read", pgStore)(
				http.HandlerFunc(logHandler.SearchLogs),
			),
		),
	)

	mux.Handle("/api/logs/timeline",
		middleware.UserAuth(jwtManager, pg)(
			middleware.RequirePermission("logs.read", pgStore)(
				api.LogsTimelineHandler(chStore),
			),
		),
	)

	// -------- Public --------
	mux.Handle("/api/login", api.LoginHandler(pg, jwtManager, pgStore))

	// -------- Health --------
	mux.HandleFunc("/health/live", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "alive",
			"version": ServerVersion,
		})
	})

	mux.HandleFunc("/health/ready", func(w http.ResponseWriter, r *http.Request) {

		status := "ready"
		httpStatus := http.StatusOK

		if err := pg.PingContext(r.Context()); err != nil {
			status = "not_ready"
			httpStatus = http.StatusServiceUnavailable
		}

		if err := chStore.DB().PingContext(r.Context()); err != nil {
			status = "not_ready"
			httpStatus = http.StatusServiceUnavailable
		}

		utils.WriteJSON(w, httpStatus, map[string]interface{}{
			"status":  status,
			"version": ServerVersion,
		})
	})

	// -------- UI --------
	uiPath := "./ui/dist"

	// Serve static assets (JS/CSS)
	mux.Handle("/assets/",
		http.StripPrefix("/assets/",
			http.FileServer(http.Dir(filepath.Join(uiPath, "assets"))),
		),
	)

	// Serve favicon
	mux.Handle("/favicon.ico",
		http.FileServer(http.Dir(uiPath)),
	)

	// SPA fallback
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		// Block API routes
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		filePath := filepath.Join(uiPath, r.URL.Path)

		// If file exists, serve it
		if _, err := os.Stat(filePath); err == nil {
			http.ServeFile(w, r, filePath)
			return
		}

		// Otherwise serve index.html (React Router support)
		http.ServeFile(w, r, filepath.Join(uiPath, "index.html"))
	})

	// ---------------- HTTP Server ----------------
	addr := fmt.Sprintf(":%d", cfg.Server.ListenPort)

	handler := middleware.Recovery(
		middleware.RequestID(
			middleware.RateLimit(
				cfg.Server.RateLimit.Requests,
				time.Duration(cfg.Server.RateLimit.WindowSec)*time.Second,
			)(
				limitBody(mux, cfg.Security.MaxBodyMB),
			),
		),
	)

	srv := &http.Server{
		Addr:           addr,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	go func() {
		logger.Log.Info("server listening", zap.String("address", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("server error", zap.Error(err))
		}
	}()

	// ---------------- Graceful Shutdown ----------------
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Log.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = srv.Shutdown(ctx)
	_ = chDB.Close()
	_ = pg.Close()

	logger.Log.Info("server stopped")
}

func checkDBHealth(name string, db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		logger.Log.Error("database health check failed",
			zap.String("database", name),
			zap.Error(err),
		)
		return err
	}

	logger.Log.Info("database healthy",
		zap.String("database", name),
	)

	return nil
}

func bootstrapAPIKeyIfNeeded(s *store.PostgresStore, db *sql.DB) error {

	count, err := s.CountAPIKeys()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	var tenantID string
	err = db.QueryRow(`SELECT id FROM tenants WHERE slug = 'default-tenant'`).Scan(&tenantID)
	if err != nil {
		return err
	}

	rawKey := generateAPIKey()
	hash := store.HashAPIKey(rawKey)

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
	fmt.Println(" THIS KEY IS SHOWN ONLY ONCE")
	fmt.Println(" API KEY:", rawKey)
	fmt.Println("========================================")

	return nil
}

func generateAPIKey() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return "opl_" + hex.EncodeToString(b)
}

func limitBody(next http.Handler, maxMB int) http.Handler {
	if maxMB <= 0 {
		maxMB = 10
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, int64(maxMB)<<20)
		next.ServeHTTP(w, r)
	})
}
