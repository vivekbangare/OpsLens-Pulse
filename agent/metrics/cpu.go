// agent/metrics/cpu.go
package metrics

import "github.com/shirou/gopsutil/v3/cpu"

func CPUPercent() float64 {
	p, err := cpu.Percent(0, false)
	if err != nil || len(p) == 0 {
		return 0
	}
	return p[0]
}
