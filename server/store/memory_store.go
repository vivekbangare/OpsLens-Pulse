package store

import (
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

func (m *MemoryStore) SaveContainerMetrics(metrics shared.ContainerMetrics) error {
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

func (m *MemoryStore) InsertContainerLogs(batch shared.ContainerLogBatch) error {
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
func (m *MemoryStore) SaveMetrics(metrics shared.HostMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	h, ok := m.hosts[metrics.Hostname]
	if !ok {
		h = &HostState{}
		m.hosts[metrics.Hostname] = h
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
func (m *MemoryStore) InsertLogs(batch shared.LogBatch) error {
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
			AccountID: l.AccountID,
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
	accountID string,
	hostname string,
	agentID string,
	from, to time.Time,
	level string,
	source string,
	limit int,
) ([]shared.LogEntry, error) {

	m.mu.RLock()
	defer m.mu.RUnlock()

	var out []shared.LogEntry

	for _, h := range m.hosts {
		for _, l := range h.Logs {

			if accountID != "" && l.AccountID != accountID {
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

func (m *MemoryStore) GetLogSources(accountID, agentID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	seen := make(map[string]struct{})

	for _, h := range m.hosts {
		for _, l := range h.Logs {
			if l.AccountID != accountID || l.AgentID != agentID {
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

func (m *MemoryStore) ValidateAPIKey(rawKey string) (bool, error) {
	hash := hashAPIKey(rawKey)
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.apiKeys[hash]
	return ok, nil
}

func (m *MemoryStore) UpsertAgentHeartbeat(hb shared.Heartbeat) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	h, ok := m.hosts[hb.Hostname]
	if !ok {
		h = &HostState{}
		m.hosts[hb.Hostname] = h
	}
	h.LastSeen = hb.Timestamp.Unix()
	return nil
}

func (m *MemoryStore) ListAgents(accountID string) ([]shared.AgentInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var agents []shared.AgentInfo

	for hostname, h := range m.hosts {
		agents = append(agents, shared.AgentInfo{
			AccountID: accountID,
			Hostname:  hostname,
			LastSeen:  time.Unix(h.LastSeen, 0),
		})
	}

	return agents, nil
}

func (m *MemoryStore) GetLatestHostMetrics(accountID string) (map[string]shared.HostMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Memory store does not persist time-series metrics
	// Return empty map so UI still works
	return map[string]shared.HostMetrics{}, nil
}

func (m *MemoryStore) UpsertAgentMetadata(hm shared.HostMetrics) error {
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
		state.Metrics.IP = hm.IP
		state.Metrics.OS = hm.OS
		state.Metrics.Tags = hm.Tags
	}

	return nil
}

var _ Store = (*MemoryStore)(nil)
