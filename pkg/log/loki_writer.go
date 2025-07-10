package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// LokiWriter Loki日志写入器，实现Writer接口
// 负责将日志推送到Loki系统

type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][2]string       `json:"values"`
}

type lokiPayload struct {
	Streams []lokiStream `json:"streams"`
}

type LokiWriter struct {
	lokiURL    string
	labels     map[string]string
	httpClient *http.Client
}

// NewLokiWriter 创建Loki写入器
func NewLokiWriter(lokiURL string, labels map[string]string) *LokiWriter {
	return &LokiWriter{
		lokiURL: lokiURL,
		labels:  labels,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (lw *LokiWriter) Write(entry *LogEntry) error {
	// 使用纳秒级时间戳，Loki要求纳秒级精度
	ts := strconv.FormatInt(time.Unix(entry.Time, 0).UnixNano(), 10)

	// 组装日志内容
	line, err := FormatLogEntry(entry)
	if err != nil {
		return fmt.Errorf("failed to format log entry: %w", err)
	}

	// Loki的values字段是[[timestamp, line]]
	values := [][2]string{{ts, line}}

	// 合并标签
	labels := map[string]string{}
	for k, v := range lw.labels {
		labels[k] = v
	}
	for k, v := range entry.Labels {
		labels[k] = v
	}

	payload := lokiPayload{
		Streams: []lokiStream{{
			Stream: labels,
			Values: values,
		}},
	}

	// 重试推送
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		err := lw.pushToLoki(payload)
		if err == nil {
			return nil
		}

		// 如果不是最后一次重试，等待一下再重试
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	return fmt.Errorf("failed to push to Loki after %d retries: %w", maxRetries, err)
}

// pushToLoki 推送日志到Loki
func (lw *LokiWriter) pushToLoki(payload lokiPayload) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := lw.httpClient.Post(lw.lokiURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to post to Loki: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("Loki returned status code %d", resp.StatusCode)
	}

	return nil
}
