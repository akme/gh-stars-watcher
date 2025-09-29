package monitor

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// ProgressReporter handles progress reporting for monitoring operations
type ProgressReporter struct {
	writer io.Writer
	mu     sync.Mutex
}

// NewProgressReporter creates a new progress reporter
func NewProgressReporter(writer io.Writer) *ProgressReporter {
	if writer == nil {
		writer = os.Stdout
	}
	return &ProgressReporter{
		writer: writer,
	}
}

// ReportProgress reports progress during repository fetching
func (p *ProgressReporter) ReportProgress(ctx context.Context, current, total int, message string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if total == 0 {
		fmt.Fprintf(p.writer, "\r%s... %d repositories processed", message, current)
	} else {
		percentage := float64(current) / float64(total) * 100
		fmt.Fprintf(p.writer, "\r%s... %d/%d repositories (%.1f%%)", message, current, total, percentage)
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		fmt.Fprintf(p.writer, "\nOperation cancelled\n")
	default:
	}
}

// ReportCompletion reports completion of an operation
func (p *ProgressReporter) ReportCompletion(duration time.Duration, count int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	fmt.Fprintf(p.writer, "\nCompleted in %v - processed %d repositories\n", duration, count)
}

// ReportError reports an error during operation
func (p *ProgressReporter) ReportError(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	fmt.Fprintf(p.writer, "\nError: %v\n", err)
}

// ReportRateLimit reports rate limit information
func (p *ProgressReporter) ReportRateLimit(remaining, resetTime int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if remaining < 100 {
		resetDuration := time.Duration(resetTime) * time.Second
		fmt.Fprintf(p.writer, "\nRate limit warning: %d requests remaining (resets in %v)\n",
			remaining, resetDuration)
	}
}

// ProgressCallback is a function type for progress callbacks
type ProgressCallback func(current, total int, message string)

// NewProgressCallback creates a progress callback function
func NewProgressCallback(reporter *ProgressReporter, ctx context.Context) ProgressCallback {
	return func(current, total int, message string) {
		reporter.ReportProgress(ctx, current, total, message)
	}
}

// ProgressTracker tracks overall progress across multiple operations
type ProgressTracker struct {
	mu         sync.Mutex
	operations map[string]*OperationProgress
	reporter   *ProgressReporter
	startTime  time.Time
}

// OperationProgress tracks progress for a single operation
type OperationProgress struct {
	Name      string
	Current   int
	Total     int
	Completed bool
	Error     error
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(reporter *ProgressReporter) *ProgressTracker {
	return &ProgressTracker{
		operations: make(map[string]*OperationProgress),
		reporter:   reporter,
		startTime:  time.Now(),
	}
}

// StartOperation starts tracking a new operation
func (pt *ProgressTracker) StartOperation(name string, total int) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.operations[name] = &OperationProgress{
		Name:    name,
		Total:   total,
		Current: 0,
	}
}

// UpdateOperation updates progress for an operation
func (pt *ProgressTracker) UpdateOperation(name string, current int) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if op, exists := pt.operations[name]; exists {
		op.Current = current
		pt.reportOverallProgress()
	}
}

// CompleteOperation marks an operation as completed
func (pt *ProgressTracker) CompleteOperation(name string, err error) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if op, exists := pt.operations[name]; exists {
		op.Completed = true
		op.Error = err
		if err == nil {
			op.Current = op.Total
		}
	}
}

// reportOverallProgress reports the overall progress across all operations
func (pt *ProgressTracker) reportOverallProgress() {
	totalCurrent := 0
	totalExpected := 0
	completed := 0

	for _, op := range pt.operations {
		totalCurrent += op.Current
		totalExpected += op.Total
		if op.Completed {
			completed++
		}
	}

	message := fmt.Sprintf("Processing (%d/%d operations complete)", completed, len(pt.operations))
	if totalExpected > 0 {
		percentage := float64(totalCurrent) / float64(totalExpected) * 100
		fmt.Fprintf(pt.reporter.writer, "\r%s - %.1f%% complete", message, percentage)
	} else {
		fmt.Fprintf(pt.reporter.writer, "\r%s", message)
	}
}

// Finish completes all tracking and reports final results
func (pt *ProgressTracker) Finish() {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	duration := time.Since(pt.startTime)
	totalProcessed := 0
	errors := 0

	for _, op := range pt.operations {
		totalProcessed += op.Current
		if op.Error != nil {
			errors++
		}
	}

	fmt.Fprintf(pt.reporter.writer, "\n")
	pt.reporter.ReportCompletion(duration, totalProcessed)

	if errors > 0 {
		fmt.Fprintf(pt.reporter.writer, "Completed with %d errors\n", errors)
	}
}

// SpinnerProgress provides a simple spinner for indeterminate progress
type SpinnerProgress struct {
	writer   io.Writer
	spinner  []string
	index    int
	running  bool
	stopChan chan bool
	mu       sync.Mutex
}

// NewSpinnerProgress creates a new spinner progress indicator
func NewSpinnerProgress(writer io.Writer, message string) *SpinnerProgress {
	if writer == nil {
		writer = os.Stdout
	}

	sp := &SpinnerProgress{
		writer:   writer,
		spinner:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		stopChan: make(chan bool),
	}

	go sp.spin(message)
	return sp
}

// spin runs the spinner animation
func (sp *SpinnerProgress) spin(message string) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	sp.mu.Lock()
	sp.running = true
	sp.mu.Unlock()

	for {
		select {
		case <-sp.stopChan:
			return
		case <-ticker.C:
			sp.mu.Lock()
			if sp.running {
				fmt.Fprintf(sp.writer, "\r%s %s", sp.spinner[sp.index], message)
				sp.index = (sp.index + 1) % len(sp.spinner)
			}
			sp.mu.Unlock()
		}
	}
}

// Stop stops the spinner
func (sp *SpinnerProgress) Stop() {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	if sp.running {
		sp.running = false
		sp.stopChan <- true
		fmt.Fprintf(sp.writer, "\r")
	}
}
