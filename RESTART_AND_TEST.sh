#!/bin/bash
echo "=== Restarting backend ==="
sudo systemctl stop arabella-api
cp /var/www/arabella/backend/arabella-api /var/www/arabella/backend/bin/api
sudo systemctl start arabella-api
sleep 3
echo "=== Testing image proxy ==="
curl -I "http://localhost:8112/api/v1/proxy/image?url=http%3A%2F%2Fnanobanana.uz%2Fapi%2Fuploads%2Fimages%2Fc85e3ffd-cffe-4483-b78a-2de212908e94.png" 2>&1 | head -8

