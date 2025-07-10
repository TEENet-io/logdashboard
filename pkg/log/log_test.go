package log

import (
	"testing"
	"time"
)

func TestGoLogWithLabels(t *testing.T) {
	cfg := Config{
		Level:    "info",
		FilePath: "./test.log",
		LokiURL:  "http://localhost:3100/loki/api/v1/push",
		Labels: map[string]string{
			"service": "testservice",
			"env":     "test",
			"host":    "localhost",
			"feature": "label-demo",
		},
	}
	Init(cfg)

	// 测试不同级别的日志
	Debug("这是一条debug日志（应该被过滤）", NewField("debug_info", "test"))
	Info("这是一条info日志", NewField("user", "alice"), NewField("action", "login"))
	Warn("这是一条warning日志", NewField("user", "bob"), NewField("action", "logout"))
	Error("这是一条error日志", NewField("user", "carol"), NewField("action", "fail"))

	// 等待一下，确保日志被写入
	time.Sleep(100 * time.Millisecond)
}

func TestLogLevel(t *testing.T) {
	// 测试不同的日志级别配置
	testCases := []struct {
		level    string
		expected bool
	}{
		{"debug", true},
		{"info", true},
		{"warn", true},
		{"error", true},
	}

	for _, tc := range testCases {
		cfg := Config{
			Level:    tc.level,
			FilePath: "./test_level.log",
			Labels: map[string]string{
				"test": "level",
			},
		}
		Init(cfg)

		// 测试所有级别的日志
		Debug("debug message")
		Info("info message")
		Warn("warn message")
		Error("error message")

		time.Sleep(50 * time.Millisecond)
	}
}

func TestLogWithoutConfig(t *testing.T) {
	// 测试没有配置的情况
	cfg := Config{}
	Init(cfg)

	// 应该不会崩溃
	Info("test message without config")
}

func TestFieldFunc(t *testing.T) {
	cfg := Config{
		Level:    "info",
		FilePath: "./test_field.log",
		Labels: map[string]string{
			"test": "field",
		},
	}
	Init(cfg)

	// 测试新的Field函数
	Info("测试NewField函数", NewField("key1", "value1"), NewField("key2", 42))

	// 测试向后兼容的FieldFunc
	Info("测试FieldFunc函数", FieldFunc("key1", "value1"), FieldFunc("key2", 42))
}
