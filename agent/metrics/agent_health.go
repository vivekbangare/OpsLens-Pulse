package metrics

import (
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

var (
	startTime       = time.Now()
	metricsFailures uint64
	logFailures     uint64
)

type AgentHealth struct {
	CPUPercent      float32 `json:"agent_cpu_percent"`
	MemoryMB        float32 `json:"agent_mem_mb"`
	Goroutines      int     `json:"agent_goroutines"`
	UptimeSec       uint64  `json:"agent_uptime_sec"`
	MetricsFailures uint64  `json:"metrics_send_failures"`
	LogFailures     uint64  `json:"logs_send_failures"`
}

// Call this when metrics sending fails
func IncMetricsFailure() {
	atomic.AddUint64(&metricsFailures, 1)
}

// Call this when logs sending fails
func IncLogFailure() {
	atomic.AddUint64(&logFailures, 1)
}

func GetAgentHealth() AgentHealth {

	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return AgentHealth{}
	}

	mem, _ := p.MemoryInfo()
	cpu, _ := p.CPUPercent()

	return AgentHealth{
		CPUPercent:      float32(cpu),
		MemoryMB:        float32(mem.RSS) / 1024 / 1024,
		Goroutines:      runtime.NumGoroutine(),
		UptimeSec:       uint64(time.Since(startTime).Seconds()),
		MetricsFailures: atomic.LoadUint64(&metricsFailures),
		LogFailures:     atomic.LoadUint64(&logFailures),
	}
}
