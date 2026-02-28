package collector

import (
	"encoding/json"
	"strings"
)

func ExtractLevel(line string) string {

	// 1️⃣ JSON structured logs
	if strings.Contains(line, "\"level\"") {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(line), &obj); err == nil {
			if lvl, ok := obj["level"].(string); ok {
				return normalizeLevel(lvl)
			}
		}
	}

	// 2️⃣ Plain text logs (timestamp LEVEL message)
	fields := strings.Fields(line)
	if len(fields) >= 3 {
		return normalizeLevel(fields[2])
	}

	// 3️⃣ Keyword inference
	lower := strings.ToLower(line)

	if strings.Contains(lower, "panic") ||
		strings.Contains(lower, "fatal") ||
		strings.Contains(lower, "critical") {
		return "critical"
	}

	if strings.Contains(lower, "error") ||
		strings.Contains(lower, "failed") ||
		strings.Contains(lower, "exception") {
		return "error"
	}

	if strings.Contains(lower, "warn") {
		return "warn"
	}

	return "info"
}

func normalizeLevel(lvl string) string {
	l := strings.ToLower(lvl)

	switch l {
	case "info":
		return "info"
	case "warn", "warning":
		return "warn"
	case "error":
		return "error"
	case "critical", "fatal", "panic":
		return "critical"
	case "debug":
		return "debug"
	default:
		return "info"
	}
}
