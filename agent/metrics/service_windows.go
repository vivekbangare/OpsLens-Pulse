//go:build windows
// +build windows

package metrics

import (
	"opslense-pulse/shared"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

var prevServiceState = make(map[string]bool)

func GetServiceStatus(names []string) []shared.ServiceStatus {

	var results []shared.ServiceStatus

	m, err := mgr.Connect()
	if err != nil {
		return results
	}
	defer m.Disconnect()

	for _, name := range names {

		s, err := m.OpenService(name)

		if err != nil {
			results = append(results, shared.ServiceStatus{
				Name:    name,
				Running: false,
				Enabled: false,
				Status:  "down",
			})
			continue
		}

		status, err := s.Query()
		cfg, _ := s.Config()

		running := false
		enabled := false

		if err == nil {
			running = status.State == svc.Running
			enabled = cfg.StartType != mgr.StartDisabled
		}

		serviceStatus := "healthy"
		if !enabled {
			serviceStatus = "disabled"
		} else if !running {
			serviceStatus = "down"
		}
		previousRunning := prevServiceState[name]

		crashDected := false

		if previousRunning && !running && enabled {
			crashDected = true
		}

		prevServiceState[name] = running
		results = append(results, shared.ServiceStatus{
			Name:          name,
			Running:       running,
			Enabled:       enabled,
			Status:        serviceStatus,
			CrashDetected: crashDected,
		})

		s.Close()
	}

	return results
}
