#!/bin/bash
set -e

BASE_URL="http://localhost:8090"

echo "=== Starting Insight Hub ==="
cd /Users/wh/data/worklib/src/github.com/kitsnail/insight-hub
./bin/insight-hub -config config.postgres.yaml 2>&1 &
PID=$!
sleep 2

echo ""
echo "=== Health Check ==="
curl -s "$BASE_URL/api/v1/health" | jq .

echo ""
echo "=== Test 1: Single Create with metadata ==="
curl -s -X POST "$BASE_URL/api/v1/items" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "news",
    "source_system": "picoclaw",
    "title": "Test with metadata",
    "metadata": {
      "external_id": "test-001",
      "importance": 0.9
    }
  }' | jq .

echo ""
echo "=== Test 2: Batch Create without metadata ==="
curl -s -X POST "$BASE_URL/api/v1/items/batch" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {
        "type": "news",
        "source_system": "picoclaw",
        "title": "Test News 1"
      }
    ]
  }' | jq .

echo ""
echo "=== Test 3: Batch Create with metadata ==="
curl -s -X POST "$BASE_URL/api/v1/items/batch" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {
        "type": "news",
        "source_system": "picoclaw",
        "title": "Test News 2",
        "metadata": {
          "external_id": "test-002"
        }
      }
    ]
  }' | jq .

echo ""
echo "=== Test 4: Check Duplicate ==="
curl -s -X POST "$BASE_URL/api/v1/items/check-duplicate" \
  -H "Content-Type: application/json" \
  -d '{
    "source_system": "picoclaw",
    "external_id": "test-001"
  }' | jq .

echo ""
echo "=== Checking service status ==="
if ps -p $PID > /dev/null 2>&1; then
  echo "Service is running (PID: $PID)"
  kill $PID 2>/dev/null
else
  echo "Service stopped unexpectedly"
fi
