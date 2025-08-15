#!/bin/bash

# 翻译功能测试脚本

echo "=== SVG 图像生成服务翻译功能测试 ==="
echo

# 测试服务是否运行
echo "1. 检查服务状态..."
if ! curl -s http://localhost:8080/ping > /dev/null; then
    echo "❌ 服务未运行，请先启动服务: go run ."
    exit 1
fi
echo "✅ 服务正在运行"
echo

# 测试中文翻译功能
echo "2. 测试中文提示词翻译 (JSON 响应)..."
echo "请求: 一只可爱的卡通狐狸"
response=$(curl -s -X POST http://localhost:8080/v1/images \
  -H 'Content-Type: application/json' \
  -d '{"prompt":"一只可爱的卡通狐狸","style":"卡通"}')

echo "响应:"
echo "$response" | jq .
echo

# 检查是否包含翻译信息
if echo "$response" | jq -e '.was_translated' > /dev/null; then
    echo "✅ 翻译功能正常工作"
    original=$(echo "$response" | jq -r '.original_prompt')
    translated=$(echo "$response" | jq -r '.translated_prompt')
    echo "原文: $original"
    echo "译文: $translated"
else
    echo "⚠️  翻译功能可能未启用或失败"
fi
echo

# 测试英文输入（应该跳过翻译）
echo "3. 测试英文提示词（应该跳过翻译）..."
echo "请求: A cute cartoon fox"
response=$(curl -s -X POST http://localhost:8080/v1/images \
  -H 'Content-Type: application/json' \
  -d '{"prompt":"A cute cartoon fox","style":"cartoon"}')

was_translated=$(echo "$response" | jq -r '.was_translated')
if [ "$was_translated" = "false" ]; then
    echo "✅ 英文输入正确跳过了翻译"
else
    echo "⚠️  英文输入意外进行了翻译"
fi
echo

# 测试直接 SVG 下载
echo "4. 测试中文提示词直接生成 SVG..."
echo "请求: 简约的猫头鹰图标"
curl -s -X POST http://localhost:8080/v1/images/svg \
  -H 'Content-Type: application/json' \
  -d '{"prompt":"简约的猫头鹰图标","style":"线条风格"}' \
  -D headers.txt \
  -o test_owl.svg

if [ -f "test_owl.svg" ] && [ -s "test_owl.svg" ]; then
    echo "✅ SVG 文件生成成功: test_owl.svg"
    echo "文件大小: $(wc -c < test_owl.svg) 字节"
    
    # 检查响应头中的翻译信息
    if grep -q "X-Was-Translated: true" headers.txt; then
        echo "✅ SVG 响应头包含翻译信息"
        echo "原始提示词: $(grep "X-Original-Prompt:" headers.txt | cut -d' ' -f2-)"
        echo "翻译提示词: $(grep "X-Translated-Prompt:" headers.txt | cut -d' ' -f2-)"
    else
        echo "⚠️  SVG 响应头未包含翻译信息"
    fi
else
    echo "❌ SVG 文件生成失败"
fi

# 清理临时文件
rm -f headers.txt

echo
echo "=== 测试完成 ==="
echo
echo "注意事项:"
echo "- 如果翻译功能未工作，请检查 OPENAI_API_KEY 环境变量是否设置"
echo "- 翻译失败不会中断图像生成流程，会使用原始提示词继续处理"
echo "- 可以使用 skip_translate: true 参数强制跳过翻译"
