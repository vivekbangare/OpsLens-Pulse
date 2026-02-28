package metrics

import (
	"math"
	"sync"
)

const (
	cpuCriticalThreshold = 95.0
	cpuSpikeThreshold    = 30.0
	memCriticalThreshold = 90.0
	memPressureThreshold = 80.0
)

var (
	prevCPU float64
	cpuMu   sync.Mutex
)

type CPUAnomaly struct {
	Critical bool
	Spike    bool
}

type MemAnomaly struct {
	Critical bool
	Pressure bool
}

func DetectCPUAnomaly(current float64) CPUAnomaly {
	cpuMu.Lock()
	defer cpuMu.Unlock()

	critical := current >= cpuCriticalThreshold

	spike := false
	if prevCPU > 0 {
		if math.Abs(current-prevCPU) >= cpuSpikeThreshold {
			spike = true
		}
	}

	prevCPU = current

	return CPUAnomaly{
		Critical: critical,
		Spike:    spike,
	}
}

func DetectMemoryAnomaly(usedMB, totalMB uint64) MemAnomaly {

	if totalMB == 0 {
		return MemAnomaly{}
	}

	usedPercent := (float64(usedMB) / float64(totalMB)) * 100

	return MemAnomaly{
		Critical: usedPercent >= memCriticalThreshold,
		Pressure: usedPercent >= memPressureThreshold,
	}
}
