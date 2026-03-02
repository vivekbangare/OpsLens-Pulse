package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"opslense-pulse/agent/collector"
	"opslense-pulse/agent/config"
	"opslense-pulse/agent/containers"
	"opslense-pulse/agent/heartbeat"
	"opslense-pulse/agent/identity"
	"opslense-pulse/agent/internal"
	"opslense-pulse/agent/metrics"
	"opslense-pulse/agent/queue"
	"opslense-pulse/agent/retry"
	"opslense-pulse/agent/safe"
	"opslense-pulse/agent/sender"
	"opslense-pulse/shared"
)

const AgentVersion = "1.0.0"

func checkServerConnectivity(serverURL string) error {
	healthURL := fmt.Sprintf("%s/health/live", serverURL)

	resp, err := http.Get(healthURL)
	if err != nil {
		return fmt.Errorf("cannot connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server unhealthy, status: %d", resp.StatusCode)
	}

	return nil
}

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
	// Server Connectivity Check
	// -------------------------

	shared.Info("checking server connectivity...", "url", serverURL)

	err = retry.Do(5, 3*time.Second, func() error {
		return checkServerConnectivity(serverURL)
	})

	if err != nil {
		shared.Error("❌ unable to reach server after retries", "error", err.Error())
		log.Fatal("server connectivity check failed")
	}

	shared.Info("✅ connected to server successfully", "url", serverURL)
	// -------------------------
	// Agent Registration (ONCE)
	// -------------------------

	shared.Info("Registering agent metadata...")

	provider, publicIP, region := detectCloud()
	privateIP := getPrivateIP()
	k8sIP := getK8sNodeIP()
	osName, _ := metrics.HostInfo()

	metadata := shared.AgentInfo{
		AgentID:       agentID,
		Hostname:      hostname,
		PrivateIP:     privateIP,
		PublicIP:      publicIP,
		K8sNodeIP:     k8sIP,
		OS:            osName,
		Version:       AgentVersion,
		Environment:   cfg.Tags["env"],
		CloudProvider: provider,
		CloudRegion:   region,
		Tags:          cfg.Tags,
		SystemTags: map[string]string{
			"cloud_provider": provider,
			"cloud_region":   region,
		},
	}

	if err := retry.Do(3, 2*time.Second, func() error {
		return sender.SendAgentRegistration(serverURL, apiKey, metadata)
	}); err != nil {
		shared.Error("agent registration failed", "error", err.Error())
	} else {
		shared.Info("agent registered successfully")
	}

	lastMetaHash := loadLastMetaHash()
	currentHash := metadataHash(metadata)

	if currentHash != lastMetaHash {

		shared.Info("sending initial metadata to server")

		err := retry.Do(10, 3*time.Second, func() error {
			return sender.SendAgentRegistration(serverURL, apiKey, metadata)
		})

		if err != nil {
			shared.Error("fatal: cannot register agent", "error", err.Error())
			log.Fatal("registration failed permanently")
		}

		saveMetaHash(currentHash)
		lastMetaHash = currentHash

		shared.Info("agent registered successfully")
	}

	// -------------------------
	// Context + Shutdown
	// -------------------------

	ctx, cancel := context.WithCancel(context.Background())

	// -------------------------
	// Metadata Watcher (IP Change Detection)
	// -------------------------

	safe.Go("metadata-watcher", func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for {
			select {

			case <-ctx.Done():
				return

			case <-ticker.C:

				newHostname, _ := os.Hostname()
				newProvider, newPublicIP, newRegion := detectCloud()
				newPrivateIP := getPrivateIP()
				newK8sIP := getK8sNodeIP()
				newOS, _ := metrics.HostInfo()

				newMetadata := shared.AgentInfo{
					AgentID:       agentID,
					Hostname:      newHostname,
					PrivateIP:     newPrivateIP,
					PublicIP:      newPublicIP,
					K8sNodeIP:     newK8sIP,
					OS:            newOS,
					Version:       AgentVersion,
					Environment:   cfg.Tags["env"],
					CloudProvider: newProvider,
					CloudRegion:   newRegion,
					Tags:          cfg.Tags,
					SystemTags: map[string]string{
						"cloud_provider": newProvider,
						"cloud_region":   newRegion,
					},
				}

				newHash := metadataHash(newMetadata)

				if newHash == lastMetaHash {
					continue
				}

				shared.Info("metadata change detected, updating agent")

				err := retry.Do(3, 2*time.Second, func() error {
					return sender.SendAgentRegistration(serverURL, apiKey, newMetadata)
				})

				if err != nil {
					shared.Error("metadata update failed", "error", err.Error())
					continue
				}

				lastMetaHash = newHash

				shared.Info("agent metadata updated successfully")
				internal.IncMetadataUpdate()
			}
		}
	})
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

	interval := cfg.Agent.IntervalSeconds

	if internal.FileSize(filepath.Join(collector.QueueDir, "metrics.queue")) > 100*1024*1024 {
		interval = interval * 2 // slow down if queue > 100MB
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
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
			health.QueueSizeBytes = internal.FileSize(filepath.Join(collector.QueueDir, "metrics.queue"))
			health.LogQueueSizeBytes = internal.FileSize(filepath.Join(collector.QueueDir, "logs.queue"))
			health.ConsecutiveFailures = internal.GetFailures()
			health.LastSendSuccessUnix = internal.GetLastSuccess()
			health.MetadataUpdates = internal.GetMetadataUpdates()
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

func getPrivateIP() string {

	// Step 1: detect default route interface
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()

		localAddr := conn.LocalAddr().(*net.UDPAddr)
		if localAddr.IP != nil {
			return localAddr.IP.String()
		}
	}

	// Step 2: fallback to scanning interfaces
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range ifaces {

		// Skip down or loopback
		if iface.Flags&net.FlagUp == 0 ||
			iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Skip virtual interfaces
		if strings.HasPrefix(iface.Name, "docker") ||
			strings.HasPrefix(iface.Name, "veth") ||
			strings.HasPrefix(iface.Name, "cni") ||
			strings.HasPrefix(iface.Name, "flannel") {
			continue
		}

		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				ip := ipnet.IP
				if ip == nil || ip.To4() == nil {
					continue
				}

				if isPrivateIP(ip) {
					return ip.String()
				}
			}
		}
	}

	return ""
}

