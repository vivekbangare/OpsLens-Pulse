package shared

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

var componentName string

func InitLogger(component string) {
	componentName = component
	log.SetOutput(os.Stdout)
	log.SetFlags(0)
}

func logWithLevel(level string, msg string, kv ...interface{}) {

	entry := map[string]interface{}{
		"time":      time.Now().Format(time.RFC3339),
		"level":     level,
		"component": componentName,
		"message":   msg,
	}

	// Add structured fields
	for i := 0; i+1 < len(kv); i += 2 {
		key, ok := kv[i].(string)
		if !ok {
			continue
		}
		entry[key] = kv[i+1]
	}

	b, err := json.Marshal(entry)
	if err != nil {
		log.Println("failed to marshal log entry:", err)
		return
	}

	// 🔥 DO NOT use server/logger
	log.Println(string(b))
}

func Info(msg string, kv ...interface{}) {
	logWithLevel("info", msg, kv...)
}

func Warn(msg string, kv ...interface{}) {
	logWithLevel("warn", msg, kv...)
}

func Error(msg string, kv ...interface{}) {
	logWithLevel("error", msg, kv...)
}
