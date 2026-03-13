# 算能云 (Sophnet) 渠道接入说明

## 概述

算能云是一个异步图片和视频生成服务，支持多种图像和视频生成模型。本项目已完整集成算能云的异步任务接口。

## 支持的模型

### 图片生成模型
- `Qwen-Image` - 通义千问图像生成(文生图)
- `Qwen-Image-Plus` - 通义千问图像生成 Plus 版(文生图)
- `Qwen-Image-Edit-2509` - 通义千问图像编辑(图生图)
- `Z-Image-Turbo` - Z-Image 高速版(文生图)
- `Wan2.6-T2I` - 万相2.6文生图

### 视频生成模型
- `Qwen-Video-Turbo` - 通义千问视频生成 Turbo 版
- `Qwen-Video-Plus` - 通义千问视频生成 Plus 版

## 配置渠道

### 1. 添加渠道

在管理后台中添加新的渠道:

- **渠道类型**: 选择 "Sophnet"
- **Base URL**: `https://www.sophnet.com` (默认,可不填)
- **API Key**: 填入您的算能云 API Key (格式: `Bearer {your_api_key}`)
- **模型**: 从上述支持的模型列表中选择

### 2. 模型映射

系统会自动将 action 映射为模型名称:
- `sophnet_generate` - 默认的图片生成任务

## API 使用

### 提交图片生成任务

```bash
POST /sophnet/submit/generate
Authorization: Bearer {your_new_api_token}
Content-Type: application/json

{
  "model": "Qwen-Image",
  "input": {
    "prompt": "一只可爱的小猫在花园里玩耍",
    "negative_prompt": "模糊,低质量"  // 可选
  },
  "parameters": {
    "size": "1328*1328",        // 可选,默认 1328*1328
    "seed": 12345,              // 可选,随机种子
    "prompt_extend": true,      // 可选,是否开启 prompt 智能改写
    "watermark": false,         // 可选,是否添加水印
    "save_to_jpeg": false       // 可选,是否输出 jpg 格式
  }
}
```

响应示例:

```json
{
  "code": 200,
  "message": "success",
  "data": "task_abc123xyz..."
}
```

### 查询任务状态

```bash
GET /sophnet/fetch/{task_id}
Authorization: Bearer {your_new_api_token}
```

或

```bash
POST /sophnet/fetch
Authorization: Bearer {your_new_api_token}
Content-Type: application/json

{
  "task_id": "task_abc123xyz..."
}
```

响应示例(进行中):

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "task_id": "task_abc123xyz...",
    "status": "IN_PROGRESS",
    "progress": "50%"
  }
}
```

响应示例(成功):

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "task_id": "task_abc123xyz...",
    "status": "SUCCESS",
    "progress": "100%",
    "data": [
      {
        "url": "https://example.com/image1.jpg"
      }
    ]
  }
}
```

### 图生图 (Qwen-Image-Edit-2509)

```bash
POST /sophnet/submit/generate
Authorization: Bearer {your_new_api_token}
Content-Type: application/json

{
  "model": "Qwen-Image-Edit-2509",
  "input": {
    "prompt": "将图片改成油画风格",
    "images": [
      "https://example.com/input.jpg"
      // 或使用 Base64: "data:image/jpeg;base64,/9j/4AAQ..."
    ]
  }
}
```

## 任务状态说明

- `PENDING` - 任务已提交,等待处理
- `RUNNING` - 任务执行中
- `SUCCEEDED` - 任务成功完成
- `FAILED` - 任务失败
- `CANCELED` - 任务已取消

## 轮询机制

系统会自动每 15 秒轮询一次未完成的任务。当任务完成或失败后,会自动停止轮询。

## 计费说明

### 图片生成任务
- **预扣费**: 任务提交时根据模型价格预扣费
- **结算**: 图片任务按次计费，预扣费即为最终费用
- **退款**: 任务失败时自动全额退款

### 视频生成任务
- **预扣费**: 任务提交时根据请求的视频时长（duration 参数）预扣费
  - 公式: `预扣费 = 模型价格 × 分组倍率 × 模型倍率 × 请求时长（秒）`
- **差额结算**: 任务成功后，支持两种计费模式：
  
  **模式 1：按 Token 计费**（优先）
  - 如果上游返回 `usage.completion_tokens`，系统自动按 token 重新计算费用
  - 公式: `实际费用 = 模型价格 × 分组倍率 × 模型倍率 × completion_tokens`
  - 适用于支持 token 计费的模型
  
  **模式 2：按时长计费**（备选）
  - 如果没有 token 信息但有 `usage.duration`，按实际时长重新计算费用
  - 公式: `实际费用 = 模型价格 × 分组倍率 × 模型倍率 × 实际时长（秒）`
  - 适用于按时长计费的模型
  
  系统会自动选择合适的计费模式，并进行差额结算：
  - 如果实际费用 > 预扣费：补扣差额
  - 如果实际费用 < 预扣费：退还差额
  
- **退款**: 任务失败时自动全额退款

### 计费示例

假设视频模型价格为 10，分组倍率为 1.0，模型倍率为 1.0：

**示例 1：按 Token 计费**

