package containers

import (
	"context"
	"encoding/json"
	"fmt"
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

var dockerClient *client.Client

// Initialize reusable Docker client
func init() {
	var err error

	dockerClient, err = client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)

	if err != nil {
		// Do NOT panic — allow agent to run without docker
		dockerClient = nil
	}
}

// ListRunning returns running containers
func ListRunning() ([]types.Container, error) {
	if dockerClient == nil {
		return nil, fmt.Errorf("docker client not initialized")
	}

	return dockerClient.ContainerList(
		context.Background(),
		types.ContainerListOptions{},
	)
}

// GetContainerMetrics returns CPU + Memory stats
func GetContainerMetrics(c types.Container) (RawContainerMetrics, error) {
	if dockerClient == nil {
		return RawContainerMetrics{}, fmt.Errorf("docker client not initialized")
	}

	stats, err := dockerClient.ContainerStatsOneShot(
		context.Background(),
		c.ID,
	)
	if err != nil {
		return RawContainerMetrics{}, err
	}
	defer stats.Body.Close()

	var s types.StatsJSON
	if err := json.NewDecoder(stats.Body).Decode(&s); err != nil {
		return RawContainerMetrics{}, err
	}

	cpuDelta := float64(
		s.CPUStats.CPUUsage.TotalUsage -
			s.PreCPUStats.CPUUsage.TotalUsage,
	)

	sysDelta := float64(
		s.CPUStats.SystemUsage -
			s.PreCPUStats.SystemUsage,
	)

	cpuPercent := 0.0
	if cpuDelta > 0 && sysDelta > 0 {
		cpuPercent = (cpuDelta / sysDelta) *
			float64(len(s.CPUStats.CPUUsage.PercpuUsage)) * 100
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

// GetContainerLogs fetches recent container logs
func GetContainerLogs(containerID string, tail int) ([]string, error) {
	if dockerClient == nil {
		return nil, fmt.Errorf("docker client not initialized")
	}

	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Tail:       fmt.Sprintf("%d", tail),
	}

	reader, err := dockerClient.ContainerLogs(
		context.Background(),
		containerID,
		options,
	)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSuffix(string(data), "\n"), "\n"), nil
}

func Shutdown() {
	if dockerClient != nil {
		dockerClient.Close()
	}
}
