# Image-to-Video Setup

## Changes Made

The Wan AI provider now uses **image-to-video generation** instead of text-to-video for better quality results.

### Model
- **Model**: `wan2.6-i2v` (Image-to-Video)
- **Quality**: Higher quality than text-to-video
- **Features**: Supports audio, multi-shot, and better visual consistency

### Configuration
- **Default Image**: Uses `https://cdn.translate.alibaba.com/r/wanx-demo-1.png` as test image
- **Resolution**: Supports 480P, 720P, 1080P
- **Duration**: Supports 5, 10, or 15 seconds
- **Audio**: Enabled by default
- **Shot Type**: Single shot (can be changed to "multi" for multi-shot)

### API Parameters
- Uses `resolution` parameter (not `size`) for i2v models
- Includes `img_url` in the input
- Supports `audio`, `shot_type`, `prompt_extend` parameters

## Deployment

```bash
cd /var/www/arabella/backend
sudo systemctl restart arabella-api
```

## Testing

After restarting, try generating a video. The system will:
1. Use the test image from DashScope demo
2. Generate video using `wan2.6-i2v` model
3. Include audio in the generated video

## Future Enhancement

To use custom images:
1. Add `ImageURL` field to `VideoParams` struct
2. Pass image URL from template or user input
3. Update the code to use the custom image URL instead of the default

## Verification

Check logs to see image-to-video requests:
```bash
sudo journalctl -u arabella-api -f | grep -i "image-to-video\|wan2.6-i2v"
```

You should see:
```
DashScope (Wan AI) API request - Image-to-Video {"model": "wan2.6-i2v", "image_url": "https://cdn.translate.alibaba.com/r/wanx-demo-1.png", ...}
```





