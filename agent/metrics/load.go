package metrics

import (
	"runtime"

	"github.com/shirou/gopsutil/v3/load"
)

func LoadAverage() (float64, float64, float64) {
	if runtime.GOOS == "windows" {
		return 0, 0, 0
	}

	l, err := load.Avg()
	if err != nil {
		return 0, 0, 0
	}

	return l.Load1, l.Load5, l.Load15
}
