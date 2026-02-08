package shared

import (
	"log"
	"os"
)

var Logger *log.Logger

// InitLogger initializes a global logger
func InitLogger(prefix string) {
	Logger = log.New(os.Stdout, "["+prefix+"] ", log.LstdFlags|log.Lshortfile)
}
