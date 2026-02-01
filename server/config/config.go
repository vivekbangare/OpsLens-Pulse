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
			Token:      "",
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
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return yaml.NewEncoder(f).Encode(cfg)
}
