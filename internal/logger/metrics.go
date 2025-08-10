package logger

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// MetricsManager manages performance and operational metrics
type MetricsManager struct {
	// Performance metrics
	httpMetrics     *HTTPMetrics
	logMetrics      *LogMetrics
	systemMetrics   *SystemMetrics
	
	// Collection settings
	enabled         bool
	collectionStart time.Time
	
	mutex sync.RWMutex
}

// HTTPMetrics tracks HTTP request statistics
type HTTPMetrics struct {
	TotalRequests      int64             `json:"total_requests"`
	TotalDuration      int64             `json:"total_duration_ns"`
	StatusCounts       map[int]int64     `json:"status_counts"`
	MethodCounts       map[string]int64  `json:"method_counts"`
	PathCounts         map[string]int64  `json:"path_counts"`
	ErrorCount         int64             `json:"error_count"`
	AvgResponseTime    float64           `json:"avg_response_time_ms"`
	MaxResponseTime    int64             `json:"max_response_time_ms"`
	MinResponseTime    int64             `json:"min_response_time_ms"`
	RequestsPerSecond  float64           `json:"requests_per_second"`
	BytesTransferred   int64             `json:"bytes_transferred"`
	ActiveConnections  int32             `json:"active_connections"`
	
	mutex sync.RWMutex
}

// LogMetrics tracks logging system performance
type LogMetrics struct {
	TotalLogs          int64             `json:"total_logs"`
	LogsByLevel        map[string]int64  `json:"logs_by_level"`
	LogsByComponent    map[string]int64  `json:"logs_by_component"`
	AsyncBufferSize    int32             `json:"async_buffer_size"`
	AsyncBufferUsage   float64           `json:"async_buffer_usage_percent"`
	BufferOverflows    int64             `json:"buffer_overflows"`
	WriteErrors        int64             `json:"write_errors"`
	FlushCount         int64             `json:"flush_count"`
	LastFlushTime      time.Time         `json:"last_flush_time"`
	AverageWriteTime   time.Duration     `json:"average_write_time"`
	FileSizes          map[string]int64  `json:"file_sizes"`
	RotationCount      int64             `json:"rotation_count"`
	CompressionSaved   int64             `json:"compression_saved_bytes"`
	
	mutex sync.RWMutex
}

// SystemMetrics tracks system-level metrics
type SystemMetrics struct {
	StartTime          time.Time         `json:"start_time"`
	Uptime             time.Duration     `json:"uptime"`
	MemoryUsage        int64             `json:"memory_usage_bytes"`
	GoroutineCount     int               `json:"goroutine_count"`
	CPUUsagePercent    float64           `json:"cpu_usage_percent"`
	DiskSpaceUsed      int64             `json:"disk_space_used_bytes"`
	DiskSpaceAvailable int64             `json:"disk_space_available_bytes"`
	
	mutex sync.RWMutex
}

// NewMetricsManager creates a new metrics manager
func NewMetricsManager(enabled bool) *MetricsManager {
	return &MetricsManager{
		httpMetrics: &HTTPMetrics{
			StatusCounts: make(map[int]int64),
			MethodCounts: make(map[string]int64),
			PathCounts:   make(map[string]int64),
			MinResponseTime: int64(^uint64(0) >> 1), // Max int64
		},
		logMetrics: &LogMetrics{
			LogsByLevel:     make(map[string]int64),
			LogsByComponent: make(map[string]int64),
			FileSizes:       make(map[string]int64),
		},
		systemMetrics: &SystemMetrics{
			StartTime: time.Now(),
		},
		enabled:         enabled,
		collectionStart: time.Now(),
	}
}

