package containers

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type RawContainerMetrics struct {
	ID         string
	Name       string
	Image      string
	Status     string
	CPU        float64
	MemUsedMB  uint64
	MemLimitMB uint64
	Timestamp  time.Time
}

func ListRunning() ([]types.Container, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	return cli.ContainerList(context.Background(), types.ContainerListOptions{})
}

func GetContainerMetrics(c types.Container) (RawContainerMetrics, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return RawContainerMetrics{}, err
	}
	defer cli.Close()

	stats, err := cli.ContainerStatsOneShot(context.Background(), c.ID)
	if err != nil {
		return RawContainerMetrics{}, err
	}
	defer stats.Body.Close()

	var s types.StatsJSON
	if err := json.NewDecoder(stats.Body).Decode(&s); err != nil {
		return RawContainerMetrics{}, err
	}

	cpuDelta := float64(s.CPUStats.CPUUsage.TotalUsage - s.PreCPUStats.CPUUsage.TotalUsage)
	sysDelta := float64(s.CPUStats.SystemUsage - s.PreCPUStats.SystemUsage)

	cpuPercent := 0.0
	if cpuDelta > 0 && sysDelta > 0 {
		cpuPercent = (cpuDelta / sysDelta) * float64(len(s.CPUStats.CPUUsage.PercpuUsage)) * 100
	}
	name := ""
	if len(c.Names) > 0 {
		name = strings.TrimPrefix(c.Names[0], "/")
	}
	return RawContainerMetrics{
		ID:         c.ID,
		Name:       name,
		Image:      c.Image,
		Status:     c.State,
		CPU:        cpuPercent,
		MemUsedMB:  s.MemoryStats.Usage / 1024 / 1024,
		MemLimitMB: s.MemoryStats.Limit / 1024 / 1024,
		Timestamp:  time.Now(),
	}, nil
}

func GetContainerLogs(containerID string, tail int) ([]string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Tail:       "100",
	}

	reader, err := cli.ContainerLogs(context.Background(), containerID, options)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// Split into lines
	lines := strings.Split(string(data), "\n")
	return lines, nil
}
