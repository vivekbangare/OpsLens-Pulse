// collector/base.go
package collector

import "runtime"

var BaseDir string

func init() {
	if runtime.GOOS == "windows" {
		BaseDir = `C:\ProgramData\OpsLens-Pulse`
	} else {
		BaseDir = "/var/lib/opslens-pulse"
	}
}
