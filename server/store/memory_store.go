package store

import (
	"context"
	"opslense-pulse/shared"
	"sync"
	"time"
)

// -------------------------------
// MemoryStore struct
// -------------------------------
type MemoryStore struct {
	mu         sync.RWMutex
	hosts      map[string]*HostState
	apiKeys    map[string]shared.APIKey
	containers map[string]*ContainerState
}

type ContainerState struct {
	Metrics  shared.ContainerMetrics
	Logs     []string
	LastSeen time.Time
}

// -------------------------------
// Constructor
// -------------------------------
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		hosts:      make(map[string]*HostState),
		apiKeys:    make(map[string]shared.APIKey),
		containers: make(map[string]*ContainerState),
	}
}

func (m *MemoryStore) SaveContainerMetrics(ctx context.Context, metrics shared.ContainerMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := metrics.Hostname + ":" + metrics.ContainerID
	c, ok := m.containers[key]
	if !ok {
		c = &ContainerState{}
		m.containers[key] = c
	}
	c.Metrics = metrics
	c.LastSeen = time.Now()
	return nil
}

func (m *MemoryStore) InsertContainerLogs(ctx context.Context, batch shared.ContainerLogBatch) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, l := range batch.Logs {
		key := batch.Hostname + ":" + l.ContainerID
		c, ok := m.containers[key]
		if !ok {
			c = &ContainerState{}
			m.containers[key] = c
		}

		t := time.Now()
		if l.Timestamp > 0 {
			t = time.Unix(l.Timestamp, 0)
		}
		c.Logs = append(c.Logs, l.Message+" ["+t.Format("2006-01-02 15:04:05")+"]")
		c.LastSeen = t
	}
	return nil
}

// -------------------------------
// SaveMetrics: store host metrics
// -------------------------------
func (m *MemoryStore) SaveMetrics(ctx context.Context, metrics shared.HostMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	h, ok := m.hosts[metrics.AgentID]
	if !ok {
		h = &HostState{}
		m.hosts[metrics.AgentID] = h
	}

	h.Metrics = metrics
	h.LastSeen = time.Now().Unix()

	if ip, ok := metrics.Tags["ip"]; ok {
		h.IP = ip
	}

	return nil
}

// -------------------------------
// Logs
// -------------------------------
func (m *MemoryStore) InsertLogs(ctx context.Context, batch shared.LogBatch) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	h, ok := m.hosts[batch.Hostname]
	if !ok {
		h = &HostState{}
		m.hosts[batch.Hostname] = h
	}

	for _, l := range batch.Logs {
		t := time.Now()
		if l.Timestamp > 0 {
			t = time.Unix(l.Timestamp, 0)
		}
		h.Logs = append(h.Logs, shared.LogEntry{
			TenantID:  l.TenantID,
			AgentID:   l.AgentID,
			Hostname:  l.Hostname,
			Level:     l.Level,
			Message:   l.Message + " [" + t.Format("2006-01-02 15:04:05") + "]",
			Timestamp: l.Timestamp,
			Tags:      l.Tags,
		})
	}
	return nil
}

// -------------------------------
// Fetch logs
// -------------------------------
func (m *MemoryStore) GetLogs(
	ctx context.Context,
	TenantID string,
	hostname string,
	agentID string,
	from, to time.Time,
	level string,
	source string,
	source_type string,
	limit int,
) ([]shared.LogEntry, error) {

	m.mu.RLock()
	defer m.mu.RUnlock()

	var out []shared.LogEntry

	for _, h := range m.hosts {
		for _, l := range h.Logs {

			if TenantID != "" && l.TenantID != TenantID {
				continue
			}
			if agentID != "" && l.AgentID != agentID {
				continue
			}
			if hostname != "" && l.Hostname != hostname {
				continue
			}
			if level != "" && l.Level != level {
				continue
			}
			if source != "" {
				if src, ok := l.Tags["source"]; !ok || src != source {
					continue
				}
			}

			t := time.Unix(l.Timestamp, 0)
			if !from.IsZero() && t.Before(from) {
				continue
			}
			if !to.IsZero() && t.After(to) {
				continue
			}

			out = append(out, l)

			if limit > 0 && len(out) >= limit {
				return out, nil
			}
		}
	}

	return out, nil
}

