package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	URL   string `yaml:"url"`
	APIKey string `yaml:"api_key"`
}

type AgentConfig struct {
	IntervalSeconds int  `yaml:"interval_seconds"`
	SelfUpgrade     bool `yaml:"self_upgrade"`
}

type Config struct {
	Server ServerConfig      `yaml:"server"`
	Agent  AgentConfig       `yaml:"agent"`
	Tags   map[string]string `yaml:"tags"`
}

var DefaultPath string

func init() {
	if runtime.GOOS == "windows" {
		DefaultPath = `C:\ProgramData\OpsLens-Pulse\agent-config.yaml`
	} else {
		DefaultPath = "/etc/opslens-pulse/agent-config.yaml"
	}
}

// LoadOrCreateConfig returns: Config, path, created(bool), error
func LoadOrCreateConfig(path string) (Config, string, bool, error) {
	var cfg Config
	created := false

	// Allow env ONLY to override config path
	if env := os.Getenv("OPS_AGENT_CONFIG"); env != "" {
		path = env
	}
	if path == "" {
		path = DefaultPath
	}

	// Create default config if missing
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg = defaultConfig()

		if err := save(path, cfg); err != nil {
			return cfg, path, false, err
		}

		created = true
		return cfg, path, created, nil
	}

	// Load existing config
	f, err := os.Open(path)
	if err != nil {
		return cfg, path, false, err
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return cfg, path, false, err
	}

	return cfg, path, false, nil
}

func defaultConfig() Config {
	return Config{
		Server: ServerConfig{
			URL:   "http://localhost:9898",
			APIKey: "",
		},
		Agent: AgentConfig{
			IntervalSeconds: 5,
			SelfUpgrade:     false,
		},
		Tags: map[string]string{
			"env": "dev",
		},
	}
}

func save(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		return err
	}
	defer f.Close()

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	return enc.Encode(cfg)
}

func (c Config) Validate() error {
	if c.Server.URL == "" {
		return errors.New("server.url must be set")
	}
	if c.Server.APIKey == "" {
		return errors.New("server.api_key must be set")
	}
	if c.Agent.IntervalSeconds <= 0 {
		return errors.New("agent.interval_seconds must be > 0")
	}
	return nil
}
