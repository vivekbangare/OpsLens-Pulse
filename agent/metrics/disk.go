package metrics

import (
	"math"
	"runtime"
	"sync"

	"opslense-pulse/shared"

	"github.com/shirou/gopsutil/v3/disk"
)

const (
	criticalThreshold = 90.0 // 90% usage
	spikeThreshold    = 10.0 // 10% sudden jump
)

var (
	prevDiskUsage = make(map[string]float64)
	diskMu        sync.Mutex
)

func DiskUsageAll() []shared.DiskMetric {

	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil
	}

	var results []shared.DiskMetric

	for _, p := range partitions {

		// Skip pseudo filesystems (Linux)
		if runtime.GOOS == "linux" {
			if p.Fstype == "tmpfs" ||
				p.Fstype == "devtmpfs" ||
				p.Fstype == "overlay" {
				continue
			}
		}

		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}
		diskMu.Lock()
		currentUsedPct := usage.UsedPercent
		prevUsedPct := prevDiskUsage[p.Mountpoint]

		// Critical detection
		critical := currentUsedPct >= criticalThreshold

		// Spike detection
		spike := false
		if prevUsedPct > 0 {
			delta := math.Abs(currentUsedPct - prevUsedPct)
			if delta >= spikeThreshold {
				spike = true
			}
		}

		// Save for next comparison
		prevDiskUsage[p.Mountpoint] = currentUsedPct
		diskMu.Unlock()
		results = append(results, shared.DiskMetric{
			MountPoint:    p.Mountpoint,
			FSType:        p.Fstype,
			TotalMB:       float32(usage.Total) / 1024 / 1024,
			UsedMB:        float32(usage.Used) / 1024 / 1024,
			UsedPct:       float32(currentUsedPct),
			SpikeDetected: spike,
			Critical:      critical,
		})
	}

	return results
}