// RecordHTTPRequest records metrics for an HTTP request
func (mm *MetricsManager) RecordHTTPRequest(method, path string, statusCode int, duration time.Duration, bytesWritten int64) {
	if !mm.enabled {
		return
	}
	
	durationMs := int64(duration / time.Millisecond)
	durationNs := int64(duration)
	
	mm.httpMetrics.mutex.Lock()
	defer mm.httpMetrics.mutex.Unlock()
	
	// Update counters
	atomic.AddInt64(&mm.httpMetrics.TotalRequests, 1)
	atomic.AddInt64(&mm.httpMetrics.TotalDuration, durationNs)
	atomic.AddInt64(&mm.httpMetrics.BytesTransferred, bytesWritten)
	
	// Update status counts
	mm.httpMetrics.StatusCounts[statusCode]++
	
	// Update method counts
	mm.httpMetrics.MethodCounts[method]++
	
	// Update path counts (limit to prevent memory explosion)
	if len(mm.httpMetrics.PathCounts) < 1000 {
		mm.httpMetrics.PathCounts[path]++
	}
	
	// Update response time statistics
	if durationMs > mm.httpMetrics.MaxResponseTime {
		mm.httpMetrics.MaxResponseTime = durationMs
	}
	if durationMs < mm.httpMetrics.MinResponseTime {
		mm.httpMetrics.MinResponseTime = durationMs
	}
	
	// Update error count
	if statusCode >= 400 {
		atomic.AddInt64(&mm.httpMetrics.ErrorCount, 1)
	}
	
	// Calculate derived metrics
	totalReqs := atomic.LoadInt64(&mm.httpMetrics.TotalRequests)
	totalDur := atomic.LoadInt64(&mm.httpMetrics.TotalDuration)
	
	if totalReqs > 0 {
		mm.httpMetrics.AvgResponseTime = float64(totalDur) / float64(totalReqs) / float64(time.Millisecond)
		
		elapsed := time.Since(mm.collectionStart)
		if elapsed > 0 {
			mm.httpMetrics.RequestsPerSecond = float64(totalReqs) / elapsed.Seconds()
		}
	}
}

// RecordLogEntry records metrics for a log entry
func (mm *MetricsManager) RecordLogEntry(level logrus.Level, component string, writeTime time.Duration) {
	if !mm.enabled {
		return
	}
	
	mm.logMetrics.mutex.Lock()
	defer mm.logMetrics.mutex.Unlock()
	
	// Update counters
	atomic.AddInt64(&mm.logMetrics.TotalLogs, 1)
	
	// Update level counts
	levelStr := level.String()
	mm.logMetrics.LogsByLevel[levelStr]++
	
	// Update component counts
	if component != "" {
		mm.logMetrics.LogsByComponent[component]++
	}
	
	// Update average write time
	if writeTime > 0 {
		// Simple moving average calculation
		totalLogs := atomic.LoadInt64(&mm.logMetrics.TotalLogs)
		if totalLogs == 1 {
			mm.logMetrics.AverageWriteTime = writeTime
		} else {
			// Exponential moving average with alpha = 0.1
			alpha := 0.1
			mm.logMetrics.AverageWriteTime = time.Duration(
				float64(mm.logMetrics.AverageWriteTime)*(1-alpha) + 
				float64(writeTime)*alpha,
			)
		}
	}
}

// RecordAsyncMetrics records async logging metrics
func (mm *MetricsManager) RecordAsyncMetrics(bufferSize, bufferCapacity int32, overflows, flushes int64) {
	if !mm.enabled {
		return
	}
	
	mm.logMetrics.mutex.Lock()
	defer mm.logMetrics.mutex.Unlock()
	
	mm.logMetrics.AsyncBufferSize = bufferSize
	if bufferCapacity > 0 {
		mm.logMetrics.AsyncBufferUsage = float64(bufferSize) / float64(bufferCapacity) * 100
	}
	
	atomic.StoreInt64(&mm.logMetrics.BufferOverflows, overflows)
	atomic.StoreInt64(&mm.logMetrics.FlushCount, flushes)
	mm.logMetrics.LastFlushTime = time.Now()
}

// RecordRotationEvent records a log rotation event
func (mm *MetricsManager) RecordRotationEvent() {
	if !mm.enabled {
		return
	}
	
	atomic.AddInt64(&mm.logMetrics.RotationCount, 1)
}