func (m *MemoryStore) GetLogSources(ctx context.Context, TenantID, agentID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	seen := make(map[string]struct{})

	for _, h := range m.hosts {
		for _, l := range h.Logs {
			if l.TenantID != TenantID || l.AgentID != agentID {
				continue
			}
			if src, ok := l.Tags["source"]; ok {
				seen[src] = struct{}{}
			}
		}
	}

	var out []string
	for s := range seen {
		out = append(out, s)
	}
	return out, nil
}

// -------------------------------
// API Keys
// -------------------------------
func (m *MemoryStore) CountAPIKeys() (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.apiKeys), nil
}

func (m *MemoryStore) InsertAPIKey(k shared.APIKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.apiKeys[k.KeyHash] = k
	return nil
}

func (m *MemoryStore) ValidateAPIKey(rawKey string) (bool, string, error) {
	hash := HashAPIKey(rawKey)
	m.mu.RLock()
	defer m.mu.RUnlock()
	k, ok := m.apiKeys[hash]
	if !ok {
		return false, "", nil
	}
	return true, k.TenantID, nil
}

func (m *MemoryStore) UpsertAgentHeartbeat(ctx context.Context, hb shared.Heartbeat) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	h, ok := m.hosts[hb.Hostname]
	if !ok {
		h = &HostState{}
		m.hosts[hb.Hostname] = h
	}

	if hb.Timestamp == 0 {
		h.LastSeen = time.Now().Unix()
	} else {
		h.LastSeen = hb.Timestamp
	}

	return nil
}

func (m *MemoryStore) ListAgents(ctx context.Context, TenantID string) ([]shared.AgentInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var agents []shared.AgentInfo

	for hostname, h := range m.hosts {
		agents = append(agents, shared.AgentInfo{
			TenantID: TenantID,
			Hostname: hostname,
			LastSeen: time.Unix(h.LastSeen, 0),
		})
	}

	return agents, nil
}

func (m *MemoryStore) GetLatestHostMetrics(ctx context.Context, TenantID string) (map[string]shared.HostMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Memory store does not persist time-series metrics
	// Return empty map so UI still works
	return map[string]shared.HostMetrics{}, nil
}

func (m *MemoryStore) UpsertAgentMetadata(ctx context.Context, hm shared.HostMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.hosts[hm.AgentID]
	if !ok {
		// first time we see this agent
		state = &HostState{
			Metrics:  hm,
			LastSeen: time.Now().Unix(),
		}
		m.hosts[hm.AgentID] = state
	} else {
		// update metadata only
		state.Metrics.Hostname = hm.Hostname
		state.Metrics.PrivateIP = hm.PrivateIP
		state.Metrics.PublicIP = hm.PublicIP
		state.Metrics.K8sNodeIP = hm.K8sNodeIP
		state.Metrics.OS = hm.OS
		state.Metrics.Tags = hm.Tags
	}

	return nil
}

func (m *MemoryStore) GetLogsTimeline(
	ctx context.Context,
	tenantID string,
	agentID string,
	sourceType string,
	from, to time.Time,
) (TimelineResult, error) {

	// Memory store does not persist timeline aggregation.
	// Return empty result safely.

	return TimelineResult{
		Points:  []TimelinePoint{},
		Anomaly: false,
	}, nil
}

func (m *MemoryStore) FetchContainerMetrics(
	ctx context.Context,
	tenantID string,
	agentID string,
) ([]shared.ContainerMetrics, error) {

	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []shared.ContainerMetrics

	for _, c := range m.containers {
		if c.Metrics.TenantID == tenantID &&
			c.Metrics.AgentID == agentID {
			result = append(result, c.Metrics)
		}
	}

	return result, nil
}

var _ Store = (*MemoryStore)(nil)
