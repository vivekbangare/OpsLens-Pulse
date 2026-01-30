package logs

import (
	"os"
)

func Read(path string, lines int) string {
	b, _ := os.ReadFile(path)
	return string(b)
}
