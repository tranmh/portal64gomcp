package logger

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// RotationManager manages log rotation with enhanced features
type RotationManager struct {
	config      RotationConfig
	lumberjacks map[string]*lumberjack.Logger
	mutex       sync.RWMutex
	
	// Compression management
	compressor *CompressionManager
}

// CompressionManager handles delayed compression of log files
type CompressionManager struct {
	config      RotationConfig
	stopChan    chan struct{}
	wg          sync.WaitGroup
	running     bool
	mutex       sync.Mutex
}

// NewRotationManager creates a new rotation manager
func NewRotationManager(config RotationConfig) *RotationManager {
	rm := &RotationManager{
		config:      config,
		lumberjacks: make(map[string]*lumberjack.Logger),
		compressor:  NewCompressionManager(config),
	}
	
	// Start compression manager if compression is enabled
	if config.Compress && config.CompressAfter > 0 {
		rm.compressor.Start()
	}
	
	return rm
}

// GetWriter returns a lumberjack writer for the specified log file
func (rm *RotationManager) GetWriter(logPath string) io.Writer {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	if writer, exists := rm.lumberjacks[logPath]; exists {
		return writer
	}
	
	// Ensure directory exists
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		// Return a fallback writer if directory creation fails
		return os.Stderr
	}
	
	writer := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    rm.config.MaxSize,
		MaxAge:     rm.config.MaxAge,
		MaxBackups: rm.config.MaxBackups,
		LocalTime:  true,
		Compress:   false, // We handle compression separately for better control
	}
	
	rm.lumberjacks[logPath] = writer
	return writer
}

// Rotate triggers rotation for all managed log files
func (rm *RotationManager) Rotate() error {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	var lastErr error
	for _, writer := range rm.lumberjacks {
		if err := writer.Rotate(); err != nil {
			lastErr = err
		}
	}
	
	return lastErr
}

// Close closes all managed writers and stops the compression manager
func (rm *RotationManager) Close() error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	// Stop compression manager
	rm.compressor.Stop()
	
	// Close all lumberjack writers
	var lastErr error
	for _, writer := range rm.lumberjacks {
		if err := writer.Close(); err != nil {
			lastErr = err
		}
	}
	
	return lastErr
}

// NewCompressionManager creates a new compression manager
func NewCompressionManager(config RotationConfig) *CompressionManager {
	return &CompressionManager{
		config:   config,
		stopChan: make(chan struct{}),
	}
}

// Start starts the compression manager
func (cm *CompressionManager) Start() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	if cm.running {
		return
	}
	
	cm.running = true
	cm.wg.Add(1)
	go cm.run()
}

// Stop stops the compression manager
func (cm *CompressionManager) Stop() {
	cm.mutex.Lock()
	running := cm.running
	if running {
		cm.running = false
		close(cm.stopChan)
	}
	cm.mutex.Unlock()
	
	if running {
		cm.wg.Wait()
	}
}

// run is the main compression loop
func (cm *CompressionManager) run() {
	defer cm.wg.Done()
	
	// Run compression check every hour
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	// Run initial compression check
	cm.compressOldFiles()
	
	for {
		select {
		case <-ticker.C:
			cm.compressOldFiles()
		case <-cm.stopChan:
			return
		}
	}
}

// compressOldFiles compresses log files older than the configured threshold
func (cm *CompressionManager) compressOldFiles() {
	// Find all .log files in common log directories
	logDirs := []string{"logs/app", "logs/access", "logs/error", "logs/metrics"}
	
	for _, dir := range logDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		
		files, err := filepath.Glob(filepath.Join(dir, "*.log.*"))
		if err != nil {
			continue
		}
		
		for _, file := range files {
			// Skip already compressed files
			if strings.HasSuffix(file, ".gz") {
				continue
			}
			
			// Check file age
			info, err := os.Stat(file)
			if err != nil {
				continue
			}
			
			age := time.Since(info.ModTime())
			if age >= time.Duration(cm.config.CompressAfter)*24*time.Hour {
				if err := cm.compressFile(file); err != nil {
					// Log compression error (using standard logger to avoid recursion)
					fmt.Fprintf(os.Stderr, "Failed to compress log file %s: %v\n", file, err)
				}
			}
		}
	}
}

// compressFile compresses a single log file
func (cm *CompressionManager) compressFile(filename string) error {
	// Open source file
	src, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()
	
	// Create compressed file
	dstPath := filename + ".gz"
	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create compressed file: %w", err)
	}
	defer dst.Close()
	
	// Create gzip writer
	gzWriter := gzip.NewWriter(dst)
	defer gzWriter.Close()
	
	// Copy and compress
	_, err = io.Copy(gzWriter, src)
	if err != nil {
		os.Remove(dstPath) // Clean up on error
		return fmt.Errorf("failed to compress file: %w", err)
	}
	
	// Close gzip writer to flush data
	if err := gzWriter.Close(); err != nil {
		os.Remove(dstPath)
		return fmt.Errorf("failed to finalize compressed file: %w", err)
	}
	
	// Copy file permissions
	srcInfo, _ := src.Stat()
	os.Chmod(dstPath, srcInfo.Mode())
	
	// Remove original file
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("failed to remove original file: %w", err)
	}
	
	return nil
}

