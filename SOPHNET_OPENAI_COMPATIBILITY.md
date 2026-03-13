# 算能渠道 OpenAI 格式兼容实现总结

## 实现概述

为算能云渠道添加了 OpenAI 官方格式的图片和视频生成请求支持，同时保留了原有的算能原生格式接口。

## 修改的文件

### 1. 新增文件

#### `relay/channel/task/sophnet/openai_converter.go`
- 实现了 OpenAI 格式到算能格式的转换
- `ConvertOpenAIImageRequest`: 转换图片生成请求
- `ConvertOpenAIVideoRequest`: 转换视频生成请求
- `unmarshalRawMessage`: 辅助函数，用于解析 JSON 扩展参数

#### `docs/SOPHNET_OPENAI_FORMAT.md`
- 详细的使用文档
- 包含图片和视频生成的示例
- 参数说明和模型列表
- Python 和 Node.js 示例代码

#### `test_sophnet_openai_format.sh` 和 `test_sophnet_openai_format.bat`
- 测试脚本，用于验证功能

### 2. 修改的文件

#### `relay/channel/task/sophnet/adaptor.go`
- 修改 `ValidateRequestAndSetAction` 方法，增加对 OpenAI 格式请求的识别
- 新增 `validateOpenAIImageRequest` 方法，验证 OpenAI 格式的图片请求
- 新增 `validateOpenAIVideoRequest` 方法，验证 OpenAI 格式的视频请求
- 在验证时自动调用转换函数，将 OpenAI 格式转换为算能格式

#### `middleware/distributor.go`
- 修改 `/v1/images/generations` 路径处理，检查是否为算能图片模型
- 修改 `/v1/videos` 和 `/v1/video/generations` 路径处理，检查是否为算能视频模型
- 新增 `isSophnetImageModel` 函数，判断是否为算能图片模型
- 新增 `isSophnetVideoModel` 函数，判断是否为算能视频模型
- 当检测到算能模型时，设置 `channel_type` 为 `ChannelTypeSophnet`

#### `router/relay-router.go`
- 修改 `/v1/images/generations` 路由，检查 `channel_type`
- 如果是算能渠道，路由到 `RelayTask` 而不是 `Relay`

## 工作流程

### 图片生成流程

1. 客户端发送 OpenAI 格式的图片生成请求到 `/v1/images/generations`
2. `middleware.Distribute()` 解析请求，提取模型名称
3. `isSophnetImageModel()` 判断是否为算能图片模型
4. 如果是，设置 `channel_type` 为 `ChannelTypeSophnet`
5. 路由层检查 `channel_type`，将请求路由到 `RelayTask`
6. `TaskAdaptor.ValidateRequestAndSetAction()` 识别为 OpenAI 格式
7. 调用 `ConvertOpenAIImageRequest()` 转换为算能格式
8. 将转换后的请求存储到 context 中
9. 后续处理使用算能格式的请求

### 视频生成流程

1. 客户端发送 OpenAI 格式的视频生成请求到 `/v1/videos`
2. `middleware.Distribute()` 解析请求，提取模型名称
3. `isSophnetVideoModel()` 判断是否为算能视频模型
4. 如果是，设置 `channel_type` 为 `ChannelTypeSophnet`
5. 路由层已经将 `/v1/videos` 路由到 `RelayTask`
6. `TaskAdaptor.ValidateRequestAndSetAction()` 识别为 OpenAI 格式
7. 调用 `ConvertOpenAIVideoRequest()` 转换为算能格式
8. 将转换后的请求存储到 context 中
9. 后续处理使用算能格式的请求

## 参数映射

### 图片生成参数映射

| OpenAI 参数 | 算能参数 | 说明 |
|------------|---------|------|
| `model` | `model` | 模型名称 |
| `prompt` | `input.prompt` | 图片描述 |
| `size` | `parameters.size` | 图片尺寸 |
| `watermark` | `parameters.watermark` | 是否添加水印 |
| `negative_prompt` (扩展) | `input.negative_prompt` | 负面提示词 |
| `seed` (扩展) | `parameters.seed` | 随机种子 |
| `prompt_extend` (扩展) | `parameters.prompt_extend` | 是否扩展提示词 |
| `save_to_jpeg` (扩展) | `parameters.save_to_jpeg` | 是否保存为 JPEG |
| `images` (扩展) | `input.images` | 输入图片（用于编辑） |

