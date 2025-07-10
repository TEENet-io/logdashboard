package log

// Config 日志配置结构体
// Level: 日志级别（info/warn/error）
// FilePath: 本地日志文件路径
// LokiURL: Loki推送地址
// Labels: 日志自定义标签
type Config struct {
	Level    string            // 日志级别
	FilePath string            // 本地日志文件路径
	LokiURL  string            // Loki推送地址
	Labels   map[string]string // 自定义标签
}
