package internal

import (
	"os"
	"sync/atomic"
	"time"
)

var (
	consecutiveFailures int32
	lastSuccessUnix     int64
	metadataUpdates     uint64
	startTimeUnix       int64
	circuitOpenUntil    int64
)

func init() {
	startTimeUnix = time.Now().Unix()
}

func IncFailure() {
	atomic.AddInt32(&consecutiveFailures, 1)
}

func ResetFailure() {
	atomic.StoreInt32(&consecutiveFailures, 0)
	atomic.StoreInt64(&lastSuccessUnix, time.Now().Unix())
}

func GetFailures() int32 {
	return atomic.LoadInt32(&consecutiveFailures)
}

func GetLastSuccess() int64 {
	return atomic.LoadInt64(&lastSuccessUnix)
}

func IncMetadataUpdate() {
	atomic.AddUint64(&metadataUpdates, 1)
}

func GetMetadataUpdates() uint64 {
	return atomic.LoadUint64(&metadataUpdates)
}

func GetAgentStartTime() int64 {
	return startTimeUnix
}

func FileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

/* -----------------------------
   Circuit Breaker
--------------------------------*/

func IsCircuitOpen() bool {
	return time.Now().Unix() < atomic.LoadInt64(&circuitOpenUntil)
}

func OpenCircuit(d time.Duration) {
	atomic.StoreInt64(&circuitOpenUntil, time.Now().Add(d).Unix())
}
