package utils

import (
	"fmt"
	"time"
)

// Timer tracks execution time for verbose output
type Timer struct {
	verbose   bool
	startTime time.Time
	stepStart time.Time
	steps     []stepTiming
}

type stepTiming struct {
	name     string
	duration time.Duration
	total    time.Duration
}

// NewTimer creates a new Timer instance
func NewTimer(verbose bool) *Timer {
	now := time.Now()
	return &Timer{
		verbose:   verbose,
		startTime: now,
		stepStart: now,
		steps:     make([]stepTiming, 0),
	}
}

// Step logs a step completion with timing if verbose is enabled
func (t *Timer) Step(stepName string) {
	if !t.verbose {
		return
	}
	elapsed := time.Since(t.stepStart)
	total := time.Since(t.startTime)

	t.steps = append(t.steps, stepTiming{
		name:     stepName,
		duration: elapsed,
		total:    total,
	})

	t.stepStart = time.Now()
}

// PrintSummary prints a summary of all timed steps
func (t *Timer) PrintSummary() {
	if !t.verbose || len(t.steps) == 0 {
		return
	}

	fmt.Println("\n=== Timing Summary ===")
	for _, step := range t.steps {
		fmt.Printf("  %-40s %8v (total: %v)\n",
			step.name,
			step.duration.Round(time.Millisecond),
			step.total.Round(time.Millisecond))
	}

	totalElapsed := time.Since(t.startTime)
	fmt.Printf("  %-40s %8v\n", "Total execution time", totalElapsed.Round(time.Millisecond))
	fmt.Println("======================")
}
