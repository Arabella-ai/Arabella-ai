# Wan AI Integration Setup

## ✅ Configuration Complete

### 1. **Provider Created**
   - `internal/infrastructure/provider/wanai_provider.go` - Full Wan AI provider implementation
   - Uses version 2.6 by default
   - API Key: `sk-679ac6e4ae314491bcf5169cc4d3a38e`

### 2. **Configuration Added**
   - Added `WANAI_API_KEY` and `WANAI_VERSION` to config
   - Added to `.env` file
   - Provider registered in `main.go`

### 3. **Provider Priority**
   - Wan AI is now the **preferred provider** (highest priority in selection)
   - Will be selected first if available and healthy

## 🔄 To Deploy:

```bash
cd /var/www/arabella/backend

# Rebuild the backend
go build -o bin/api ./cmd/api

# Restart the service
sudo systemctl restart arabella-api

# Check status
sudo systemctl status arabella-api

# View logs
sudo journalctl -u arabella-api -f
```

## 🧪 Testing:

1. **Test API Connection:**
   ```bash
   bash test-wanai.sh
   ```

2. **Test Video Generation:**
   - Use the frontend to generate a video
   - Or use curl:
   ```bash
   curl -X POST http://localhost:8112/api/v1/videos/generate \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "template_id": "YOUR_TEMPLATE_ID",
       "prompt": "A beautiful sunset over mountains"
     }'
   ```

## 📋 Wan AI API Details:

- **Base URL**: `https://api.wanai.dev/v1`
- **Version**: 2.6
- **Endpoint**: `POST /generate/text-to-video`
- **Status Check**: `GET /video/{video_id}/status`
- **Authentication**: Bearer token in Authorization header

## 🔍 Provider Capabilities:

- Max Duration: 60 seconds
- Max Resolution: 4K
- Supported Ratios: 16:9, 9:16, 1:1
- Estimated Time: 20 seconds per video
- Quality Tier: Premium
- Cost: 0.03 credits per second

## ⚠️ Notes:

- Wan AI will be automatically selected for video generation
- If Wan AI is unavailable, system will fall back to other providers
- Check logs for provider selection and API calls







