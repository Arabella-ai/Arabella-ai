#!/bin/bash
# Restart backend with new binary

echo "Stopping arabella-api service..."
sudo systemctl stop arabella-api

echo "Copying new binary..."
cp /var/www/arabella/backend/arabella-api /var/www/arabella/backend/bin/api

echo "Starting arabella-api service..."
sudo systemctl start arabella-api

echo "Waiting for service to start..."
sleep 2

echo "Checking service status..."
sudo systemctl status arabella-api --no-pager | head -10

echo ""
echo "Testing proxy endpoint..."
curl -I "http://localhost:8112/api/v1/proxy/image?url=http%3A%2F%2Fnanobanana.uz%2Fapi%2Fuploads%2Fimages%2Fc85e3ffd-cffe-4483-b78a-2de212908e94.png" 2>&1 | head -5

