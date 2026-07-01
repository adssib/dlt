// Package protocol defines the coordinator<->worker wire messages and a
// newline-delimited JSON codec over a net.Conn (see ADR-0004).
package protocol

import "encoding/json"

// MsgType tags an Envelope so the receiver knows how to decode the payload.
type MsgType string

const (
	MsgRegister  MsgType = "register"
	MsgProgress  MsgType = "progress"
	MsgResults   MsgType = "results"
	MsgStartTest MsgType = "start_test"
	MsgStopTest  MsgType = "stop_test"
)

// Envelope is the outer frame for every message: a type tag plus the raw JSON
// payload, decoded on demand via Decode.
type Envelope struct {
	Type    MsgType         `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Decode unmarshals the envelope payload into v (a pointer to the struct that
// matches Type).
func (e Envelope) Decode(v any) error {
	return json.Unmarshal(e.Payload, v)
}

// ---- Worker -> Coordinator ----

// Register announces a worker and its capacity when it connects.
type Register struct {
	WorkerID       string `json:"worker_id"`
	MaxConcurrency int    `json:"max_concurrency"`
}

// Progress is a live in-run update streamed by a worker.
type Progress struct {
	TestID    string `json:"test_id"`
	Completed int64  `json:"completed"`
	Errors    int64  `json:"errors"`
}

// Results is a worker's final tally for a run. Histogram is the serialized
// mergeable latency histogram (Phase 4); it stays nil during Phase 2's basic stats.
type Results struct {
	TestID     string `json:"test_id"`
	Histogram  []byte `json:"histogram,omitempty"`
	Total      int64  `json:"total"`
	Successful int64  `json:"successful"`
	Failed     int64  `json:"failed"`
	Throttled  int64  `json:"throttled"`
	DurationMS int64  `json:"duration_ms"`
}

// ---- Coordinator -> Worker ----

// StartTest is the broadcast that releases the start barrier (F3), carrying the
// per-worker slice of the workload.
type StartTest struct {
	TestID string     `json:"test_id"`
	Config TestConfig `json:"config"`
}

// StopTest asks a worker to abort the current run.
type StopTest struct {
	TestID string `json:"test_id"`
}

// TestConfig is one worker's share of the run, produced by the coordinator's planner.
type TestConfig struct {
	TargetURL            string `json:"target_url"`
	RequestsPerWorker    int    `json:"requests_per_worker"`
	ConcurrencyPerWorker int    `json:"concurrency_per_worker"`
	TimeoutMS            int    `json:"timeout_ms"`
	RampUpSeconds        int    `json:"ramp_up_seconds"`
}
