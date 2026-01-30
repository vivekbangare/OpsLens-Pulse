package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"opslense-pulse/agent/config"
	"opslense-pulse/agent/heartbeat"
	"opslense-pulse/agent/metrics"
	"opslense-pulse/agent/sender"
	"opslense-pulse/shared"
)

const AgentVersion = "1.0.0"

func printHelp() {
	fmt.Print(`
OpsLens-Pulse Agent

Usage:
  opslens-pulse-agent [options]

Options:
  --config <path>     Path to YAML config file
  --version           Show version
  --help              Show help

Defaults:
  Linux:   /etc/opslens-pulse/agent-config.yaml
  Windows: C:\ProgramData\OpsLens-Pulse\agent-config.yaml

Env:
  OPS_AGENT_CONFIG
`)
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "Config path")
	showHelp := flag.Bool("help", false, "Help")
	showVersion := flag.Bool("version", false, "Version")
	flag.Parse()

	if *showHelp {
		printHelp()
		return
	}
	if *showVersion {
		fmt.Println(AgentVersion)
		return
	}

	cfg, err := config.LoadOrCreateConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	hostname, _ := os.Hostname()
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = cfg.Server.URL
	}

	token := os.Getenv("SERVER_TOKEN")
	if token == "" {
		log.Fatal("SERVER_TOKEN env variable is required")
	}

	for {
		osName, uptime := metrics.HostInfo()
		memTotal, memUsed := metrics.Memory()

		m := shared.HostMetrics{
			Hostname:   hostname,
			OS:         osName,
			Timestamp:  time.Now(),
			Cores:      runtime.NumCPU(),
			MemTotalMB: memTotal,
			MemUsedMB:  memUsed,
			UpTimeSec:  uptime,
			CPUPercent: metrics.CPUPercent(),
		}

		// Send metrics
		if err := sender.Send(serverURL, token, m); err != nil {
			log.Println("Send failed:", err)
		}

		// Send heartbeat
		if err := heartbeat.Send(serverURL, token, hostname); err != nil {
			log.Println("Heartbeat failed:", err)
		}

		time.Sleep(time.Duration(cfg.Agent.IntervalSeconds) * time.Second)
	}
}
