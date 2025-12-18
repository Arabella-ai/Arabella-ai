#!/bin/bash
# Backend Deployment Script
# Rebuilds and restarts the backend service

set -e

echo "🚀 Deploying Arabella Backend..."
echo ""

# Navigate to backend directory
cd /var/www/arabella/backend

# Build the application
echo "📦 Building backend..."
go build -o bin/api ./cmd/api

if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

# Stop existing service
echo "🛑 Stopping existing service..."
ps aux | grep -E "arabella-api|bin/api" | grep -v grep | awk '{print $2}' | xargs kill -9 2>/dev/null || true
sleep 2

# Copy new binary
echo "📋 Installing new binary..."
rm -f arabella-api
cp bin/api arabella-api
chmod +x arabella-api

# Start service
echo "▶️  Starting backend service..."
nohup ./arabella-api > /tmp/arabella-api.log 2>&1 &
sleep 3

# Verify health
echo "🏥 Checking health..."
HEALTH=$(curl -s http://localhost:8112/health 2>/dev/null)
if echo "$HEALTH" | grep -q "healthy"; then
    echo "✅ Backend deployed successfully!"
    echo "   Health: $HEALTH"
else
    echo "❌ Backend health check failed!"
    echo "   Response: $HEALTH"
    echo "   Check logs: tail -f /tmp/arabella-api.log"
    exit 1
fi

echo ""
echo "📊 Service Status:"
ps aux | grep arabella-api | grep -v grep | head -1

echo ""
echo "✅ Deployment complete!"


