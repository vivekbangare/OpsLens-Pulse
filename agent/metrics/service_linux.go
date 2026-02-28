//go:build linux

package metrics

import (
	"opslense-pulse/shared"
	"os/exec"
	"strings"
)

var prevServiceState = make(map[string]bool)

func GetServiceStatus(names []string) []shared.ServiceStatus {

	var results []shared.ServiceStatus

	for _, name := range names {

		running, enabled := getLinuxServiceState(name)

		status := "healthy"

		if !enabled {
			status = "disabled"
		} else if !running {
			status = "down"
		}
		previousRunning := prevServiceState[name]
		crashDectected := false
		if previousRunning && !running && enabled {
			crashDectected = true
		}
		prevServiceState[name] = running

		results = append(results, shared.ServiceStatus{
			Name:          name,
			Running:       running,
			Enabled:       enabled,
			Status:        status,
			CrashDetected: crashDectected,
		})
	}

	return results
}

func getLinuxServiceState(name string) (bool, bool) {
	out, err := exec.Command(
		"systemctl",
		"show",
		name,
		"--property=ActiveState,UnitFileState",
	).Output()

	if err != nil {
		return false, false
	}

	text := string(out)

	running := strings.Contains(text, "ActiveState=active")
	enabled := strings.Contains(text, "UnitFileState=enabled")

	return running, enabled
}