### 视频生成参数映射

| OpenAI 参数 | 算能参数 | 说明 |
|------------|---------|------|
| `model` | `model` | 模型名称 |
| `prompt` | `content[0].text` | 视频描述 |
| `image` | `content[1].image.url` | 输入图片 |
| `duration` | `parameters.duration` | 视频时长 |
| `width` + `height` | `parameters.size` | 视频尺寸 |
| `seed` | `parameters.seed` | 随机种子 |
| `metadata.negative_prompt` | `content[0].negative_prompt` | 负面提示词 |
| `metadata.subdivision_level` | `parameters.subdivision_level` | 细分级别 |
| `metadata.file_format` | `parameters.file_format` | 文件格式 |
| `metadata.callback_url` | `callback_url` | 回调 URL |
| `metadata.return_last_frame` | `return_last_frame` | 是否返回最后一帧 |
| `metadata.service_tier` | `service_tier` | 服务等级 |
| `metadata.generate_audio` | `generate_audio` | 是否生成音频 |

## 兼容性

### 保留的功能
- 算能原生格式接口完全保留，不受影响
- `/imagegenerator/task` 和 `/videogenerator/generate` 路径继续工作
- 现有客户端代码无需修改

### 新增的功能
- 支持 OpenAI 标准的 `/v1/images/generations` 接口
- 支持 OpenAI 标准的 `/v1/videos` 接口
- 自动识别算能模型并路由到正确的处理器
- 参数自动转换，无需手动适配

## 支持的模型

### 图片生成模型
- Qwen-Image
- Qwen-Image-Plus
- Qwen-Image-Edit-2509
- Z-Image-Turbo
- Wan2.6-T2I

### 视频生成模型
- 万相系列：Wan2.2-T2V-Plus, Wan2.2-I2V-Plus, Wan2.5-T2V-Preview, Wan2.5-I2V-Preview, Wan2.6-T2V, Wan2.6-I2V, Wan2.2-T2V-A14B, Wan2.2-I2V-A14B
- 字节跳动系列：Seedance-1.5-Pro, Seedance-1.0-Pro, Seedance-1.0-Pro-Fast, Seedance-1.0-Lite-T2V, Seedance-1.0-Lite-I2V, Doubao-Seed3D
- 生数系列：ViduQ2, ViduQ2-turbo, ViduQ2-pro, ViduQ2-pro-fast, ViduQ1, ViduQ1-classic, Vidu2.0, Vidu1.5

## 测试方法

### 使用测试脚本

Linux/Mac:
```bash
chmod +x test_sophnet_openai_format.sh
./test_sophnet_openai_format.sh
```

Windows:
```cmd
test_sophnet_openai_format.bat
```

### 手动测试

#### 图片生成
```bash
curl -X POST "http://localhost:3000/v1/images/generations" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "model": "Qwen-Image-Plus",
    "prompt": "一只可爱的猫咪在草地上玩耍",
    "size": "1024x1024"
  }'
```

#### 视频生成
```bash
curl -X POST "http://localhost:3000/v1/videos" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "model": "Wan2.6-T2V",
    "prompt": "一只猫咪在草地上奔跑",
    "duration": 5.0,
    "width": 1280,
    "height": 720
  }'
```

## 注意事项

1. **模型识别**：系统通过模型名称自动识别是否为算能模型
2. **异步处理**：图片和视频生成都是异步任务，返回任务 ID
3. **参数扩展**：可以在 OpenAI 格式中添加算能特有的参数
4. **向后兼容**：原有的算能原生格式接口完全保留
5. **错误处理**：转换失败会返回明确的错误信息

## 未来改进

1. 支持更多 OpenAI 图片参数（如 `quality`, `style` 等）
2. 支持批量生成（`n` 参数）
3. 添加更详细的错误信息和日志
4. 优化参数验证逻辑
5. 添加单元测试

## 相关文档

- [算能云 OpenAI 格式使用文档](./docs/SOPHNET_OPENAI_FORMAT.md)
- [算能云文本生成](./SOPHNET_CHAT.md)
- [算能云图片生成（原生格式）](./docs/SOPHNET_IMAGE_GENERATOR.md)
- [算能云视频生成（原生格式）](./docs/SOPHNET_VIDEO_GENERATOR.md)
