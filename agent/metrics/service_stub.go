//go:build !linux && !windows

package metrics

import (
	"opslense-pulse/shared"
)

func GetServiceStatus(names []string) []shared.ServiceStatus {
	return nil
}
