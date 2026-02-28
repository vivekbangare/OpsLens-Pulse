package metrics

import (
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/net"
)

type InterfaceRate struct {
	Name       string
	SentPerSec float64
	RecvPerSec float64
}

var (
	netMu   sync.Mutex
	prevNet = make(map[string]struct {
		Sent uint64
		Recv uint64
		Time time.Time
	})
)

func NetworkRates() []InterfaceRate {
	netMu.Lock()
	defer netMu.Unlock()

	stats, err := net.IOCounters(true)
	if err != nil {
		return nil
	}

	now := time.Now()
	var results []InterfaceRate

	for _, s := range stats {

		// Skip unwanted interfaces
		if strings.HasPrefix(s.Name, "lo") ||
			strings.HasPrefix(s.Name, "docker") ||
			strings.HasPrefix(s.Name, "veth") ||
			strings.HasPrefix(s.Name, "cni") {
			continue
		}

		prev, exists := prevNet[s.Name]

		if !exists {
			prevNet[s.Name] = struct {
				Sent uint64
				Recv uint64
				Time time.Time
			}{s.BytesSent, s.BytesRecv, now}
			continue
		}

		elapsed := now.Sub(prev.Time).Seconds()
		if elapsed <= 0 {
			continue
		}

		sentRate := float64(s.BytesSent-prev.Sent) / elapsed
		recvRate := float64(s.BytesRecv-prev.Recv) / elapsed

		results = append(results, InterfaceRate{
			Name:       s.Name,
			SentPerSec: sentRate,
			RecvPerSec: recvRate,
		})

		prevNet[s.Name] = struct {
			Sent uint64
			Recv uint64
			Time time.Time
		}{s.BytesSent, s.BytesRecv, now}
	}
	// Cleanup removed interfaces
	for name := range prevNet {
		found := false
		for _, s := range stats {
			if s.Name == name {
				found = true
				break
			}
		}
		if !found {
			delete(prevNet, name)
		}
	}
	return results
}
