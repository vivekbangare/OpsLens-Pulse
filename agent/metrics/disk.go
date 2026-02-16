package metrics

import (
	"runtime"

	"github.com/shirou/gopsutil/v3/disk"
)

func DiskUsage() (uint64, uint64) {

	path := "/"
	if runtime.GOOS == "windows" {
		path = "C:\\"
	}
	usage, err := disk.Usage(path)

	if err != nil {
		return 0, 0
	}

	total := usage.Total / 1024 / 1024
	used := usage.Used / 1024 / 1024

	return total, used
}
