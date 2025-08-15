#!/bin/bash

# SVG Demo 快速演示脚本
# 使用方法: chmod +x demo_script.sh && ./demo_script.sh

echo "🎨 SVG Demo 项目演示开始..."
echo "================================"

# 服务器地址
BASE_URL="http://localhost:8080"

# 检查服务器状态
echo "📡 检查服务器状态..."
curl -s "$BASE_URL/health" | grep -q "ok" && echo "✅ 服务器运行正常" || echo "❌ 服务器未启动"
echo ""

# 演示1: SVG.IO + 翻译功能
echo "🌍 演示1: SVG.IO Provider (中文翻译)"
echo "提示词: '一只戴帽子的可爱猫咪,卡通风格'"
echo "处理中..."
curl -X POST "$BASE_URL/v1/images/svgio" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "一只戴帽子的可爱猫咪", "style": "cartoon"}' \
  -s -o demo_svgio.svg
echo "✅ SVG.IO 生成完成 → demo_svgio.svg"
echo ""

# 演示2: Recraft 无背景
echo "🎯 演示2: Recraft Provider (中文直接支持 + 无背景)"
echo "提示词: '简约的科技齿轮图标'"
echo "处理中..."
curl -X POST "$BASE_URL/v1/images/recraft/svg" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "简约的科技齿轮图标", "style": "vector_illustration"}' \
  -s -o demo_recraft.svg
echo "✅ Recraft 生成完成 → demo_recraft.svg"
echo ""

# 演示3: Claude AI代码生成
echo "🤖 演示3: Claude Provider (AI直接生成SVG代码)"
echo "提示词: 'geometric mountain sunset landscape'"
echo "处理中..."
curl -X POST "$BASE_URL/v1/images/claude/svg" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "geometric mountain sunset landscape", "style": "modern minimalist", "negative_prompt": "complex details, realistic textures"}' \
  -s -o demo_claude.svg
echo "✅ Claude 生成完成 → demo_claude.svg"
echo ""

# JSON元数据演示
echo "📊 演示4: JSON元数据响应 (Claude)"
echo "获取详细的生成信息..."
curl -X POST "$BASE_URL/v1/images/claude" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "a smiling sun with rays", "style": "cheerful cartoon"}' \
  -s | python3 -m json.tool
echo ""

# 性能测试演示
echo "⚡ 演示5: 并发性能测试"
echo "同时发送5个请求到不同provider..."

# 并发测试
(curl -X POST "$BASE_URL/v1/images/claude" -H "Content-Type: application/json" -d '{"prompt": "star"}' -s > /dev/null && echo "✅ Claude 请求完成") &
(curl -X POST "$BASE_URL/v1/images/claude" -H "Content-Type: application/json" -d '{"prompt": "moon"}' -s > /dev/null && echo "✅ Claude 请求完成") &
(curl -X POST "$BASE_URL/v1/images/claude" -H "Content-Type: application/json" -d '{"prompt": "tree"}' -s > /dev/null && echo "✅ Claude 请求完成") &
(curl -X POST "$BASE_URL/v1/images/claude" -H "Content-Type: application/json" -d '{"prompt": "flower"}' -s > /dev/null && echo "✅ Claude 请求完成") &
(curl -X POST "$BASE_URL/v1/images/claude" -H "Content-Type: application/json" -d '{"prompt": "house"}' -s > /dev/null && echo "✅ Claude 请求完成") &

wait # 等待所有后台任务完成
echo "🎯 并发测试完成!"
echo ""

# 展示生成的文件
echo "📁 本次演示生成的文件:"
ls -la demo_*.svg 2>/dev/null || echo "请检查网络连接和API配置"
echo ""

echo "🎉 演示完成! 主要亮点:"
echo "  ✨ 多Provider智能路由"
echo "  ✨ 中文原生支持 + 自动翻译"
echo "  ✨ 无背景矢量图优化"
echo "  ✨ AI直接生成SVG代码"
echo "  ✨ 高并发处理能力"
echo ""
echo "🔗 查看生成的SVG文件以验证质量差异"
echo "================================"
