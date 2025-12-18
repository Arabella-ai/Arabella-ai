# DashScope Access Setup Guide

## Current Issue
Error: `AccessDenied.Unpurchased` - Your API key doesn't have access to DashScope video models.

## Solution: Activate Free Quota

DashScope offers **200 seconds of free video generation** for new users (valid for 90 days) for the `wan2.1-t2v-plus` model.

### Steps to Activate Free Access:

1. **Login to Alibaba Cloud**
   - Visit: https://www.alibabacloud.com/
   - Sign up or log in to your account

2. **Access Model Studio**
   - Go to: https://www.alibabacloud.com/product/model-studio
   - **IMPORTANT**: Select the **Singapore** region (free quota is only available there)

3. **Activate Model Studio**
   - Look for "Activate Now" button at the top
   - Click to activate (it's free, no charges until quota is used)
   - If you don't see the button, it's already activated

4. **Verify Your API Key**
   - Go to: https://www.alibabacloud.com/help/doc-detail/2840914.html
   - Check that your API key has access to `wan2.1-t2v-plus`
   - The free quota should show: **200 seconds** (valid for 90 days)

5. **Update Your API Key (if needed)**
   - If you created a new API key, update it in your `.env` file:
   ```bash
   WANAI_API_KEY=your_new_api_key_here
   ```

6. **Restart Backend**
   ```bash
   cd /var/www/arabella/backend
   sudo systemctl restart arabella-api
   ```

## Alternative: Use Mock Provider

If you can't get DashScope access, you can use the mock provider for testing:

1. **Enable Mock Provider** in `.env`:
   ```bash
   USE_MOCK_PROVIDER=true
   ```

2. **Restart Backend**:
   ```bash
   sudo systemctl restart arabella-api
   ```

The mock provider will generate sample videos for testing purposes.

## Verification

After activating free quota, test video generation. Check logs:
```bash
sudo journalctl -u arabella-api -f | grep -i "dashscope\|error"
```

You should see successful requests instead of `AccessDenied.Unpurchased` errors.

## Free Quota Details

- **Model**: `wan2.1-t2v-plus`
- **Free Quota**: 200 seconds of video generation
- **Validity**: 90 days from activation
- **Region**: Singapore only
- **After Quota**: Pay-as-you-go pricing

## Troubleshooting

If you still get `AccessDenied.Unpurchased` after activation:

1. **Check Region**: Make sure you activated Model Studio in **Singapore** region
2. **Check API Key**: Verify the API key is from the Singapore region
3. **Wait a few minutes**: Activation might take a few minutes to propagate
4. **Contact Support**: If issues persist, contact Alibaba Cloud support