func isPrivateIP(ip net.IP) bool {

	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}

	for _, cidr := range privateRanges {
		_, block, _ := net.ParseCIDR(cidr)
		if block.Contains(ip) {
			return true
		}
	}

	return false
}

var cachedProvider string
var cachedRegion string
var cachedPublicIP string
var lastCloudCheck time.Time

func detectCloud() (string, string, string) {

	// Cache for 15 minutes
	if time.Since(lastCloudCheck) < 15*time.Minute && cachedProvider != "" {
		return cachedProvider, cachedPublicIP, cachedRegion
	}

	client := &http.Client{Timeout: 2 * time.Second}

	// ===============================
	// AWS IMDSv2
	// ===============================

	tokenReq, _ := http.NewRequest("PUT",
		"http://169.254.169.254/latest/api/token",
		nil)

	tokenReq.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", "21600")

	tokenResp, err := client.Do(tokenReq)
	if err == nil && tokenResp.StatusCode == 200 {
		defer tokenResp.Body.Close()

		tokenBytes, _ := io.ReadAll(tokenResp.Body)
		token := strings.TrimSpace(string(tokenBytes))

		metaReq, _ := http.NewRequest("GET",
			"http://169.254.169.254/latest/meta-data/public-ipv4",
			nil)

		metaReq.Header.Set("X-aws-ec2-metadata-token", token)

		metaResp, err2 := client.Do(metaReq)
		if err2 == nil && metaResp.StatusCode == 200 {
			defer metaResp.Body.Close()
			body, _ := io.ReadAll(metaResp.Body)

			cachedProvider = "aws"
			cachedPublicIP = strings.TrimSpace(string(body))
			cachedRegion = ""
			lastCloudCheck = time.Now()

			return cachedProvider, cachedPublicIP, cachedRegion
		}
	}

	// ===============================
	// GCP
	// ===============================

	gcpReq, _ := http.NewRequest("GET",
		"http://metadata.google.internal/computeMetadata/v1/instance",
		nil)
	gcpReq.Header.Set("Metadata-Flavor", "Google")

	if resp, err := client.Do(gcpReq); err == nil && resp.StatusCode == 200 {
		resp.Body.Close()

		cachedProvider = "gcp"
		cachedPublicIP = ""
		cachedRegion = ""
		lastCloudCheck = time.Now()

		return cachedProvider, cachedPublicIP, cachedRegion
	}

	// ===============================
	// Fallback (On-Prem)
	// ===============================

	if resp, err := client.Get("https://api.ipify.org"); err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		cachedProvider = "onprem"
		cachedPublicIP = strings.TrimSpace(string(body))
		cachedRegion = ""
		lastCloudCheck = time.Now()

		return cachedProvider, cachedPublicIP, cachedRegion
	}

	return "unknown", "", ""
}

func getK8sNodeIP() string {

	// 1️⃣ If explicitly injected (best way)
	if ip := os.Getenv("K8S_NODE_IP"); ip != "" {
		return ip
	}

	// 2️⃣ Detect if running inside Kubernetes
	if _, err := os.Stat("/var/run/secrets/kubernetes.io"); err == nil {

		// Inside K8s, but node IP not injected
		// Fallback to private IP (safe assumption)
		return getPrivateIP()
	}

	return ""
}

func metadataHash(m shared.AgentInfo) string {

	data := strings.Join([]string{
		m.AgentID,
		m.Hostname,
		m.PrivateIP,
		m.PublicIP,
		m.K8sNodeIP,
		m.OS,
		m.Version,
		m.Environment,
		m.CloudProvider,
		m.CloudRegion,
	}, "|")

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func loadLastMetaHash() string {
	path := filepath.Join(collector.StateDir, "metadata.hash")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func saveMetaHash(hash string) {
	path := filepath.Join(collector.StateDir, "metadata.hash")
	tmp := path + ".tmp"
	_ = os.WriteFile(tmp, []byte(hash), 0600)
	_ = os.Rename(tmp, path)
}
