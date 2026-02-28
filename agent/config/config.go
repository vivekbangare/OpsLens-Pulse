package config

import (
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	URL                string `yaml:"url"`
	APIKey             string `yaml:"api_key"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
}

type LogSource struct {
	Name    string   `yaml:"name"`
	Type    string   `yaml:"type"` // file | directory | journal | dmesg | windows-event
	Path    string   `yaml:"path,omitempty"`
	Include []string `yaml:"include,omitempty"`
	Exclude []string `yaml:"exclude,omitempty"`
	Channel string   `yaml:"channel,omitempty"`
}

type AgentConfig struct {
	IntervalSeconds              int         `yaml:"interval_seconds"`
	SelfUpgrade                  bool        `yaml:"self_upgrade"`
	Logs                         []LogSource `yaml:"logs"`
	LogCollectionIntervalSeconds int         `yaml:"log_collection_interval_seconds"`
	MonitoredServices            []string    `yaml:"monitored_services"`
}

type Config struct {
	Version string            `yaml:"version"`
	Server  ServerConfig      `yaml:"server"`
	Agent   AgentConfig       `yaml:"agent"`
	Tags    map[string]string `yaml:"tags"`
}

var DefaultPath string

func init() {
	if runtime.GOOS == "windows" {
		DefaultPath = `C:\ProgramData\OpsLens-Pulse\agent-config.yaml`
		return
	}

	if os.Geteuid() == 0 {
		DefaultPath = "/etc/opslens-pulse/agent-config.yaml"
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		DefaultPath = "./agent-config.yaml"
		return
	}

	DefaultPath = filepath.Join(home, ".opslens-pulse", "agent-config.yaml")
}

func LoadOrCreateConfig(path string) (Config, string, bool, error) {
	var cfg Config

	if env := os.Getenv("OPS_AGENT_CONFIG"); env != "" {
		path = env
	}
	if path == "" {
		path = DefaultPath
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg = defaultConfig()
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

	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return cfg, path, false, err
	}

	if cfg.Version == "" {
		cfg.Version = "1"
	}

	return cfg, path, false, nil
}

func defaultConfig() Config {
	return Config{
		Version: "1",
		Server: ServerConfig{
			URL:    "http://localhost:9898",
			APIKey: "",
		},
		Agent: AgentConfig{
			IntervalSeconds:              5,
			LogCollectionIntervalSeconds: 10,
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
