#!/bin/bash
# 一键端到端测试脚本：Loki+Grafana+go-log
# 适用于Linux环境
set -e



# 3. 运行 go-log 测试用例
if [ ! -d pkg/log ]; then
  echo "[ERROR] pkg/log 目录不存在，请确认代码结构！"
  exit 1
fi

echo "[INFO] 运行 go-log 测试用例..."
export PATH=$PATH:/usr/local/go/bin
go test ./pkg/log

# 4. 检查本地日志文件内容
LOGFILE="pkg/log/test.log"
if [ -f "$LOGFILE" ]; then
  echo "[INFO] 本地日志内容如下："
  cat "$LOGFILE"
else
  echo "[WARN] 未找到本地日志文件 $LOGFILE"
fi

# 5. 输出Grafana访问说明
echo "\n[INFO] Grafana 已启动，请访问：http://localhost:3000"
echo "      默认用户名：admin  密码：admin"
echo "      添加 Loki 数据源：http://loki:3100"
echo "      Explore 页面可按标签过滤日志（如 service、env、host 等）" 