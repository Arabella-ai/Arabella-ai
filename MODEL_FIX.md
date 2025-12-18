# Model Name Fix

## Issue
DashScope API returned error: `"Model not exist."` for `wan2.6-t2v-preview`

## Solution
Updated to use `wan2.5-t2v-preview` which is the available model in DashScope.

## Changes Made

1. **Provider Code** (`wanai_provider.go`):
   - Changed model name from `wan2.6-t2v-preview` to `wan2.5-t2v-preview`
   - Removed version 2.6 logic (not available yet)

2. **Config** (`config.go`):
   - Changed default version from `2.6` to `2.5`

3. **Provider Constructor**:
   - Updated default version from `2.6` to `2.5`

## Deployment

```bash
cd /var/www/arabella/backend
go build -o bin/api ./cmd/api
sudo systemctl restart arabella-api
```

## Verification

After restart, check logs:
```bash
sudo journalctl -u arabella-api -f | grep -i "model\|dashscope"
```

You should see:
```
DashScope (Wan AI) API request {"model": "wan2.5-t2v-preview", ...}
```

## Note

When `wan2.6-t2v-preview` becomes available in DashScope, update:
- `wanai_provider.go`: Add back version 2.6 logic
- `config.go`: Change default back to `2.6` if desired






