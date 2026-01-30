package config

import (
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

type AgentConfig struct {
	IntervalSeconds int  `yaml:"interval_seconds"`
	SelfUpgrade     bool `yaml:"self_upgrade"`
}

type Config struct {
	Server ServerConfig `yaml:"server"`
	Agent  AgentConfig  `yaml:"agent"`
}

var DefaultPath string

func init() {
	if runtime.GOOS == "windows" {
		DefaultPath = `C:\ProgramData\OpsLens-Pulse\agent-config.yaml`
	} else {
		DefaultPath = "/etc/opslens-pulse/agent-config.yaml"
	}
}

func LoadOrCreateConfig(path string) (Config, error) {
	var cfg Config

	if env := os.Getenv("OPS_AGENT_CONFIG"); env != "" {
		path = env
	}
	if path == "" {
		path = DefaultPath
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg = Config{
			Server: ServerConfig{
				URL:   "http://localhost:9898",
				Token: "changeme",
			},
			Agent: AgentConfig{
				IntervalSeconds: 10,
				SelfUpgrade:     false,
			},
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

func save(path string, cfg Config) error {
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
