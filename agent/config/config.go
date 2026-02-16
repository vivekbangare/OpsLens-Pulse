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
	URL    string `yaml:"url"`
	APIKey string `yaml:"api_key"`
}

type LogSource struct {
	Name    string   `yaml:"name"`
	Type    string   `yaml:"type"` // file | directory | command
	Path    string   `yaml:"path,omitempty"`
	Command string   `yaml:"command,omitempty"`
	Include []string `yaml:"include,omitempty"`
	Exclude []string `yaml:"exclude,omitempty"`
}

type AgentConfig struct {
	IntervalSeconds              int         `yaml:"interval_seconds"`
	SelfUpgrade                  bool        `yaml:"self_upgrade"`
	Logs                         []LogSource `yaml:"logs"`
	LogCollectionIntervalSeconds int         `yaml:"log_collection_interval_seconds"`
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

	// Logs are OPTIONAL – defaults may be injected later
	if len(c.Agent.Logs) > 0 {
		if c.Agent.LogCollectionIntervalSeconds <= 0 {
			return errors.New("agent.log_collection_interval_seconds must be > 0")
		}

		for i, src := range c.Agent.Logs {
			if src.Name == "" {
				return fmt.Errorf("agent.logs[%d].name is required", i)
			}
			if src.Type == "" {
				return fmt.Errorf("agent.logs[%d].type is required", i)
			}

			switch src.Type {
			case "file":
				if src.Path == "" {
					return fmt.Errorf("agent.logs[%d].path is required for type=file", i)
				}
			case "directory":
				if src.Path == "" {
					return fmt.Errorf("agent.logs[%d].path is required for type=directory", i)
				}
			case "command":
				if src.Command == "" {
					return fmt.Errorf("agent.logs[%d].command is required for type=command", i)
				}
			default:
				return fmt.Errorf(
					"agent.logs[%d].type must be one of: file, directory, command",
					i,
				)
			}
		}
	}

	return nil
}

func defaultConfig() Config {
	return Config{
		Server: ServerConfig{
			URL:    "http://localhost:9898",
			APIKey: "",
		},
		Agent: AgentConfig{
			IntervalSeconds:              5,
			SelfUpgrade:                  false,
			Logs:                         nil,
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
