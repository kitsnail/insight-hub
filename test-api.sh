#!/bin/bash
# Insight Hub API 测试脚本

BASE_URL="http://127.0.0.1:8090"

echo "=== Insight Hub API 测试 ==="
echo ""

echo "1. 健康检查..."
curl -s "$BASE_URL/api/v1/health" | jq . 2>/dev/null || curl -s "$BASE_URL/api/v1/health"
echo ""

echo "2. 获取统计..."
curl -s "$BASE_URL/api/v1/stats" | jq . 2>/dev/null || curl -s "$BASE_URL/api/v1/stats"
echo ""

echo "3. 创建测试 news item..."
RESULT=$(curl -s -X POST "$BASE_URL/api/v1/items" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "news",
    "source_system": "picoclaw",
    "source_agent": "news-agent",
    "title": "Test News from picoclaw",
    "content": "This is a test news item for integration testing",
    "metadata": {
      "importance": 0.8,
      "category": "tech",
      "keywords": ["kubernetes", "cloud-native"]
    }
  }')
echo "$RESULT" | jq . 2>/dev/null || echo "$RESULT"
echo ""

echo "4. 查询 news items..."
curl -s "$BASE_URL/api/v1/items?type=news&limit=10" | jq . 2>/dev/null || curl -s "$BASE_URL/api/v1/items?type=news&limit=10"
echo ""

echo "5. 按来源统计..."
curl -s "$BASE_URL/api/v1/stats/by-source" | jq . 2>/dev/null || curl -s "$BASE_URL/api/v1/stats/by-source"
echo ""

echo "=== 测试完成 ==="
