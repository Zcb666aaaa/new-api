# 算能云异步图片生成功能 - 快速开始

## 功能说明

已成功为 new-api 项目集成算能云异步图片生成功能，支持通过 `/imagegenerator/task` 路径提交和查询图片生成任务。

## 支持的模型

- **Qwen-Image** - 通义千问图像生成（开源）
- **Qwen-Image-Plus** - 通义千问图像生成Plus版
- **Qwen-Image-Edit-2509** - 通义千问图像编辑（图生图）
- **Z-Image-Turbo** - Z-Image高速版
- **Wan2.6-T2I** - 万相2.6文生图

## 快速使用

### 1. 创建任务

```bash
curl -X POST "http://your-domain/imagegenerator/task" \
  -H "Authorization: Bearer sk-xxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "Qwen-Image",
    "input": {
      "prompt": "一只可爱的猫咪在弹钢琴"
    }
  }'
```

返回：
```json
{
  "code": "success",
  "message": "success",
  "data": "task_abc123xyz"
}
```

### 2. 查询任务

```bash
curl -X GET "http://your-domain/imagegenerator/task/task_abc123xyz" \
  -H "Authorization: Bearer sk-xxxx"
```

返回（成功时）：
```json
{
  "code": "success",
  "message": "success",
  "data": {
    "task_id": "task_abc123xyz",
    "status": "success",
    "progress": "100%",
    "url": "https://example.com/image.jpg"
  }
}
```

## 渠道配置

在管理后台创建渠道：
- **渠道类型**: Sophnet
- **Base URL**: `https://www.sophnet.com`
- **API Key**: 你的算能云 API Key
- **模型**: 添加以下模型（注意：必须添加实际的模型名称，不是 sophnet_generate）
  - `Qwen-Image`
  - `Qwen-Image-Plus`
  - `Qwen-Image-Edit-2509`
  - `Z-Image-Turbo`
  - `Wan2.6-T2I`

## 主要参数

### 必填参数
- `model` - 模型名称
- `input.prompt` - 提示词

### 可选参数
- `input.images` - 参考图片（图生图时必填）
- `input.negative_prompt` - 反向提示词
- `parameters.size` - 图片尺寸（如 "1328*1328"）
- `parameters.seed` - 随机种子
- `parameters.prompt_extend` - 是否智能改写提示词（默认 true）
- `parameters.watermark` - 是否添加水印（默认 false）

## 任务状态

- `queued` - 排队中
- `in_progress` - 生成中
- `success` - 成功
- `failure` - 失败

## 测试脚本

### Linux/Mac
```bash
export API_BASE="http://localhost:3000"
export API_KEY="sk-your-key"
bash test_sophnet_image.sh
```

### Windows
```cmd
set API_BASE=http://localhost:3000
set API_KEY=sk-your-key
test_sophnet_image.bat
```

## 完整文档

详细文档请查看：
- `docs/SOPHNET_IMAGE_GENERATOR.md` - API 使用文档
- `docs/SOPHNET_IMAGE_INTEGRATION.md` - 集成说明

## 注意事项

1. **图编辑模型** `Qwen-Image-Edit-2509` 必须提供 `input.images` 参数
2. **轮询建议**: 每 2-5 秒查询一次任务状态
3. **图片保存**: 生成的图片 URL 有时效性，请及时下载

## 兼容路径

除了新路径外，仍支持旧路径：
- `/sophnet/submit/generate` - 提交任务
- `/sophnet/fetch/{taskId}` - 查询任务

推荐使用新路径 `/imagegenerator/task`，更符合算能云官方规范。
