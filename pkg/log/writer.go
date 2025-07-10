package log

// Writer 日志写入器接口
// 实现本地文件、Loki等多种写入方式

type Writer interface {
	Write(entry *LogEntry) error
}

// LogEntry 日志条目结构体
// 包含时间、级别、消息、标签、字段等

type LogEntry struct {
	Level   string                 // 日志级别
	Message string                 // 日志内容
	Labels  map[string]string      // 标签
	Fields  map[string]interface{} // 额外字段
	Time    int64                  // 时间戳（秒）
}
