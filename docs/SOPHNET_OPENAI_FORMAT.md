# 算能云 OpenAI 格式兼容说明

## 概述

算能云渠道现已支持 OpenAI 官方格式的图片和视频生成请求，同时保留原有的算能原生格式接口。

## 图片生成

### OpenAI 格式请求

使用标准的 OpenAI `/v1/images/generations` 接口：

```bash
curl https://your-new-api-domain/v1/images/generations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "model": "Qwen-Image-Plus",
    "prompt": "一只可爱的猫咪在草地上玩耍",
    "size": "1024x1024",
    "n": 1,
    "watermark": false
  }'
```

### 支持的参数

#### 标准 OpenAI 参数
- `model`: 模型名称（必填）
  - `Qwen-Image`
  - `Qwen-Image-Plus`
  - `Qwen-Image-Edit-2509`
  - `Z-Image-Turbo`
  - `Wan2.6-T2I`
- `prompt`: 图片描述（必填）
- `size`: 图片尺寸，如 `1024x1024`
- `n`: 生成数量（暂不支持，固定为 1）
- `watermark`: 是否添加水印

#### 算能扩展参数（通过 extra 字段）

```json
{
  "model": "Qwen-Image-Plus",
  "prompt": "一只可爱的猫咪",
  "size": "1024x1024",
  "negative_prompt": "模糊，低质量",
  "seed": 12345,
  "prompt_extend": true,
  "save_to_jpeg": true
}
```

扩展参数说明：
- `negative_prompt`: 负面提示词
- `seed`: 随机种子
- `prompt_extend`: 是否扩展提示词
- `save_to_jpeg`: 是否保存为 JPEG 格式

### 图片编辑（Qwen-Image-Edit-2509）

```json
{
  "model": "Qwen-Image-Edit-2509",
  "prompt": "将猫咪的颜色改为橙色",
  "images": [
    "https://example.com/cat.jpg"
  ]
}
```

### 算能原生格式（仍然支持）

```bash
curl https://your-new-api-domain/imagegenerator/task \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "model": "Qwen-Image-Plus",
    "input": {
      "prompt": "一只可爱的猫咪在草地上玩耍",
      "negative_prompt": "模糊，低质量"
    },
    "parameters": {
      "size": "1024x1024",
      "seed": 12345,
      "watermark": false
    }
  }'
```

## 视频生成

### OpenAI 格式请求

使用标准的 OpenAI `/v1/videos` 或 `/v1/video/generations` 接口：

```bash
curl https://your-new-api-domain/v1/videos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "model": "Wan2.6-T2V",
    "prompt": "一只猫咪在草地上奔跑",
    "duration": 5.0,
    "width": 1280,
    "height": 720,
    "seed": 12345
  }'
```

### 支持的参数

#### 标准参数
- `model`: 模型名称（必填）
  - T2V 模型：`Wan2.6-T2V`, `Seedance-1.5-Pro`, `ViduQ2` 等
  - I2V 模型：`Wan2.6-I2V`, `Seedance-1.0-Lite-I2V` 等
- `prompt`: 视频描述（必填，T2V 模型）
- `image`: 输入图片 URL 或 base64（必填，I2V 模型）
- `duration`: 视频时长（秒）
- `width`: 视频宽度
- `height`: 视频高度
- `seed`: 随机种子

#### 算能扩展参数（通过 metadata 字段）

```json
{
  "model": "Wan2.6-T2V",
  "prompt": "一只猫咪在草地上奔跑",
  "duration": 5.0,
  "width": 1280,
  "height": 720,
  "metadata": {
    "negative_prompt": "模糊，抖动",
    "subdivision_level": "high",
    "file_format": "mp4",
    "callback_url": "https://your-domain.com/callback",
    "return_last_frame": true,
    "service_tier": "premium",
    "generate_audio": false
  }
}
```

扩展参数说明：
- `negative_prompt`: 负面提示词
- `subdivision_level`: 细分级别
- `file_format`: 文件格式
- `callback_url`: 回调 URL
- `return_last_frame`: 是否返回最后一帧
- `service_tier`: 服务等级
- `generate_audio`: 是否生成音频

### 图生视频（I2V）

```json
{
  "model": "Wan2.6-I2V",
  "prompt": "猫咪开始奔跑",
  "image": "https://example.com/cat.jpg",
  "duration": 5.0
}
```

或使用 base64：

```json
{
  "model": "Wan2.6-I2V",
  "prompt": "猫咪开始奔跑",
  "image": "data:image/jpeg;base64,/9j/4AAQSkZJRg...",
  "duration": 5.0
}
```

### 算能原生格式（仍然支持）

```bash
curl https://your-new-api-domain/videogenerator/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "model": "Wan2.6-T2V",
    "content": [
      {
        "type": "text",
        "text": "一只猫咪在草地上奔跑",
        "negative_prompt": "模糊，抖动"
      }
    ],
    "parameters": {
      "size": "1280x720",
      "duration": 5,
      "seed": "12345"
    }
  }'
```

## 响应格式

### 图片生成响应

OpenAI 格式响应：