// RecordCompressionSaved records bytes saved by compression
func (mm *MetricsManager) RecordCompressionSaved(bytesSaved int64) {
	if !mm.enabled {
		return
	}
	
	atomic.AddInt64(&mm.logMetrics.CompressionSaved, bytesSaved)
}

// UpdateFileSizes updates log file size metrics
func (mm *MetricsManager) UpdateFileSizes(fileSizes map[string]int64) {
	if !mm.enabled {
		return
	}
	
	mm.logMetrics.mutex.Lock()
	defer mm.logMetrics.mutex.Unlock()
	
	mm.logMetrics.FileSizes = fileSizes
}

// IncrementActiveConnections increments active HTTP connections
func (mm *MetricsManager) IncrementActiveConnections() {
	if !mm.enabled {
		return
	}
	
	atomic.AddInt32(&mm.httpMetrics.ActiveConnections, 1)
}

// DecrementActiveConnections decrements active HTTP connections
func (mm *MetricsManager) DecrementActiveConnections() {
	if !mm.enabled {
		return
	}
	
	atomic.AddInt32(&mm.httpMetrics.ActiveConnections, -1)
}

// GetHTTPMetrics returns current HTTP metrics
func (mm *MetricsManager) GetHTTPMetrics() HTTPMetrics {
	mm.httpMetrics.mutex.RLock()
	defer mm.httpMetrics.mutex.RUnlock()
	
	// Create a copy to avoid race conditions
	metrics := HTTPMetrics{
		TotalRequests:      atomic.LoadInt64(&mm.httpMetrics.TotalRequests),
		TotalDuration:      atomic.LoadInt64(&mm.httpMetrics.TotalDuration),
		StatusCounts:       make(map[int]int64),
		MethodCounts:       make(map[string]int64),
		PathCounts:         make(map[string]int64),
		ErrorCount:         atomic.LoadInt64(&mm.httpMetrics.ErrorCount),
		AvgResponseTime:    mm.httpMetrics.AvgResponseTime,
		MaxResponseTime:    mm.httpMetrics.MaxResponseTime,
		MinResponseTime:    mm.httpMetrics.MinResponseTime,
		RequestsPerSecond:  mm.httpMetrics.RequestsPerSecond,
		BytesTransferred:   atomic.LoadInt64(&mm.httpMetrics.BytesTransferred),
		ActiveConnections:  atomic.LoadInt32(&mm.httpMetrics.ActiveConnections),
	}
	
	// Copy maps
	for k, v := range mm.httpMetrics.StatusCounts {
		metrics.StatusCounts[k] = v
	}
	for k, v := range mm.httpMetrics.MethodCounts {
		metrics.MethodCounts[k] = v
	}
	for k, v := range mm.httpMetrics.PathCounts {
		metrics.PathCounts[k] = v
	}
	
	return metrics
}

// GetLogMetrics returns current log metrics
func (mm *MetricsManager) GetLogMetrics() LogMetrics {
	mm.logMetrics.mutex.RLock()
	defer mm.logMetrics.mutex.RUnlock()
	
	// Create a copy to avoid race conditions
	metrics := LogMetrics{
		TotalLogs:          atomic.LoadInt64(&mm.logMetrics.TotalLogs),
		LogsByLevel:        make(map[string]int64),
		LogsByComponent:    make(map[string]int64),
		AsyncBufferSize:    mm.logMetrics.AsyncBufferSize,
		AsyncBufferUsage:   mm.logMetrics.AsyncBufferUsage,
		BufferOverflows:    atomic.LoadInt64(&mm.logMetrics.BufferOverflows),
		WriteErrors:        atomic.LoadInt64(&mm.logMetrics.WriteErrors),
		FlushCount:         atomic.LoadInt64(&mm.logMetrics.FlushCount),
		LastFlushTime:      mm.logMetrics.LastFlushTime,
		AverageWriteTime:   mm.logMetrics.AverageWriteTime,
		FileSizes:          make(map[string]int64),
		RotationCount:      atomic.LoadInt64(&mm.logMetrics.RotationCount),
		CompressionSaved:   atomic.LoadInt64(&mm.logMetrics.CompressionSaved),
	}
	
	// Copy maps
	for k, v := range mm.logMetrics.LogsByLevel {
		metrics.LogsByLevel[k] = v
	}
	for k, v := range mm.logMetrics.LogsByComponent {
		metrics.LogsByComponent[k] = v
	}
	for k, v := range mm.logMetrics.FileSizes {
		metrics.FileSizes[k] = v
	}
	
	return metrics
}

