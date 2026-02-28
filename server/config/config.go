package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultPath = "/etc/opslens-pulse/server-config.yaml"

type Config struct {
	Server     ServerSection     `yaml:"server"`
	ClickHouse ClickHouseSection `yaml:"clickhouse"`
	Postgres   PostgresSection   `yaml:"postgres"`
	Security   SecuritySection   `yaml:"security"`
	Logging    LoggingSection    `yaml:"logging"`
}

type ServerSection struct {
	ListenPort      int `yaml:"listen_port"`
	ReadTimeoutSec  int `yaml:"read_timeout_sec"`
	WriteTimeoutSec int `yaml:"write_timeout_sec"`
	IdleTimeoutSec  int `yaml:"idle_timeout_sec"`
	MaxHeaderBytes  int `yaml:"max_header_bytes"`

	RateLimit struct {
		Requests  int `yaml:"requests"`
		WindowSec int `yaml:"window_sec"`
	} `yaml:"rate_limit"`
}

type ClickHouseSection struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Database string `yaml:"database"`

	Password string `yaml:"-"`

	// Production tuning
	MaxOpenConns       int `yaml:"max_open_conns"`
	MaxIdleConns       int `yaml:"max_idle_conns"`
	ConnMaxLifetimeMin int `yaml:"conn_max_lifetime_min"`

	ConnectRetries int `yaml:"connect_retries"`
	RetryDelaySec  int `yaml:"retry_delay_sec"`
}

type PostgresSection struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Database string `yaml:"database"`

	// Not from YAML
	Password string `yaml:"-"`

	// Production settings
	SSLMode            string `yaml:"ssl_mode"` // disable | require | verify-full
	MaxOpenConns       int    `yaml:"max_open_conns"`
	MaxIdleConns       int    `yaml:"max_idle_conns"`
	ConnMaxLifetimeMin int    `yaml:"conn_max_lifetime_min"`
	ConnectRetries     int    `yaml:"connect_retries"`
	RetryDelaySec      int    `yaml:"retry_delay_sec"`
}

type SecuritySection struct {
	MaxBodyMB   int    `yaml:"max_body_mb"`
	JWTSecret   string `yaml:"-"`
	JWTIssuer   string `yaml:"jwt_issuer"`
	JWTAudience string `yaml:"jwt_audience"`
}

type LoggingSection struct {
	Level string `yaml:"level"`
}

func resolveSecret(envKey string, defaultFile string) (string, error) {

	// 1️⃣ ENV override (Docker / K8s)
	if v := os.Getenv(envKey); v != "" {
		return v, nil
	}

	// 2️⃣ Secret file (On-Prem / Mounted secret)
	if _, err := os.Stat(defaultFile); err == nil {
		b, err := os.ReadFile(defaultFile)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(b)), nil
	}

	return "", fmt.Errorf("%s not configured", envKey)
}

func Load(path string) (*Config, error) {

	if path == "" {
		if env := os.Getenv("OPS_SERVER_CONFIG"); env != "" {
			path = env
		} else {
			path = DefaultPath
		}
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg := &Config{}
	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, err
	}

	// 🔐 Resolve Secrets

	cfg.Security.JWTSecret, err = resolveSecret(
		"JWT_SECRET",
		"/etc/opslens-pulse/secrets/jwt.secret",
	)
	if err != nil {
		return nil, err
	}

	cfg.Postgres.Password, err = resolveSecret(
		"POSTGRES_PASSWORD",
		"/etc/opslens-pulse/secrets/postgres.secret",
	)
	if err != nil {
		return nil, err
	}

	cfg.ClickHouse.Password, err = resolveSecret(
		"CLICKHOUSE_PASSWORD",
		"/etc/opslens-pulse/secrets/clickhouse.secret",
	)
	if err != nil {
		return nil, err
	}

	return cfg, cfg.Validate()
}

func (c *Config) Validate() error {

	if c.Server.ListenPort <= 0 || c.Server.ListenPort > 65535 {
		return errors.New("invalid listen_port")
	}

	if c.Security.JWTSecret == "" {
		return errors.New("JWT secret missing")
	}

	if c.Security.JWTIssuer == "" {
		return errors.New("jwt_issuer missing")
	}

	if c.Security.JWTAudience == "" {
		return errors.New("jwt_audience missing")
	}

	if c.Postgres.Host == "" {
		return errors.New("postgres host missing")
	}

	if c.ClickHouse.Host == "" {
		return errors.New("clickhouse host missing")
	}

	return nil
}
