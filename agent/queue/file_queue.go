package queue

import (
	"bufio"
	"encoding/json"
	"fmt"
	"opslense-pulse/shared"
	"os"
	"path/filepath"
	"sync"

	"github.com/shirou/gopsutil/v3/disk"
)

const maxQueueSize = 500 * 1024 * 1024 // 500MB
const minFreeDiskPercent = 5.0         // 5% minimum free disk

func hasEnoughDiskSpace(path string) bool {
	usage, err := disk.Usage(filepath.Dir(path))
	if err != nil {
		return true // fail open (avoid blocking agent)
	}

	freePercent := 100.0 - usage.UsedPercent
	return freePercent > minFreeDiskPercent
}

type FileQueue struct {
	path string
	mu   sync.Mutex
}

func New(path string) *FileQueue {
	// 🔐 Use secure permissions
	_ = os.MkdirAll(filepath.Dir(path), 0700)
	return &FileQueue{path: path}
}

func (q *FileQueue) Enqueue(v interface{}) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	// Disk free protection
	if !hasEnoughDiskSpace(q.path) {
		shared.Error("low disk space - queue write blocked", "path", q.path)
		return fmt.Errorf("low disk space")
	}

	// Check current queue size BEFORE writing
	info, err := os.Stat(q.path)
	if err == nil {
		if info.Size() > maxQueueSize {
			shared.Error("queue size limit exceeded", "path", q.path)
			return fmt.Errorf("queue size limit exceeded")
		}
	}

	f, err := os.OpenFile(q.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	_, err = f.Write(append(data, '\n'))
	return err
}

func (q *FileQueue) DequeueAll() ([]json.RawMessage, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	f, err := os.Open(q.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var msgs []json.RawMessage

	scanner := bufio.NewScanner(f)

	// ✅ 10MB buffer to avoid token overflow
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		msgs = append(msgs, json.RawMessage(scanner.Bytes()))
	}

	return msgs, scanner.Err()
}

func (q *FileQueue) Clear() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	return os.Truncate(q.path, 0)
}
