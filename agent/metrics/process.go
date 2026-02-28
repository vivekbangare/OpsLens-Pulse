package metrics

import "github.com/shirou/gopsutil/v3/process"

func ProcessStats() (int, int) {
	procs, err := process.Processes()
	if err != nil {
		return 0, 0
	}

	total := len(procs)
	running := 0

	for _, p := range procs {
		status, err := p.Status()
		if err != nil {
			continue
		}
		for _, s := range status {
			if s == "R" {
				running++
				break
			}
		}
	}

	return total, running
}
