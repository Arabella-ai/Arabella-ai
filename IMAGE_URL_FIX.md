# Image URL Fix for DashScope

## Problem
DashScope API returned error: `"Download the media resource timed out during the data inspection process"`

The image URL from `nanobanana.uz` was timing out when DashScope tried to download it.

## Solution
Added validation to skip `nanobanana.uz` images and use a default test image that DashScope can access.

### Changes Made

1. **Image URL Validation**:
   - Checks if image URL is from `nanobanana.uz` (which times out)
   - Skips inaccessible URLs and uses default test image
   - Validates URLs are publicly accessible (not localhost/internal IPs)

2. **Fallback Mechanism**:
   - If template thumbnail is from `nanobanana.uz` → uses default test image
   - If template thumbnail is missing → uses default test image
   - If template thumbnail is accessible → uses it

3. **Default Test Image**:
   - Uses: `https://cdn.translate.alibaba.com/r/wanx-demo-1.png`
   - This is a publicly accessible DashScope demo image

## Current Behavior

- **Template images from `nanobanana.uz`**: Skipped, uses default test image
- **Other template images**: Used if publicly accessible
- **Missing template images**: Uses default test image
- **Prompt**: Uses combined prompt (template base + user input)

## Future Enhancement

To use `nanobanana.uz` images:
1. Make images publicly accessible (no auth required)
2. Ensure images load quickly (< 5 seconds)
3. Or proxy images through your backend/CDN
4. Or upload images to a CDN that DashScope can access

## Deployment

```bash
cd /var/www/arabella/backend
sudo systemctl restart arabella-api
```

## Verification

After restarting, check logs:
```bash
sudo journalctl -u arabella-api -f | grep -i "default test image\|template thumbnail"
```

You should see:
```
Using default test image {"reason": "template thumbnail not accessible or from restricted domain", "original_thumbnail": "https://nanobanana.uz/..."}
```

Or if using a different template:
```
Using template thumbnail for image-to-video {"thumbnail_url": "https://..."}
```




