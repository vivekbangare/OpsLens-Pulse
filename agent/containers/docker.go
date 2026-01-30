package containers

import "github.com/shirou/gopsutil/v3/docker"

type Container struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func List() []Container {
	stats, _ := docker.GetDockerStat()
	var out []Container
	for _, c := range stats {
		out = append(out, Container{
			ID:     c.ContainerID,
			Name:   c.Name,
			Status: c.Status,
		})
	}
	return out
}
