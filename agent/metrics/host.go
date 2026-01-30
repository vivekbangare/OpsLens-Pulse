// agent/metrics/host.go
package metrics

import (
	"github.com/shirou/gopsutil/v3/host"
	"runtime"
)

func HostInfo() (string, uint64) {
	h, _ := host.Info()
	return runtime.GOOS, h.Uptime
}
