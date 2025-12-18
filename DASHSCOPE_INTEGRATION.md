# Alibaba Cloud DashScope Integration

## Overview

The Wan AI provider has been updated to use **Alibaba Cloud DashScope** API, which is the underlying infrastructure for Wan AI video generation.

## Changes Made

1. **Updated API Endpoint**: Changed from `https://api.wanai.dev/v1` (non-existent) to `https://dashscope-intl.aliyuncs.com/compatible-mode/v1` (Singapore region)

2. **Updated API Format**: Changed from custom Wan AI format to DashScope's API format:
   - Endpoint: `/api/v1/services/aigc/text2video/generation`
   - Model: `wan2.5-t2v-preview` (or `wan2.6-t2v-preview` if available)
   - Request format: Uses `model`, `input`, `parameters` structure
   - Async mode: Enabled via `X-DashScope-Async: enable` header

3. **Task Status**: Uses DashScope's task status endpoint (`/api/v1/tasks/{task_id}`)

4. **Video URL Retrieval**: Added `GetVideoURL()` method to fetch video URL from completed tasks

## Configuration

The base URL is configurable via environment variable:

```bash
# In .env file:
WANAI_API_KEY=sk-679ac6e4ae314491bcf5169cc4d3a38e
WANAI_VERSION=2.6
WANAI_BASE_URL=https://dashscope-intl.aliyuncs.com/compatible-mode/v1
```

### Region Options

- **Singapore (International)**: `https://dashscope-intl.aliyuncs.com/compatible-mode/v1` (default)
- **Beijing (China)**: `https://dashscope.aliyuncs.com/compatible-mode/v1`

## API Request Format

### Video Generation Request

```json
{
  "model": "wan2.5-t2v-preview",
  "input": {
    "prompt": "A cat running on the grass"
  },
  "parameters": {
    "resolution": "720P",
    "duration": 5,
    "prompt_extend": true,
    "watermark": false,
    "audio": false
  }
}
```

### Response

```json
{
  "request_id": "...",
  "output": {
    "task_id": "task_123",
    "task_status": "PENDING"
  }
}
```

### Task Status Response

```json
{
  "output": {
    "task_id": "task_123",
    "task_status": "SUCCEEDED",
    "video_url": "https://..."
  }
}
```

## Status Mapping

- `SUCCEEDED` → `COMPLETED` (100%)
- `RUNNING` / `PROCESSING` → `PROCESSING` (50%)
- `PENDING` / `QUEUED` → `PENDING` (10%)
- `FAILED` / `ERROR` → `FAILED` (0%)

## Next Steps

1. **Deploy the updated backend**:
   ```bash
   cd /var/www/arabella/backend
   sudo systemctl restart arabella-api
   ```

2. **Verify API key**: Ensure your DashScope API key is correct

3. **Test video generation**: Try generating a video through the frontend

4. **Check logs**: Monitor logs for any API errors
   ```bash
   sudo journalctl -u arabella-api -f
   ```

## Notes

- The API key format (`sk-...`) matches DashScope's format
- Video generation is asynchronous - the system polls for completion
- Video URLs are fetched from DashScope's task response when completed
- Cancellation is not supported by DashScope (tasks complete or fail on their own)






