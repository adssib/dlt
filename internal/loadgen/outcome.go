// Package loadgen is the hand-written HTTP load engine (ADR-0006): a
// bounded-concurrency request generator that measures every request and
// classifies its outcome.
package loadgen

import "net/http"

// Outcome classifies the result of a single request. Throttled (429) is tracked
// separately from Failed so the tester can prove a rate limiter trips (F13).
type Outcome int

const (
	OutcomeSuccess Outcome = iota
	OutcomeThrottled
	OutcomeFailed
)

// Classify maps an HTTP response/error to an Outcome:
//
//	transport error           -> Failed
//	status 429                -> Throttled
//	status >= 500             -> Failed
//	otherwise                 -> Success
//
// TODO(you): implement the classification.
func Classify(resp *http.Response, err error) Outcome {
	return OutcomeSuccess // TODO(you)
}
