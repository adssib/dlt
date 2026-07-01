package coordinator

import "github.com/adssib/dlt/internal/protocol"

// Report is the aggregated outcome of a run across all workers. Phase 2 carries
// basic counts + throughput; correct latency percentiles arrive in Phase 4 when
// Results carries merged histograms (ADR-0002).
type Report struct {
	Total      int64
	Successful int64
	Failed     int64
	Throttled  int64
	DurationMS int64
	Partial    bool // some workers died mid-run (F11)
}

// Aggregate folds each worker's Results into one Report. If fewer results arrive
// than workers that were sent StartTest, the report is marked Partial (F11).
//
// TODO(you): sum the counts across results; set Partial when len(results) < expectedWorkers.
func Aggregate(results []protocol.Results, expectedWorkers int) Report {
	return Report{} // TODO(you)
}

// Render formats the report for output (text or json, per report.format).
//
// TODO(you): render a human-readable summary (and a json branch).
func (r Report) Render(format string) string {
	return "" // TODO(you)
}