```json
{
  "code": 0,
  "message": "success",
  "data": "task_abc123"
}
```

### 视频生成响应

OpenAI 格式响应：

```json
{
  "code": 0,
  "message": "success",
  "data": "task_xyz789"
}
```

## 任务查询

使用统一的任务查询接口：

```bash
# 查询图片任务
curl https://your-new-api-domain/imagegenerator/task/task_abc123 \
  -H "Authorization: Bearer YOUR_TOKEN"

# 查询视频任务
curl https://your-new-api-domain/videogenerator/generate/task_xyz789 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

或使用 OpenAI 格式：

```bash
# 查询视频任务
curl https://your-new-api-domain/v1/videos/task_xyz789 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 支持的模型列表

### 图片生成模型
- `Qwen-Image`: 通义千问图片生成
- `Qwen-Image-Plus`: 通义千问图片生成增强版
- `Qwen-Image-Edit-2509`: 通义千问图片编辑
- `Z-Image-Turbo`: Z-Image 快速版
- `Wan2.6-T2I`: 万相文生图

### 视频生成模型

#### 万相系列
- `Wan2.2-T2V-Plus`: 万相 2.2 文生视频增强版
- `Wan2.2-I2V-Plus`: 万相 2.2 图生视频增强版
- `Wan2.5-T2V-Preview`: 万相 2.5 文生视频预览版
- `Wan2.5-I2V-Preview`: 万相 2.5 图生视频预览版
- `Wan2.6-T2V`: 万相 2.6 文生视频
- `Wan2.6-I2V`: 万相 2.6 图生视频
- `Wan2.2-T2V-A14B`: 万相 2.2 文生视频 A14B
- `Wan2.2-I2V-A14B`: 万相 2.2 图生视频 A14B

#### 字节跳动系列
- `Seedance-1.5-Pro`: Seedance 1.5 专业版
- `Seedance-1.0-Pro`: Seedance 1.0 专业版
- `Seedance-1.0-Pro-Fast`: Seedance 1.0 专业快速版
- `Seedance-1.0-Lite-T2V`: Seedance 1.0 轻量文生视频
- `Seedance-1.0-Lite-I2V`: Seedance 1.0 轻量图生视频
- `Doubao-Seed3D`: 豆包 Seed3D

#### 生数系列
- `ViduQ2`: 生数 Q2
- `ViduQ2-turbo`: 生数 Q2 快速版
- `ViduQ2-pro`: 生数 Q2 专业版
- `ViduQ2-pro-fast`: 生数 Q2 专业快速版
- `ViduQ1`: 生数 Q1
- `ViduQ1-classic`: 生数 Q1 经典版
- `Vidu2.0`: 生数 2.0
- `Vidu1.5`: 生数 1.5

## 注意事项

1. **自动识别**：系统会根据模型名称自动识别是否为算能模型，并路由到相应的处理器
2. **格式兼容**：OpenAI 格式和算能原生格式可以同时使用，互不影响
3. **异步处理**：图片和视频生成都是异步任务，需要通过任务 ID 查询结果
4. **参数映射**：OpenAI 格式的参数会自动转换为算能格式
5. **扩展参数**：可以通过额外字段传递算能特有的参数

## 示例代码

### Python (OpenAI SDK)

```python
from openai import OpenAI

client = OpenAI(
    api_key="YOUR_TOKEN",
    base_url="https://your-new-api-domain/v1"
)

# 图片生成
response = client.images.generate(
    model="Qwen-Image-Plus",
    prompt="一只可爱的猫咪在草地上玩耍",
    size="1024x1024",
    n=1
)

print(response)

# 视频生成（需要自定义请求）
import requests

response = requests.post(
    "https://your-new-api-domain/v1/videos",
    headers={
        "Authorization": f"Bearer YOUR_TOKEN",
        "Content-Type": "application/json"
    },
    json={
        "model": "Wan2.6-T2V",
        "prompt": "一只猫咪在草地上奔跑",
        "duration": 5.0,
        "width": 1280,
        "height": 720
    }
)

print(response.json())
```

### Node.js

```javascript
const OpenAI = require('openai');

const client = new OpenAI({
  apiKey: 'YOUR_TOKEN',
  baseURL: 'https://your-new-api-domain/v1'
});

// 图片生成
async function generateImage() {
  const response = await client.images.generate({
    model: 'Qwen-Image-Plus',
    prompt: '一只可爱的猫咪在草地上玩耍',
    size: '1024x1024',
    n: 1
  });
  
  console.log(response);
}

// 视频生成
async function generateVideo() {
  const response = await fetch('https://your-new-api-domain/v1/videos', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer YOUR_TOKEN',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      model: 'Wan2.6-T2V',
      prompt: '一只猫咪在草地上奔跑',
      duration: 5.0,
      width: 1280,
      height: 720
    })
  });
  
  console.log(await response.json());
}
```

## 相关文档

- [算能云文本生成](./SOPHNET_CHAT.md)
- [算能云图片生成（原生格式）](./SOPHNET_IMAGE_GENERATOR.md)
- [算能云视频生成（原生格式）](./SOPHNET_VIDEO_GENERATOR.md)
