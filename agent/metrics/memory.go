// agent/metrics/memory.go
package metrics

import "github.com/shirou/gopsutil/v3/mem"

func Memory() (uint64, uint64) {
	m, _ := mem.VirtualMemory()
	return m.Total / 1024 / 1024, m.Used / 1024 / 1024
}
