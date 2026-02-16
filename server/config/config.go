package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	ListenPort int `yaml:"listen_port"`
}

type ClickHouseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

func LoadClickHouse() ClickHouseConfig {
	cfg := ClickHouseConfig{
		Host:     os.Getenv("CLICKHOUSE_HOST"),
		Port:     9000,
		User:     os.Getenv("CLICKHOUSE_USER"),
		Password: os.Getenv("CLICKHOUSE_PASSWORD"),
		Database: os.Getenv("CLICKHOUSE_DATABASE"),
	}
	if port := os.Getenv("CLICKHOUSE_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Port)
	}
	return cfg
}

func (c ClickHouseConfig) DSN() string {
	if c.Database == "" {
		c.Database = "default"
	}
	return fmt.Sprintf(
		"tcp://%s:%d/%s?username=%s&password=%s",
		c.Host,
		c.Port,
		c.Database,
		c.User,
		c.Password,
	)
}

var DefaultPath string

func init() {
	if runtime.GOOS == "windows" {
		DefaultPath = `C:\ProgramData\OpsLens-Pulse\server-config.yaml`
	} else {
		DefaultPath = "/etc/opslens-pulse/server-config.yaml"
	}
}

// LoadOrCreate returns: ServerConfig, path, created(bool), error
func LoadOrCreate(path string) (ServerConfig, string, bool, error) {
	var cfg ServerConfig

	if env := os.Getenv("OPS_SERVER_CONFIG"); env != "" {
		path = env
	}
	if path == "" {
		path = DefaultPath
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// create default config
		cfg = ServerConfig{
			ListenPort: 9898,
		}
		if err := save(path, cfg); err != nil {
			return cfg, path, false, err
		}

		return cfg, path, true, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return cfg, path, false, err
	}
	defer f.Close()

	err = yaml.NewDecoder(f).Decode(&cfg)
	return cfg, path, false, err
}

func save(path string, cfg ServerConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		return err
	}
	defer f.Close()

	return yaml.NewEncoder(f).Encode(cfg)
}

func (c ServerConfig) Validate() error {
	if c.ListenPort <= 0 || c.ListenPort > 65535 {
		return errors.New("listen_port must be between 1 and 65535")
	}
	return nil
}
