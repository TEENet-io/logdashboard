package log

import (
	"sync"
	"time"
)

// 日志主入口，暴露统一API

var (
	loggers []Writer
	cfg     Config
	mu      sync.Mutex
)

// 日志级别优先级
var levelPriority = map[string]int{
	"debug": 0,
	"info":  1,
	"warn":  2,
	"error": 3,
}

// Init 初始化日志模块，配置本地文件、Loki、标签等
func Init(c Config) {
	mu.Lock()
	defer mu.Unlock()
	cfg = c
	loggers = []Writer{}

	// 设置默认级别
	if cfg.Level == "" {
		cfg.Level = "info"
	}

	// 初始化本地文件写入器
	if c.FilePath != "" {
		fw, err := NewFileWriter(c.FilePath)
		if err == nil {
			loggers = append(loggers, fw)
		}
	}
	// 初始化Loki写入器
	if c.LokiURL != "" {
		lw := NewLokiWriter(c.LokiURL, c.Labels)
		loggers = append(loggers, lw)
	}
}

// Debug 打印Debug级别日志
func Debug(msg string, fields ...Field) {
	logWithLevel("debug", msg, fields...)
}

// Info 打印Info级别日志
func Info(msg string, fields ...Field) {
	logWithLevel("info", msg, fields...)
}

// Warn 打印Warn级别日志
func Warn(msg string, fields ...Field) {
	logWithLevel("warn", msg, fields...)
}

// Error 打印Error级别日志
func Error(msg string, fields ...Field) {
	logWithLevel("error", msg, fields...)
}

// shouldLog 检查是否应该输出该级别的日志
func shouldLog(level string) bool {
	configLevel, exists := levelPriority[cfg.Level]
	if !exists {
		configLevel = levelPriority["info"]
	}

	currentLevel, exists := levelPriority[level]
	if !exists {
		return false
	}

	return currentLevel >= configLevel
}

// logWithLevel 内部通用日志处理
func logWithLevel(level, msg string, fields ...Field) {
	mu.Lock()
	defer mu.Unlock()

	// 检查日志级别
	if !shouldLog(level) {
		return
	}

	if len(loggers) == 0 {
		return
	}

	// 合并标签
	labels := map[string]string{}
	for k, v := range cfg.Labels {
		labels[k] = v
	}

	// 添加默认标签
	labels["level"] = level

	// 处理额外字段
	extraFields := map[string]interface{}{}
	for _, f := range fields {
		extraFields[f.Key] = f.Value
	}

	// 构建日志条目
	entry := &LogEntry{
		Level:   level,
		Message: msg,
		Labels:  labels,
		Fields:  extraFields,
		Time:    time.Now().Unix(),
	}

	// 分发到所有Writer
	for _, w := range loggers {
		_ = w.Write(entry)
	}
}
