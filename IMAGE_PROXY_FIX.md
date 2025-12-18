# Image Proxy Fix for Template Images

## Problem
DashScope API was timing out when trying to download images from `nanobanana.uz`, causing video generation to fail. The system was falling back to a default test image (cat), so all videos were generated with the same image regardless of the template.

## Solution
Created an image proxy endpoint that fetches images from `nanobanana.uz` and serves them through the backend. This allows DashScope to access the template images reliably.

### Changes Made

1. **Image Proxy Endpoint** (`/proxy/image`):
   - Added in `cmd/api/main.go`
   - Fetches images from external URLs (like `nanobanana.uz`)
   - Serves them with proper headers for DashScope
   - Includes timeout handling and error responses

2. **Provider Updates**:
   - Updated `WanAIProvider` to use proxy for `nanobanana.uz` images
   - Added `serverBaseURL` parameter to construct proxy URLs
   - Properly URL-encodes image URLs in proxy requests

3. **Configuration**:
   - Added `API_BASE_URL` to `ServerConfig` (defaults to `https://api.arabella.uz`)
   - Can be configured via `API_BASE_URL` environment variable

## How It Works

1. When a template has a thumbnail from `nanobanana.uz`:
   - Provider constructs proxy URL: `https://api.arabella.uz/proxy/image?url=<encoded-image-url>`
   - DashScope requests the image from the proxy endpoint
   - Backend fetches the image from `nanobanana.uz` and streams it to DashScope

2. For other template images:
   - Used directly if publicly accessible
   - No proxy needed

3. If no template thumbnail:
   - Falls back to default test image

## Deployment

1. **Set API Base URL** (if different from default):
   ```bash
   # In .env file
   API_BASE_URL=https://api.arabella.uz
   ```

2. **Rebuild and restart**:
   ```bash
   cd /var/www/arabella/backend
   go build -o bin/api ./cmd/api
   sudo systemctl restart arabella-api
   ```

## Verification

After restarting, check logs:
```bash
sudo journalctl -u arabella-api -f | grep -i "proxy\|template thumbnail"
```

You should see:
```
Using proxy for nanobanana.uz image {"original_url": "https://nanobanana.uz/...", "proxy_url": "https://api.arabella.uz/proxy/image?url=..."}
```

Or for other templates:
```
Using template thumbnail for image-to-video {"thumbnail_url": "https://..."}
```

## Testing

1. Generate a video with a template that has a `nanobanana.uz` thumbnail
2. Check that the video uses the template image, not the default cat image
3. Verify proxy endpoint works: `curl https://api.arabella.uz/proxy/image?url=<your-image-url>`

## Future Improvements

- Add caching for proxied images
- Add rate limiting for proxy endpoint
- Support for other domains that may timeout
- Automatic retry logic for failed image fetches




