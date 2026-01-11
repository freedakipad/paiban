#!/bin/bash
# 等待服务就绪脚本

set -e

PAIBAN_URL="${PAIBAN_URL:-http://localhost:7012}"
MAX_RETRIES=30
RETRY_INTERVAL=2

echo "等待 PaiBan 服务就绪..."
echo "URL: $PAIBAN_URL/health"

for i in $(seq 1 $MAX_RETRIES); do
    if curl -sf "$PAIBAN_URL/health" > /dev/null 2>&1; then
        echo "✅ 服务已就绪!"
        exit 0
    fi
    echo "⏳ 等待中... ($i/$MAX_RETRIES)"
    sleep $RETRY_INTERVAL
done

echo "❌ 服务未能在预期时间内就绪"
exit 1

