// agent/metrics/host.go
package metrics

import (
	"github.com/shirou/gopsutil/v3/host"
	"runtime"
)

func HostInfo() (string, uint64) {
	h, err := host.Info()
	if err != nil {
		return runtime.GOOS, 0
	}
	return runtime.GOOS, h.Uptime
}
