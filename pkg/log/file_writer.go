package log

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// FileWriter 本地文件写入器，实现Writer接口
// 负责将日志写入本地文件

type FileWriter struct {
	filePath string
	file     *os.File
	mu       sync.Mutex
}

// NewFileWriter 创建本地文件写入器
func NewFileWriter(filePath string) (*FileWriter, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	return &FileWriter{filePath: filePath, file: file}, nil
}

// Write 实现Writer接口，将日志写入本地文件
func (fw *FileWriter) Write(entry *LogEntry) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.file == nil {
		file, err := os.OpenFile(fw.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to reopen file %s: %w", fw.filePath, err)
		}
		fw.file = file
	}

	// 格式化日志条目并添加时间戳
	line, err := FormatLogEntry(entry)
	if err != nil {
		return fmt.Errorf("failed to format log entry: %w", err)
	}

	// 添加本地时间戳前缀
	timestamp := time.Unix(entry.Time, 0).Format("2006-01-02 15:04:05")
	fullLine := fmt.Sprintf("[%s] %s", timestamp, line)

	_, err = fw.file.WriteString(fullLine + "\n")
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	// 刷新文件缓冲区
	err = fw.file.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	return nil
}

// Close 关闭文件
func (fw *FileWriter) Close() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.file != nil {
		err := fw.file.Close()
		fw.file = nil
		return err
	}
	return nil
}
