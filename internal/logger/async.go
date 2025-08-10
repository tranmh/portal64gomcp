package logger

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// AsyncWriter manages async log writing with buffering
type AsyncWriter struct {
	buffer          chan *logrus.Entry
	writers         []logrus.Hook
	flushInterval   time.Duration
	shutdownTimeout time.Duration
	
	// Control channels
	flushChan    chan struct{}
	shutdownChan chan struct{}
	doneChan     chan struct{}
	
	// State
	running      int32
	wg           sync.WaitGroup
	metrics      *AsyncMetrics
}

// AsyncMetrics tracks async writer performance
type AsyncMetrics struct {
	BufferSize      int32 `json:"buffer_size"`
	BufferCapacity  int32 `json:"buffer_capacity"`
	TotalWrites     int64 `json:"total_writes"`
	BufferOverflows int64 `json:"buffer_overflows"`
	FlushCount      int64 `json:"flush_count"`
	LastFlushTime   time.Time `json:"last_flush_time"`
}

// NewAsyncWriter creates a new async writer
func NewAsyncWriter(config AsyncConfig) *AsyncWriter {
	writer := &AsyncWriter{
		buffer:          make(chan *logrus.Entry, config.BufferSize),
		writers:         make([]logrus.Hook, 0),
		flushInterval:   config.FlushInterval,
		shutdownTimeout: config.ShutdownTimeout,
		flushChan:       make(chan struct{}, 1),
		shutdownChan:    make(chan struct{}, 1),
		doneChan:        make(chan struct{}),
		metrics: &AsyncMetrics{
			BufferCapacity: int32(config.BufferSize),
		},
	}
	
	return writer
}

// AddWriter adds a writer to the async writer
func (aw *AsyncWriter) AddWriter(hook logrus.Hook) {
	aw.writers = append(aw.writers, hook)
}

// Start starts the async writer
func (aw *AsyncWriter) Start() {
	if !atomic.CompareAndSwapInt32(&aw.running, 0, 1) {
		return // Already running
	}
	
	aw.wg.Add(1)
	go aw.run()
}

// Stop stops the async writer gracefully
func (aw *AsyncWriter) Stop() {
	if !atomic.CompareAndSwapInt32(&aw.running, 1, 0) {
		return // Already stopped
	}
	
	select {
	case aw.shutdownChan <- struct{}{}:
	default:
	}
	
	// Wait for shutdown with timeout
	select {
	case <-aw.doneChan:
	case <-time.After(aw.shutdownTimeout):
		// Force close after timeout
	}
	
	aw.wg.Wait()
}

// Write writes an entry asynchronously
func (aw *AsyncWriter) Write(entry *logrus.Entry) error {
	if atomic.LoadInt32(&aw.running) == 0 {
		// If async writer is stopped, write synchronously
		return aw.writeSync(entry)
	}
	
	select {
	case aw.buffer <- entry:
		atomic.AddInt32(&aw.metrics.BufferSize, 1)
		atomic.AddInt64(&aw.metrics.TotalWrites, 1)
		return nil
	default:
		// Buffer is full, record overflow and write synchronously
		atomic.AddInt64(&aw.metrics.BufferOverflows, 1)
		return aw.writeSync(entry)
	}
}

// Flush triggers an immediate flush
func (aw *AsyncWriter) Flush() {
	if atomic.LoadInt32(&aw.running) == 0 {
		return
	}
	
	select {
	case aw.flushChan <- struct{}{}:
	default:
		// Flush already pending
	}
}

// GetMetrics returns current metrics
func (aw *AsyncWriter) GetMetrics() AsyncMetrics {
	return AsyncMetrics{
		BufferSize:      atomic.LoadInt32(&aw.metrics.BufferSize),
		BufferCapacity:  aw.metrics.BufferCapacity,
		TotalWrites:     atomic.LoadInt64(&aw.metrics.TotalWrites),
		BufferOverflows: atomic.LoadInt64(&aw.metrics.BufferOverflows),
		FlushCount:      atomic.LoadInt64(&aw.metrics.FlushCount),
		LastFlushTime:   aw.metrics.LastFlushTime,
	}
}

// run is the main async writer loop
func (aw *AsyncWriter) run() {
	defer aw.wg.Done()
	defer close(aw.doneChan)
	
	ticker := time.NewTicker(aw.flushInterval)
	defer ticker.Stop()
	
	entries := make([]*logrus.Entry, 0, 100)
	
	for {
		select {
		case entry := <-aw.buffer:
			entries = append(entries, entry)
			atomic.AddInt32(&aw.metrics.BufferSize, -1)
			
			// Batch process entries if buffer has more
			if len(entries) < cap(entries) {
				continue
			}
			aw.flushEntries(entries)
			entries = entries[:0]
			
		case <-ticker.C:
			if len(entries) > 0 {
				aw.flushEntries(entries)
				entries = entries[:0]
			}
			
		case <-aw.flushChan:
			if len(entries) > 0 {
				aw.flushEntries(entries)
				entries = entries[:0]
			}
			
		case <-aw.shutdownChan:
			// Flush remaining entries on shutdown
			for len(aw.buffer) > 0 {
				entries = append(entries, <-aw.buffer)
				atomic.AddInt32(&aw.metrics.BufferSize, -1)
			}
			
			if len(entries) > 0 {
				aw.flushEntries(entries)
			}
			
			return
		}
	}
}

// flushEntries writes entries to all configured writers
func (aw *AsyncWriter) flushEntries(entries []*logrus.Entry) {
	if len(entries) == 0 {
		return
	}
	
	for _, entry := range entries {
		aw.writeSync(entry)
	}
	
	atomic.AddInt64(&aw.metrics.FlushCount, 1)
	aw.metrics.LastFlushTime = time.Now()
}

// writeSync writes entry synchronously to all writers
func (aw *AsyncWriter) writeSync(entry *logrus.Entry) error {
	var lastErr error
	
	for _, writer := range aw.writers {
		if err := writer.Fire(entry); err != nil {
			lastErr = err
		}
	}
	
	return lastErr
}

// AsyncHook implements logrus.Hook for async writing
type AsyncHook struct {
	writer *AsyncWriter
	levels []logrus.Level
}

// NewAsyncHook creates a new async hook
func NewAsyncHook(writer *AsyncWriter, levels []logrus.Level) *AsyncHook {
	return &AsyncHook{
		writer: writer,
		levels: levels,
	}
}

// Levels returns the levels this hook is interested in
func (hook *AsyncHook) Levels() []logrus.Level {
	if len(hook.levels) == 0 {
		return logrus.AllLevels
	}
	return hook.levels
}

// Fire is called when a log entry is written
func (hook *AsyncHook) Fire(entry *logrus.Entry) error {
	// Clone the entry to avoid race conditions
	cloned := &logrus.Entry{
		Logger:  entry.Logger,
		Data:    make(logrus.Fields),
		Time:    entry.Time,
		Level:   entry.Level,
		Caller:  entry.Caller,
		Message: entry.Message,
		Buffer:  nil, // Don't copy buffer
	}
	
	// Copy data fields
	for k, v := range entry.Data {
		cloned.Data[k] = v
	}
	
	return hook.writer.Write(cloned)
}
