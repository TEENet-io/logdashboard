#!/bin/bash
# 快速入门脚本：一键启动 Loki+Grafana+go-log 系统
# 适用于 Linux/macOS 环境

set -e

echo "🚀 go-log 快速入门向导"
echo "=========================="

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ 请先安装 Docker"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ 请先安装 docker-compose"
    exit 1
fi

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    echo "❌ 请先安装 Go 语言环境"
    exit 1
fi

echo "✅ 环境检查通过"

# 1. 启动 Docker 服务
echo "🔧 启动 Loki 和 Grafana 服务..."
docker-compose up -d

# 等待服务启动
echo "⏳ 等待服务启动（30秒）..."
sleep 30

# 2. 检查服务状态
echo "🔍 检查服务状态..."
if curl -s http://localhost:3100/ready > /dev/null; then
    echo "✅ Loki 服务已启动"
else
    echo "❌ Loki 服务启动失败"
    exit 1
fi

if curl -s http://localhost:3000/api/health > /dev/null; then
    echo "✅ Grafana 服务已启动"
else
    echo "❌ Grafana 服务启动失败"
    exit 1
fi

# 3. 运行示例程序
echo "🎯 运行示例程序..."
cd example
go run main.go
cd ..

# 4. 运行测试
echo "🧪 运行测试用例..."
go test ./pkg/log

# 5. 检查日志文件
echo "📝 检查生成的日志文件..."
if [ -f "example/app.log" ]; then
    echo "✅ 示例日志文件已生成:"
    echo "---"
    head -5 example/app.log
    echo "---"
else
    echo "❌ 示例日志文件未生成"
fi

if [ -f "pkg/log/test.log" ]; then
    echo "✅ 测试日志文件已生成:"
    echo "---"
    head -5 pkg/log/test.log
    echo "---"
else
    echo "❌ 测试日志文件未生成"
fi

# 6. 输出访问信息
echo ""
echo "🎉 快速入门完成！"
echo "===================="
echo "📊 Grafana 控制台: http://localhost:3000"
echo "   用户名: admin"
echo "   密码: admin"
echo ""
echo "🔍 Loki API: http://localhost:3100"
echo ""
echo "📖 后续步骤:"
echo "1. 在 Grafana 中添加 Loki 数据源: http://loki:3100"
echo "2. 在 Explore 页面查看日志数据"
echo "3. 使用以下查询语句:"
echo "   {service=\"example-app\"}"
echo "   {service=\"example-app\", level=\"error\"}"
echo "   {service=\"testservice\"}"
echo ""
echo "💡 提示: 查看 example/main.go 了解完整使用方法"
echo ""
echo "🛑 停止服务: docker-compose down" 