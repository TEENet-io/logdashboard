package log

import (
	"encoding/json"
)

// FormatLogEntry 将LogEntry格式化为JSON字符串
func FormatLogEntry(entry *LogEntry) (string, error) {
	b, err := json.Marshal(entry)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