// GetSystemMetrics returns current system metrics
func (mm *MetricsManager) GetSystemMetrics() SystemMetrics {
	mm.systemMetrics.mutex.RLock()
	defer mm.systemMetrics.mutex.RUnlock()
	
	metrics := *mm.systemMetrics
	metrics.Uptime = time.Since(metrics.StartTime)
	
	return metrics
}

// GetAllMetrics returns all metrics in a single structure
func (mm *MetricsManager) GetAllMetrics() map[string]interface{} {
	return map[string]interface{}{
		"http":   mm.GetHTTPMetrics(),
		"logs":   mm.GetLogMetrics(),
		"system": mm.GetSystemMetrics(),
		"collection_enabled": mm.enabled,
		"collection_uptime":  time.Since(mm.collectionStart).String(),
	}
}

// ResetMetrics resets all collected metrics
func (mm *MetricsManager) ResetMetrics() {
	if !mm.enabled {
		return
	}
	
	mm.httpMetrics.mutex.Lock()
	mm.logMetrics.mutex.Lock()
	mm.systemMetrics.mutex.Lock()
	
	defer mm.httpMetrics.mutex.Unlock()
	defer mm.logMetrics.mutex.Unlock()
	defer mm.systemMetrics.mutex.Unlock()
	
	// Reset HTTP metrics
	atomic.StoreInt64(&mm.httpMetrics.TotalRequests, 0)
	atomic.StoreInt64(&mm.httpMetrics.TotalDuration, 0)
	atomic.StoreInt64(&mm.httpMetrics.ErrorCount, 0)
	atomic.StoreInt64(&mm.httpMetrics.BytesTransferred, 0)
	atomic.StoreInt32(&mm.httpMetrics.ActiveConnections, 0)
	
	mm.httpMetrics.StatusCounts = make(map[int]int64)
	mm.httpMetrics.MethodCounts = make(map[string]int64)
	mm.httpMetrics.PathCounts = make(map[string]int64)
	mm.httpMetrics.MaxResponseTime = 0
	mm.httpMetrics.MinResponseTime = int64(^uint64(0) >> 1)
	mm.httpMetrics.AvgResponseTime = 0
	mm.httpMetrics.RequestsPerSecond = 0
	
	// Reset log metrics
	atomic.StoreInt64(&mm.logMetrics.TotalLogs, 0)
	atomic.StoreInt64(&mm.logMetrics.BufferOverflows, 0)
	atomic.StoreInt64(&mm.logMetrics.WriteErrors, 0)
	atomic.StoreInt64(&mm.logMetrics.FlushCount, 0)
	atomic.StoreInt64(&mm.logMetrics.RotationCount, 0)
	atomic.StoreInt64(&mm.logMetrics.CompressionSaved, 0)
	
	mm.logMetrics.LogsByLevel = make(map[string]int64)
	mm.logMetrics.LogsByComponent = make(map[string]int64)
	mm.logMetrics.FileSizes = make(map[string]int64)
	mm.logMetrics.AverageWriteTime = 0
	
	// Reset collection start time
	mm.collectionStart = time.Now()
}

// Enable enables metrics collection
func (mm *MetricsManager) Enable() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	mm.enabled = true
	mm.collectionStart = time.Now()
}

// Disable disables metrics collection
func (mm *MetricsManager) Disable() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	mm.enabled = false
}

// IsEnabled returns whether metrics collection is enabled
func (mm *MetricsManager) IsEnabled() bool {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	
	return mm.enabled
}
