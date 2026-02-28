package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func (c Config) Validate() error {

	if c.Version != "" && c.Version != "1" {
		return fmt.Errorf("unsupported config version: %s", c.Version)
	}

	// ---------------- Server ----------------

	if c.Server.URL == "" {
		return errors.New("server.url must be set")
	}

	parsed, err := url.Parse(c.Server.URL)
	if err != nil {
		return fmt.Errorf("invalid server.url: %v", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("server.url must start with http:// or https://")
	}

	if parsed.Host == "" {
		return errors.New("server.url must contain host")
	}

	if c.Server.APIKey == "" || len(c.Server.APIKey) < 16 {
		return errors.New("server.api_key must be set and >= 16 characters")
	}

	// ---------------- Agent ----------------

	if c.Agent.IntervalSeconds < 5 || c.Agent.IntervalSeconds > 3600 {
		return errors.New("agent.interval_seconds must be between 5 and 3600")
	}

	if len(c.Agent.Logs) > 0 {

		if c.Agent.LogCollectionIntervalSeconds < 5 ||
			c.Agent.LogCollectionIntervalSeconds > 3600 {
			return errors.New("agent.log_collection_interval_seconds must be between 5 and 3600")
		}

		seen := make(map[string]bool)

		for i, src := range c.Agent.Logs {

			if src.Name == "" {
				return fmt.Errorf("agent.logs[%d].name is required", i)
			}
			if seen[src.Name] {
				return fmt.Errorf("duplicate log source name: %s", src.Name)
			}
			seen[src.Name] = true

			switch src.Type {

			case "file":
				if src.Path == "" {
					return fmt.Errorf("agent.logs[%d].path is required for type=file", i)
				}
				if info, err := os.Stat(src.Path); err == nil && info.IsDir() {
					return fmt.Errorf("agent.logs[%d].path must be a file", i)
				}

			case "directory":
				if src.Path == "" {
					return fmt.Errorf("agent.logs[%d].path is required for type=directory", i)
				}
				info, err := os.Stat(src.Path)
				if err != nil || !info.IsDir() {
					return fmt.Errorf("agent.logs[%d].path must be a valid directory", i)
				}

			case "journal", "dmesg":
				if runtime.GOOS != "linux" {
					return fmt.Errorf("%s supported only on Linux", src.Type)
				}

			case "windows-event":
				if runtime.GOOS != "windows" {
					return fmt.Errorf("windows-event supported only on Windows")
				}
				if src.Channel == "" {
					return fmt.Errorf("agent.logs[%d].channel is required", i)
				}

			default:
				return fmt.Errorf(
					"agent.logs[%d].type must be file, directory, journal, dmesg, or windows-event",
					i,
				)
			}
		}

		if err := validateLogSourcesAdvanced(c.Agent.Logs); err != nil {
			return err
		}
	}

	for k := range c.Tags {
		if strings.TrimSpace(k) == "" {
			return errors.New("tags cannot contain empty keys")
		}
	}

	return nil
}

// ---------------- Advanced Validation ----------------

func validateLogSourcesAdvanced(sources []LogSource) error {

	absPaths := make(map[string]string)
	var dirPaths []string

	for _, src := range sources {

		if src.Path == "" {
			continue
		}

		abs, err := filepath.Abs(src.Path)
		if err != nil {
			continue
		}

		if prevType, exists := absPaths[abs]; exists {
			return fmt.Errorf("duplicate log path detected: %s used for both %s and %s",
				abs, prevType, src.Type)
		}
		absPaths[abs] = src.Type

		if src.Type == "directory" {
			dirPaths = append(dirPaths, abs)
		}

		for _, inc := range src.Include {
			for _, exc := range src.Exclude {
				if inc == exc {
					return fmt.Errorf("conflicting include/exclude pattern '%s' in %s",
						inc, abs)
				}
			}
		}
	}

	for path, typ := range absPaths {
		if typ != "file" {
			continue
		}
		for _, dir := range dirPaths {
			if isSubPath(dir, path) {
				return fmt.Errorf("log source overlap: file %s inside directory %s",
					path, dir)
			}
		}
	}

	for i := 0; i < len(dirPaths); i++ {
		for j := i + 1; j < len(dirPaths); j++ {
			if isSubPath(dirPaths[i], dirPaths[j]) ||
				isSubPath(dirPaths[j], dirPaths[i]) {
				return fmt.Errorf("nested directory log sources: %s and %s",
					dirPaths[i], dirPaths[j])
			}
		}
	}

	return nil
}

func isSubPath(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, "..")
}
