package config

import (
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	ListenPort int    `yaml:"listen_port"`
	Token      string `yaml:"token"`
}

var DefaultPath string

func init() {
	if runtime.GOOS == "windows" {
		DefaultPath = `C:\ProgramData\OpsLens-Pulse\server-config.yaml`
	} else {
		DefaultPath = "/etc/opslens-pulse/server-config.yaml"
	}
}

func LoadOrCreate(path string) (ServerConfig, error) {
	var cfg ServerConfig

	if env := os.Getenv("OPS_SERVER_CONFIG"); env != "" {
		path = env
	}
	if path == "" {
		path = DefaultPath
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg = ServerConfig{
			ListenPort: 9898,
			Token:      "changeme",
		}
		return cfg, save(path, cfg)
	}

	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	err = yaml.NewDecoder(f).Decode(&cfg)
	return cfg, err
}

func save(path string, cfg ServerConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return yaml.NewEncoder(f).Encode(cfg)
}
