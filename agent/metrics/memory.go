package metrics

import "github.com/shirou/gopsutil/v3/mem"

func Memory() (uint64, uint64) {
	m, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0
	}
	return m.Total / 1024 / 1024, m.Used / 1024 / 1024
}
