package main

import (
	"errors"
	"time"

	"github.com/TEENet-io/logdashboard/pkg/log"
)

func main() {
	// 初始化日志系统
	log.Init(log.Config{
		Level:    "debug", // 支持debug, info, warn, error
		FilePath: "./app.log",
		LokiURL:  "http://localhost:3100/loki/api/v1/push",
		Labels: map[string]string{
			"service":    "example-app",
			"env":        "development",
			"version":    "1.0.0",
			"component":  "main",
			"datacenter": "local",
		},
	})

	// 示例1：基本日志记录
	log.Info("应用程序启动", log.NewField("startup_time", time.Now().Format(time.RFC3339)))

	// 示例2：Debug级别日志
	log.Debug("调试信息", log.NewField("debug_flag", true))

	// 示例3：带多个字段的日志
	log.Info("用户登录",
		log.NewField("user_id", "12345"),
		log.NewField("username", "alice"),
		log.NewField("ip", "192.168.1.100"),
		log.NewField("user_agent", "Mozilla/5.0"),
	)

	// 示例4：警告日志
	log.Warn("数据库连接缓慢",
		log.NewField("connection_time", 5.2),
		log.NewField("threshold", 3.0),
		log.NewField("database", "postgres"),
	)

	// 示例5：错误日志
	err := errors.New("数据库连接失败")
	log.Error("数据库错误",
		log.NewField("error", err.Error()),
		log.NewField("database", "mysql"),
		log.NewField("retry_count", 3),
	)

	// 示例6：业务逻辑日志
	simulateBusinessLogic()

	log.Info("应用程序结束")
}

// simulateBusinessLogic 模拟业务逻辑产生的日志
func simulateBusinessLogic() {
	log.Info("开始处理订单")

	// 模拟订单处理
	for i := 1; i <= 5; i++ {
		log.Info("处理订单",
			log.NewField("order_id", i),
			log.NewField("customer_id", 100+i),
			log.NewField("amount", 99.99*float64(i)),
			log.NewField("status", "processing"),
		)

		// 模拟处理时间
		time.Sleep(100 * time.Millisecond)

		if i == 3 {
			log.Warn("订单处理缓慢",
				log.NewField("order_id", i),
				log.NewField("processing_time", 2.5),
			)
		}
	}

	log.Info("订单处理完成", log.NewField("total_orders", 5))
}