// FileHook is a logrus hook that writes to rotating files
type FileHook struct {
	rotationManager *RotationManager
	logPath         string
	levels          []logrus.Level
	formatter       logrus.Formatter
	writer          io.Writer
}

// NewFileHook creates a new file hook with rotation
func NewFileHook(rotationManager *RotationManager, logPath string, levels []logrus.Level, formatter logrus.Formatter) *FileHook {
	hook := &FileHook{
		rotationManager: rotationManager,
		logPath:         logPath,
		levels:          levels,
		formatter:       formatter,
	}
	
	// Get the rotating writer
	hook.writer = rotationManager.GetWriter(logPath)
	
	return hook
}

// Levels returns the levels this hook handles
func (hook *FileHook) Levels() []logrus.Level {
	if len(hook.levels) == 0 {
		return logrus.AllLevels
	}
	return hook.levels
}

// Fire writes the log entry to the file
func (hook *FileHook) Fire(entry *logrus.Entry) error {
	// Format the entry
	msg, err := hook.formatter.Format(entry)
	if err != nil {
		return fmt.Errorf("failed to format log entry: %w", err)
	}
	
	// Write to file
	_, err = hook.writer.Write(msg)
	return err
}

// SeparatedFileHook manages multiple file hooks for log separation
type SeparatedFileHook struct {
	hooks map[string]*FileHook
}

// NewSeparatedFileHook creates hooks for separated logging
func NewSeparatedFileHook(rotationManager *RotationManager, config *Config, formatter logrus.Formatter) *SeparatedFileHook {
	separated := &SeparatedFileHook{
		hooks: make(map[string]*FileHook),
	}
	
	// Main application log (all levels)
	if appPath := config.GetLogFilePath("app"); appPath != "" {
		separated.hooks["app"] = NewFileHook(
			rotationManager, 
			appPath, 
			logrus.AllLevels, 
			formatter,
		)
	}
	
	// Error log (error and above)
	if errorPath := config.GetLogFilePath("error"); errorPath != "" {
		separated.hooks["error"] = NewFileHook(
			rotationManager, 
			errorPath, 
			[]logrus.Level{logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel}, 
			formatter,
		)
	}
	
	// Access log will be handled separately in middleware
	// Metrics log will be handled separately in metrics collection
	
	return separated
}

// Levels returns all levels (we'll filter in Fire method)
func (hook *SeparatedFileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire distributes log entries to appropriate hooks
func (hook *SeparatedFileHook) Fire(entry *logrus.Entry) error {
	var lastErr error
	
	for _, fileHook := range hook.hooks {
		// Check if this hook handles this level
		shouldHandle := false
		for _, level := range fileHook.Levels() {
			if level == entry.Level {
				shouldHandle = true
				break
			}
		}
		
		if shouldHandle {
			if err := fileHook.Fire(entry); err != nil {
				lastErr = err
			}
		}
	}
	
	return lastErr
}

// GetRotationStats returns statistics about log rotation
func (rm *RotationManager) GetRotationStats() map[string]RotationStats {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	stats := make(map[string]RotationStats)
	
	for path, writer := range rm.lumberjacks {
		// Get current file info
		info, err := os.Stat(writer.Filename)
		var size int64
		var modTime time.Time
		if err == nil {
			size = info.Size()
			modTime = info.ModTime()
		}
		
		// Count rotated files
		dir := filepath.Dir(writer.Filename)
		base := filepath.Base(writer.Filename)
		pattern := filepath.Join(dir, base+".*")
		files, _ := filepath.Glob(pattern)
		
		// Separate compressed and uncompressed
		var compressed, uncompressed int
		for _, file := range files {
			if strings.HasSuffix(file, ".gz") {
				compressed++
			} else {
				uncompressed++
			}
		}
		
		stats[path] = RotationStats{
			CurrentSize:       size,
			LastModified:      modTime,
			RotatedFiles:      len(files),
			CompressedFiles:   compressed,
			UncompressedFiles: uncompressed,
		}
	}
	
	return stats
}

// RotationStats holds statistics about log rotation
type RotationStats struct {
	CurrentSize       int64     `json:"current_size"`
	LastModified      time.Time `json:"last_modified"`
	RotatedFiles      int       `json:"rotated_files"`
	CompressedFiles   int       `json:"compressed_files"`
	UncompressedFiles int       `json:"uncompressed_files"`
}
