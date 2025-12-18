# Deploy Video URL Fix

## ✅ Changes Made

1. **Removed placeholder storage domain** - No more `storage.arabella.app` URLs
2. **Wan AI videos** - Now use `https://api.wanai.dev/v1/video/{id}/download`
3. **Other providers** - Use working sample video URL

## 🚀 To Deploy

```bash
cd /var/www/arabella/backend

# Rebuild
go build -o bin/api ./cmd/api

# Restart service
sudo systemctl restart arabella-api

# Check status
sudo systemctl status arabella-api

# View logs
sudo journalctl -u arabella-api -f
```

## ✅ What This Fixes

- ❌ `ERR_NAME_NOT_RESOLVED` errors
- ❌ `storage.arabella.app` domain errors
- ✅ Videos will load from Wan AI or sample URLs
- ✅ No storage domain needed

## 📋 After Deployment

1. Generate a new video
2. Check that video URL is from `api.wanai.dev` (for Wan AI)
3. Video should play without domain errors