1. **请求 5 秒视频**：
   - 预扣费: 10 × 1.0 × 1.0 × 5 = 50
   - 实际返回 1000 tokens：实际费用 = 10 × 1.0 × 1.0 × 1000 = 10000
   - 补扣差额: 10000 - 50 = 9950
   - 最终费用: 10000

**示例 2：按时长计费**

1. **请求 5 秒视频**（无 token 信息）：
   - 预扣费: 10 × 1.0 × 1.0 × 5 = 50
   - 实际生成 5 秒：最终费用 50（无差额）
   - 实际生成 6 秒：补扣 10，最终费用 60
   - 实际生成 4 秒：退还 10，最终费用 40

**示例 3：任务失败**

1. **任务失败**：
   - 预扣费: 50
   - 失败后退款: 50
   - 最终费用: 0

## 错误处理

### 错误返回机制

算能渠道的错误处理完整且可靠：

1. **提交阶段错误**（`DoResponse`）：
   - ✅ 图片任务：检查响应中的 `code` 字段
   - ✅ 视频任务：检查响应中的 `status` 字段（非 0 表示错误）
   - ✅ 空 taskId 检查
   - ✅ 所有错误都会返回给调用者，不会丢失

2. **轮询阶段错误**（`ParseTaskResult`）：
   - ✅ 解析错误响应（`status != 0 && result == null`）
   - ✅ 任务失败状态处理
   - ✅ 完整的响应体日志记录
   - ✅ 失败原因提取和返回

3. **错误日志**：
   - 所有错误都会记录到日志系统
   - 包含完整的响应体用于调试
   - 记录渠道 ID、状态码、错误信息等

### 常见错误码

- `invalid_request` - 请求参数错误（如缺少必填字段）
- `unmarshal_response_body_failed` - 响应解析失败
- `empty_task_id` - 上游未返回任务 ID
- `video_task_failed` - 视频任务失败
- `upstream_error` - 上游服务错误
- `task_timeout` - 任务超时
- `channel_not_found` - 渠道未找到或已禁用

### 错误响应示例

```json
{
  "code": "invalid_request",
  "message": "model is required",
  "status_code": 400
}
```

## OpenAI 兼容格式

算能渠道支持 OpenAI 格式的图片和视频生成请求：

### 图片生成（OpenAI 格式）

```bash
POST /v1/images/generations
Authorization: Bearer {your_new_api_token}
Content-Type: application/json

{
  "model": "Qwen-Image",
  "prompt": "一只可爱的小猫",
  "n": 1,
  "size": "1024x1024"
}
```

### 视频生成（OpenAI 格式）

```bash
POST /v1/videos/generations
Authorization: Bearer {your_new_api_token}
Content-Type: application/json

{
  "model": "Qwen-Video-Turbo",
  "prompt": "一只小猫在花园里玩耍",
  "duration": 5
}
```

系统会自动将 OpenAI 格式转换为算能格式。

## 注意事项

1. **图编辑模型必须提供输入图片**: `Qwen-Image-Edit-2509` 模型必须在 `input.images` 中提供 1-3 张图片
2. **视频时长参数**: 视频生成任务的 `duration` 参数会影响预扣费金额，建议根据实际需求设置
3. **支持的图片格式**: URL 或 Base64 编码
4. **任务超时**: 默认超时时间由系统配置决定,超时后任务会自动标记为失败并退款
5. **并发限制**: 建议控制并发请求数量,避免触发上游限流
6. **差额结算**: 视频任务完成后会根据实际时长自动进行差额结算，无需手动处理
7. **错误重试**: 系统会自动重试失败的请求（如果配置了重试策略）

## 扣费逻辑验证

### 图片任务
- ✅ 提交时预扣费（基于模型价格）
- ✅ 成功时保持预扣费不变
- ✅ 失败时全额退款

### 视频任务
- ✅ 提交时预扣费（基于请求时长）
- ✅ 成功时根据实际消耗差额结算（支持两种模式）
  - ✅ **模式 1**：按 token 计费（优先）- 提取 `usage.completion_tokens`，系统自动重算
  - ✅ **模式 2**：按时长计费（备选）- 提取 `usage.duration`，手动计算实际费用
- ✅ 失败时全额退款
- ✅ 自动选择合适的计费模式

### 计费模式选择逻辑
```
任务完成 → 检查 taskResult.TotalTokens
   ↓
有 token？
   ├─ 是 → 返回 0，系统按 token 自动重算（模式 1）
   └─ 否 → 检查 usage.duration
       ├─ 有时长 → 手动计算实际费用并返回（模式 2）
       └─ 无信息 → 返回 0，保持预扣费不变
```

### 错误返回验证
- ✅ 所有请求阶段的错误都会返回给调用者
- ✅ 所有轮询阶段的错误都会正确处理
- ✅ 错误信息包含详细的上游响应
- ✅ 完整的日志记录用于调试

## 开发者信息

- **渠道代码**: `constant.ChannelTypeSophnet = 58`
- **平台标识**: `constant.TaskPlatformSophnet = "sophnet"`
- **适配器路径**: `relay/channel/task/sophnet/`
- **路由前缀**: `/sophnet/`

## 更新日志

- **2026-03-04**: 初始版本,支持文生图和图生图功能
