package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"time"
)

// Aggregation holds the aggregate info of how our server performed against our
// requests. The aggregation may be accessed by several goroutines concurrently.
//
// Successes and Failures are stored as counters.
// TTFB and TTLB (time to [first|last] byte) are stored as running averages.
// They are `time.Duration`s, stored as an int64 nanosecond count.
type Aggregation struct {
	Successes int           `json:"successes"`
	Failures  int           `json:"failures"`
	TTFB      time.Duration `json:"ttfb"`
	TTLB      time.Duration `json:"ttlb"`
	mutex     sync.Mutex
}

// AddSuccess increments the aggregation's success counter, and adds the TTFB
// and TTLB to their respective moving averages. We use the successes as the
// divisor for our average.
func (a *Aggregation) AddSuccess(ttfb, ttlb time.Duration) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.Successes++
	a.TTFB = a.TTFB + (ttfb-a.TTFB)/time.Duration(a.Successes)
	a.TTLB = a.TTLB + (ttlb-a.TTLB)/time.Duration(a.Successes)
}

// AddFailure increments our aggregation's failure counter.
func (a *Aggregation) AddFailure() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.Failures++
}

// PrettyPrint pretty prints our aggregation.
func (a *Aggregation) PrettyPrint() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	total := a.Successes + a.Failures
	fmt.Printf(`
Total Requests: %d
     Successes: %d
      Failures: %d
  Success Rate: %.5f%%
  Failure Rate: %.5f%%
  Average TTFB: %v
  Average TTLB: %v
         Delta: %v
`,
		total,
		a.Successes,
		a.Failures,
		100*float64(a.Successes)/float64(total),
		100*float64(a.Failures)/float64(total),
		a.TTFB,
		a.TTLB,
		a.TTLB-a.TTFB,
	)
}

// Write writes our aggregation to a file.
func (a *Aggregation) Write(path string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	data, err := json.Marshal(a)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return err
	}
	return nil
}
