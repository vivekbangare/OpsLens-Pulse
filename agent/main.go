package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"opslense-pulse/agent/collector"
	"opslense-pulse/agent/config"
	"opslense-pulse/agent/containers"
	"opslense-pulse/agent/heartbeat"
	"opslense-pulse/agent/identity"
	"opslense-pulse/agent/metrics"
	"opslense-pulse/agent/queue"
	"opslense-pulse/agent/retry"
	"opslense-pulse/agent/safe"
	"opslense-pulse/agent/sender"
	"opslense-pulse/shared"
)

const AgentVersion = "1.0.0"

func main() {

	shared.InitLogger("agent")
	shared.Info("🚀 OpsLens-Pulse Agent starting...")

	// -------------------------
	// Config
	// -------------------------

	var configPath string
	flag.StringVar(&configPath, "config", "", "Config path")
	flag.Parse()

	cfg, path, created, err := config.LoadOrCreateConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
	if created {
		log.Printf("📄 Config created at: %s\n", path)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	// -------------------------
	// TLS HTTP Client
	// -------------------------

	sender.InitHTTPClient(cfg.Server.InsecureSkipVerify)
	sender.SetAgentVersion(AgentVersion)
	// -------------------------
	// Identity
	// -------------------------

	agentID, err := identity.LoadOrCreateAgentID()
	if err != nil {
		log.Fatal(err)
	}

	hostname, _ := os.Hostname()
	serverURL := cfg.Server.URL
	apiKey := cfg.Server.APIKey

	if serverURL == "" || apiKey == "" {
		log.Fatal("server.url and server.api_key must be set")
	}

	// -------------------------
	// Context + Shutdown
	// -------------------------

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// -------------------------
	// Queues
	// -------------------------

	metricsQueue := queue.New(filepath.Join(collector.QueueDir, "metrics.queue"))
	logsQueue := queue.New(filepath.Join(collector.QueueDir, "logs.queue"))

	safe.Go("metrics-sender", func() {
		queue.StartMetricsSender(ctx, metricsQueue, serverURL, apiKey)
	})

	safe.Go("logs-sender", func() {
		queue.StartLogsSender(ctx, logsQueue, serverURL, apiKey)
	})
	// -------------------------
	// Log Sources (file / dir )
	// -------------------------

	for _, src := range cfg.Agent.Logs {
		srcCopy := src
		safe.Go("log-source-"+srcCopy.Name, func() {
			collector.StartLogSource(
				ctx,
				agentID,
				hostname,
				srcCopy,
				logsQueue,
			)
		})
	}

	// -------------------------
	// Docker collectors
	// -------------------------

	if _, err := containers.ListRunning(ctx); err == nil {
		safe.Go("container-metrics", func() {
			collector.StartContainerMetricsCollector(
				ctx,
				agentID,
				hostname,
				serverURL,
				apiKey,
				10,
			)
		})

		safe.Go("container-logs", func() {
			collector.StartContainerLogsCollector(
				ctx,
				agentID,
				hostname,
				logsQueue,
				10,
			)
		})
	}

	// -------------------------
	// Main Metrics Loop
	// -------------------------

	ticker := time.NewTicker(time.Duration(cfg.Agent.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {

		case <-ctx.Done():
			shared.Info("🛑 Agent shutting down...")
			containers.Shutdown()
			return

		case <-sig:
			cancel()

		case <-ticker.C:
			disks := metrics.DiskUsageAll()
			osName, uptime := metrics.HostInfo()
			health := metrics.GetAgentHealth()
			cpu := metrics.CPUPercent()
			cpuAnomaly := metrics.DetectCPUAnomaly(cpu)
			memTotal, memUsed := metrics.Memory()
			memAnomaly := metrics.DetectMemoryAnomaly(memUsed, memTotal)
			netRates := metrics.NetworkRates()

			var network []shared.NetworkInterface
			for _, n := range netRates {
				network = append(network, shared.NetworkInterface{
					Name:   n.Name,
					InBps:  float32(n.RecvPerSec),
					OutBps: float32(n.SentPerSec),
				})
			}

			m := shared.HostMetrics{
				AgentID:    agentID,
				Hostname:   hostname,
				OS:         osName,
				Version:    AgentVersion,
				Timestamp:  time.Now().Unix(),
				Cores:      runtime.NumCPU(),
				CPUPercent: float32(cpu),

				CPUCritical: cpuAnomaly.Critical,
				CPUSpike:    cpuAnomaly.Spike,
				MemTotalMB:  float32(memTotal),
				MemUsedMB:   float32(memUsed),
				MemCritical: memAnomaly.Critical,
				MemPressure: memAnomaly.Pressure,
				Disks:       disks,
				UptimeSec:   uptime,
				Tags:        cfg.Tags,
				Network:     network,
				Agent: shared.AgentHealth{
					CPUPercent:      health.CPUPercent,
					MemoryMB:        health.MemoryMB,
					Goroutines:      health.Goroutines,
					UptimeSec:       health.UptimeSec,
					MetricsFailures: health.MetricsFailures,
					LogFailures:     health.LogFailures,
				},
			}

			// Enqueue metrics
			if err := metricsQueue.Enqueue(m); err != nil {
				shared.Error("metrics queue error", "error", err.Error())
			}

			// Heartbeat (direct)
			_ = retry.Do(3, 2*time.Second, func() error {
				return heartbeat.Send(serverURL, apiKey, agentID, hostname)
			})
		}
	}
}
