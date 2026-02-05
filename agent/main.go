package main

import (
	"flag"
	"fmt"
	"log"
	"net"
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

func getLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range ifaces {
		// ignore loopback & down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}

			return ip.String() // return first valid IPv4
		}
	}

	return ""
}


func main() {
	shared.InitLogger("agent")
	log.Println("🚀 OpsLens-Pulse Agent starting...")

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

	cfg, path, created, err := config.LoadOrCreateConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
	if created {
		log.Printf("📄 Agent config created at: %s\n", path)
	} else {
		log.Printf("📄 Agent config loaded from: %s\n", path)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	hostname, _ := os.Hostname()
	serverURL := cfg.Server.URL
	token := cfg.Server.Token

	if serverURL == "" || token == "" {
		log.Fatal("🌐 server.url and 🔐 server.token must be set in agent config")
	}

	for {
		start := time.Now()

		// Collect metrics
		memTotal, memUsed := metrics.Memory()
		osName, uptime := metrics.HostInfo()

		m := shared.HostMetrics{
			Hostname:   hostname,
			OS:         osName,
			Timestamp:  time.Now(),
			Cores:      runtime.NumCPU(),
			MemTotalMB: memTotal,
			MemUsedMB:  memUsed,
			UpTimeSec:  uptime,
			CPUPercent: metrics.CPUPercent(),
			Tags:       cfg.Tags,
			IP:         getLocalIP(),
		}

		// Send metrics
		if err := sender.Send(serverURL, token, m); err != nil {
			log.Println("Send failed:", err)
		}

		// Send heartbeat
		if err := heartbeat.Send(serverURL, token, hostname); err != nil {
			log.Println("Heartbeat failed:", err)
		}

		// Correct sleep to avoid drift
		elapsed := time.Since(start)
		sleep := time.Duration(cfg.Agent.IntervalSeconds)*time.Second - elapsed
		if sleep > 0 {
			time.Sleep(sleep)
		}
	}
}
