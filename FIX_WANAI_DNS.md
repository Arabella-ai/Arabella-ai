# Fix Wan AI DNS Issue

## Problem
The domain `api.wanai.dev` cannot be resolved by DNS:
```
dial tcp: lookup api.wanai.dev on 127.0.0.53:53: no such host
```

## Solutions

### Option 1: Configure DNS (Recommended)
If you know the IP address of the API server, add it to `/etc/hosts`:

```bash
# Find the IP address (if you know it)
# Then add to /etc/hosts:
sudo nano /etc/hosts
# Add line:
# <IP_ADDRESS> api.wanai.dev
```

### Option 2: Use Different Base URL
The base URL is now configurable. Add to `.env`:

```bash
# If the API is at a different URL, set it here
WANAI_BASE_URL=https://<correct-domain>/v1
```

### Option 3: Check DNS Configuration
The server uses systemd-resolved. Check DNS:

```bash
# Check DNS status
resolvectl status

# Test with different DNS server
dig api.wanai.dev @8.8.8.8
```

### Option 4: Verify API Endpoint
Please verify:
1. What is the correct API endpoint URL for Wan AI version 2.6?
2. Is `api.wanai.dev` the correct domain?
3. Do you have documentation with the correct endpoint?

## Current Configuration

- API Key: `sk-679ac6e4ae314491bcf5169cc4d3a38e`
- Version: `2.6`
- Base URL: `https://api.wanai.dev/v1` (configurable via `WANAI_BASE_URL`)

## Next Steps

1. Verify the correct API endpoint URL
2. If different, set `WANAI_BASE_URL` in `.env`
3. Or add DNS entry to `/etc/hosts` if you know the IP
4. Restart backend after changes






