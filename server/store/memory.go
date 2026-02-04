package store

import (
	"sync"
	"time"

	"opslense-pulse/shared"
)

type HostState struct {
	Metrics  shared.HostMetrics
	LastSeen time.Time
	Logs     []string
}

var (
	mu    sync.RWMutex // changed to RWMutex for better read performance
	Hosts = make(map[string]*HostState)
)

func SaveMetrics(m shared.HostMetrics) {
	mu.Lock()
	defer mu.Unlock()

	h, ok := Hosts[m.Hostname]
	if !ok {
		h = &HostState{}
		Hosts[m.Hostname] = h
	}
	h.Metrics = m
}

func UpdateHeartbeat(host string, t time.Time) {
	mu.Lock()
	defer mu.Unlock()

	h, ok := Hosts[host]
	if !ok {
		h = &HostState{}
		Hosts[host] = h
	}
	h.LastSeen = t
}

func GetAll() []map[string]interface{} {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	out := []map[string]interface{}{}

	for host, h := range Hosts {
		alive := now.Sub(h.LastSeen) < 15*time.Second

		out = append(out, map[string]interface{}{
			"hostname":    host,
			"os":          h.Metrics.OS,
			"cpu_percent": h.Metrics.CPUPercent,
			"mem_used_mb": h.Metrics.MemUsedMB,
			"uptime_sec":  h.Metrics.UpTimeSec,
			"last_seen":   h.LastSeen,
			"alive":       alive,
		})
	}
	return out
}

// SaveLog adds a log entry for a host
func SaveLog(hostname string, log string) {
	mu.Lock()
	defer mu.Unlock()
	h, ok := Hosts[hostname]
	if !ok {
		h = &HostState{}
		Hosts[hostname] = h
	}
	h.Logs = append(h.Logs, log)
}

// GetLogs returns logs for a host
func GetLogs(hostname string) []string {
	mu.Lock()
	defer mu.Unlock()
	h, ok := Hosts[hostname]
	if !ok {
		return []string{}
	}
	return h.Logs
}

func GetFiltered(tags map[string]string) []map[string]interface{} {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	out := []map[string]interface{}{}

	for host, h := range Hosts {
		if !matchTags(h.Metrics.Tags, tags) {
			continue
		}

		alive := now.Sub(h.LastSeen) < 15*time.Second

		out = append(out, map[string]interface{}{
			"hostname":    host,
			"os":          h.Metrics.OS,
			"cpu_percent": h.Metrics.CPUPercent,
			"mem_used_mb": h.Metrics.MemUsedMB,
			"uptime_sec":  h.Metrics.UpTimeSec,
			"last_seen":   h.LastSeen,
			"alive":       alive,
			"tags":        h.Metrics.Tags,
		})
	}
	return out
}

func matchTags(hostTags, filters map[string]string) bool {
	for k, v := range filters {
		if hostTags == nil {
			return false
		}
		if hostTags[k] != v {
			return false
		}
	}
	return true
}
