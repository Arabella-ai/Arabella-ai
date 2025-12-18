# DashScope Free Models Configuration

## Issue
Error: `AccessDenied.Unpurchased` - The `wan2.5-t2v-preview` model requires a subscription/purchase.

## Solution
Updated to use **free models** that don't require purchase:
- `wanx_2_1_t2v_turbo` - Fast, free text-to-video model
- `wanx_2_1_t2v_plus` - Higher quality, free text-to-video model

## Free Models Available

### 1. wanx_2_1_t2v_turbo (Currently Used)
- **Status**: Free
- **Speed**: Fast generation
- **Quality**: Good
- **Use Case**: General video generation

### 2. wanx_2_1_t2v_plus
- **Status**: Free
- **Speed**: Slower than turbo
- **Quality**: Higher quality
- **Use Case**: When quality is more important than speed

### 3. wan2.5-t2v-preview (Premium)
- **Status**: Requires subscription/purchase
- **Free Quota**: 50 images for new users (Singapore region only)
- **Cost**: $0.03 per image after free quota
- **Use Case**: Latest features, best quality

## How to Get Free Quota for wan2.5

1. **Register/Login** to [Alibaba Cloud Model Studio](https://dashscope.console.aliyun.com/)
2. **Activate Model Studio** in Singapore region
3. **Free Quota**: 50 images valid for 90 days
4. **After quota**: Pay-as-you-go at $0.03 per image

## Switching Models

To use a different model, update `wanai_provider.go`:

```go
// For free fast model (current)
modelName := "wanx_2_1_t2v_turbo"

// For free high-quality model
modelName := "wanx_2_1_t2v_plus"

// For premium model (requires subscription)
modelName := "wan2.5-t2v-preview"
```

## Deployment

```bash
cd /var/www/arabella/backend
go build -o bin/api ./cmd/api
sudo systemctl restart arabella-api
```

## Verification

Check logs to confirm model is being used:
```bash
sudo journalctl -u arabella-api -f | grep -i "model"
```

You should see:
```
DashScope (Wan AI) API request {"model": "wanx_2_1_t2v_turbo", ...}
```

## Notes

- Free models may have different parameter requirements
- Video quality and duration may vary between models
- If you get access to wan2.5, you can switch back by changing the model name






