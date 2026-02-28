package metrics

import (
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
)

type DiskIORate struct {
	ReadBytesPerSec  float64
	WriteBytesPerSec float64
}

var (
	ioMu      sync.Mutex
	prevRead  uint64
	prevWrite uint64
	prevTime  time.Time
)

func DiskIO() DiskIORate {
	ioMu.Lock()
	defer ioMu.Unlock()

	stats, err := disk.IOCounters()
	if err != nil {
		return DiskIORate{}
	}

	var totalRead uint64
	var totalWrite uint64

	for _, v := range stats {
		totalRead += v.ReadBytes
		totalWrite += v.WriteBytes
	}

	now := time.Now()

	if prevTime.IsZero() {
		prevRead = totalRead
		prevWrite = totalWrite
		prevTime = now
		return DiskIORate{}
	}

	elapsed := now.Sub(prevTime).Seconds()
	if elapsed <= 0 {
		return DiskIORate{}
	}

	readRate := float64(totalRead-prevRead) / elapsed
	writeRate := float64(totalWrite-prevWrite) / elapsed

	prevRead = totalRead
	prevWrite = totalWrite
	prevTime = now

	return DiskIORate{
		ReadBytesPerSec:  readRate,
		WriteBytesPerSec: writeRate,
	}
}
