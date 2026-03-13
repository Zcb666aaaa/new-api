# 算能云异步图片生成功能集成说明

## 更改概述

本次更新为 new-api 项目添加了算能云（Sophnet）异步图片生成功能的完整支持，允许用户通过 `/imagegenerator/task` 路径提交和查询图片生成任务。

## 主要更改

### 1. 路由配置 (`router/relay-router.go`)

添加了新的路由组 `/imagegenerator`，支持：
- `POST /imagegenerator/task` - 创建图片生成任务
- `GET /imagegenerator/task/:id` - 查询任务状态

### 2. 中间件配置 (`middleware/distributor.go`)

添加了对 `/imagegenerator/task` 路径的识别和处理逻辑：
- 自动识别请求类型（创建任务 vs 查询任务）
- 设置正确的平台标识和中继模式
- 自动设置默认模型名称

### 3. 适配器更新 (`relay/channel/task/sophnet/adaptor.go`)

更新了 `ValidateRequestAndSetAction` 方法：
- 保存上游模型名称到 `info.UpstreamModelName`
- 支持从请求体中提取实际的模型名称

### 4. 模型列表更新 (`relay/channel/task/sophnet/models.go`)

添加了所有支持的图片生成模型：
- `sophnet_generate` - 默认模型
- `Qwen-Image` - 通义千问图像生成
- `Qwen-Image-Plus` - 通义千问图像生成Plus版
- `Qwen-Image-Edit-2509` - 通义千问图像编辑
- `Z-Image-Turbo` - Z-Image高速版
- `Wan2.6-T2I` - 万相2.6文生图

### 5. 渠道测试更新 (`controller/channel-test.go`)

将 `ChannelTypeSophnet` 添加到不支持测试的渠道类型列表中，因为异步任务需要特殊的测试流程。

### 6. 文档

创建了以下文档：
- `docs/SOPHNET_IMAGE_GENERATOR.md` - 完整的 API 使用文档
- `test_sophnet_image.sh` - Linux/Mac 测试脚本
- `test_sophnet_image.bat` - Windows 测试脚本

## 支持的功能

### 文生图（Text-to-Image）
- 支持模型：`Qwen-Image`, `Qwen-Image-Plus`, `Z-Image-Turbo`, `Wan2.6-T2I`
- 支持正向提示词和反向提示词
- 支持自定义分辨率、种子值等参数

### 图生图（Image-to-Image）
- 支持模型：`Qwen-Image-Edit-2509`
- 支持输入 1-3 张参考图片（URL 或 Base64）
- 支持图像编辑和风格转换

### 高级参数
- `prompt_extend` - 智能提示词改写
- `watermark` - 添加水印
- `save_to_jpeg` - 输出 JPEG 格式
- `seed` - 固定随机种子以获得可重复的结果

## 使用方法

### 1. 配置渠道

在管理后台创建算能云渠道：
- **渠道类型**: Sophnet
- **Base URL**: `https://www.sophnet.com`
- **API Key**: 你的算能云 API Key
- **模型**: 选择需要启用的模型

### 2. 创建任务

```bash
curl -X POST "http://your-api-domain/imagegenerator/task" \
  -H "Authorization: Bearer sk-xxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "Qwen-Image",
    "input": {
      "prompt": "一只可爱的猫咪在弹钢琴"
    }
  }'
```

响应：
```json
{
  "code": "success",
  "message": "success",
  "data": "task_abc123xyz"
}
```

### 3. 查询任务

```bash
curl -X GET "http://your-api-domain/imagegenerator/task/task_abc123xyz" \
  -H "Authorization: Bearer sk-xxxx"
```

响应（成功）：
```json
{
  "code": "success",
  "message": "success",
  "data": {
    "task_id": "task_abc123xyz",
    "status": "success",
    "progress": "100%",
    "url": "https://example.com/generated-image.jpg"
  }
}
```

## 兼容性

除了新的 `/imagegenerator/task` 路径外，仍然支持旧的路径：
- `/sophnet/submit/generate` - 提交任务
- `/sophnet/fetch/{taskId}` - 查询任务

推荐使用新的 `/imagegenerator/task` 路径，更符合算能云官方 API 规范。

## 测试

### 使用测试脚本

Linux/Mac:
```bash
export API_BASE="http://localhost:3000"
export API_KEY="sk-your-api-key"
bash test_sophnet_image.sh
```

Windows:
```cmd
set API_BASE=http://localhost:3000
set API_KEY=sk-your-api-key
test_sophnet_image.bat
```

### 手动测试

1. 创建任务并记录返回的 task_id
2. 使用 task_id 轮询查询任务状态
3. 当状态变为 `success` 时，从响应中获取图片 URL

## 技术细节

### 任务状态映射

算能云状态 → new-api 状态：
- `PENDING` → `queued`
- `RUNNING` → `in_progress`
- `SUCCEEDED` → `success`
- `FAILED` → `failure`
- `CANCELED` → `failure`
- `UNKNOWN` → `unknown`

### 计费逻辑

- 任务提交时预扣费
- 任务完成后根据实际使用情况结算
- 任务失败时自动退款

### 错误处理

所有错误都会返回标准的 TaskError 格式：
```json
{
  "code": "error_code",
  "message": "error message",
  "status_code": 400
}
```

## 注意事项

1. **图编辑模型**: `Qwen-Image-Edit-2509` 必须提供 `input.images` 参数
2. **轮询频率**: 建议每 2-5 秒查询一次任务状态
3. **任务超时**: 如果任务长时间未完成，建议设置超时机制
4. **图片保存**: 生成的图片 URL 有有效期，建议及时下载保存

## 相关文件

- `router/relay-router.go` - 路由配置
- `middleware/distributor.go` - 请求分发逻辑
- `relay/channel/task/sophnet/adaptor.go` - 算能云适配器
- `relay/channel/task/sophnet/dto.go` - 数据传输对象
- `relay/channel/task/sophnet/constants.go` - 常量定义
- `relay/channel/task/sophnet/models.go` - 模型列表
- `controller/channel-test.go` - 渠道测试
- `docs/SOPHNET_IMAGE_GENERATOR.md` - API 文档

## 后续优化建议

1. 添加批量任务提交支持
2. 添加任务取消功能
3. 优化轮询机制（WebSocket 推送）
4. 添加图片缓存功能
5. 支持更多图片生成参数

## 问题反馈

如遇到问题，请检查：
1. 渠道配置是否正确
2. API Key 是否有效
3. 模型名称是否正确
4. 请求参数是否符合要求
5. 查看日志获取详细错误信息
