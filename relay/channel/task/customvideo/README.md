# 自定义视频渠道 (Custom Video Channel)

## 概述

自定义视频渠道是一个透传式的异步视频生成渠道，允许你接入任何符合标准格式的视频生成服务。

## 特性

- ✅ 请求透传：直接转发客户端请求到上游服务
- ✅ 异步任务：支持异步视频生成任务
- ✅ 完善的扣费机制：基于时长和其他参数的灵活计费
- ✅ 标准化接口：符合 OpenAI 视频 API 规范

## API 端点

### 创建视频任务

**对外请求地址：**
```
POST /v1/videos/create
```

**转发地址：**
```
POST {baseURL}/v1/videos/create
```

**请求格式：**
```json
{
  "model": "custom-video-model",
  "prompt": "视频描述文本",
  "size": "1920x1080",
  "duration": 5,
  "metadata": {
    "custom_param": "value"
  }
}
```

**响应格式：**
```json
{
  "id": "task_123456",
  "task_id": "task_123456",
  "model": "custom-video-model",
  "status": "queued",
  "created_at": 1234567890
}
```

### 查询视频任务

**对外请求地址：**
```
GET /v1/videos/{task_id}
```

**转发地址：**
```
GET {baseURL}/v1/videos/{task_id}
```

**响应格式：**
```json
{
  "id": "task_123456",
  "status": "completed",
  "video_url": "https://example.com/video.mp4",
  "progress": "100%",
  "created_at": 1234567890
}
```

## 上游服务要求

### 创建任务接口

上游服务需要实现 `POST /v1/videos/create` 接口：

**请求头：**
```
Authorization: Bearer {api_key}
Content-Type: application/json
```

**请求体：**
```json
{
  "model": "string",
  "prompt": "string",
  "size": "string (optional)",
  "duration": "integer (optional)",
  "metadata": "object (optional)"
}
```

**响应体：**
```json
{
  "id": "string (必需，任务ID)",
  "status": "string (必需，任务状态)",
  "message": "string (可选，错误信息)",
  "video_url": "string (可选，视频URL)",
  "progress": "string (可选，进度)",
  "metadata": "object (可选)",
  "created_at": "integer (可选，时间戳)"
}
```

### 查询任务接口

上游服务需要实现 `GET /v1/videos/{id}` 接口：

**请求头：**
```
Authorization: Bearer {api_key}
```

**响应体：**
```json
{
  "id": "string (必需，任务ID)",
  "status": "string (必需，任务状态)",
  "message": "string (可选，错误信息)",
  "video_url": "string (可选，视频URL)",
  "progress": "string (可选，进度)",
  "metadata": "object (可选)",
  "created_at": "integer (可选，时间戳)"
}
```

### 状态映射

上游服务返回的状态会被映射为标准状态：

| 上游状态 | 标准状态 | 说明 |
|---------|---------|------|
| `queued`, `pending` | `queued` | 排队中 |
| `processing`, `running`, `in_progress` | `in_progress` | 处理中 |
| `completed`, `succeeded`, `success` | `completed` | 已完成 |
| `failed`, `error`, `canceled`, `cancelled` | `failed` | 失败 |

## 配置渠道

### 1. 添加渠道

在管理后台添加新渠道：

- **渠道类型**：选择 `CustomVideo`
- **Base URL**：填写上游服务的基础地址（例如：`https://api.example.com`）
- **API Key**：填写上游服务的认证密钥
- **模型列表**：填写支持的模型名称（例如：`custom-video-model`）

### 2. 配置计费

在模型价格配置中设置：

- **基础价格**：每次任务的基础费用
- **时长倍率**：`seconds` - 按秒计费的倍率
- **其他倍率**：可根据需要添加自定义倍率

**计费公式：**
```
总费用 = 基础价格 × seconds倍率 × 时长(秒)
```

## 使用示例

### 创建视频任务

```bash
curl -X POST https://your-api.com/v1/videos/create \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "custom-video-model",
    "prompt": "一只可爱的猫咪在草地上玩耍",
    "size": "1920x1080",
    "duration": 5
  }'
```

### 查询任务状态

```bash
curl -X GET https://your-api.com/v1/videos/task_123456 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 扩展功能

### 自定义参数

通过 `metadata` 字段传递自定义参数：

```json
{
  "model": "custom-video-model",
  "prompt": "视频描述",
  "metadata": {
    "style": "anime",
    "fps": 30,
    "quality": "high"
  }
}
```

### 自定义计费

在 `EstimateBilling` 方法中实现自定义计费逻辑：

```go
func (a *TaskAdaptor) EstimateBilling(c *gin.Context, info *relaycommon.RelayInfo) map[string]float64 {
    taskReq, _ := relaycommon.GetTaskRequest(c)
    
    otherRatios := map[string]float64{
        "seconds": float64(taskReq.Duration),
        "resolution": getResolutionRatio(taskReq.Size),
    }
    
    return otherRatios
}
```

## 故障排查

### 常见问题

1. **任务创建失败**
   - 检查 Base URL 是否正确
   - 检查 API Key 是否有效
   - 查看上游服务日志

2. **任务状态不更新**
   - 确认上游服务正确实现了查询接口
   - 检查任务轮询是否正常运行
   - 查看系统日志

3. **视频URL无法访问**
   - 确认上游服务返回的 `video_url` 可访问
   - 检查视频代理配置
   - 验证网络连接

## 技术架构

```
客户端请求
    ↓
/v1/videos/create (new-api)
    ↓
透传到上游服务
    ↓
{baseURL}/v1/videos/create
    ↓
返回任务ID
    ↓
异步轮询任务状态
    ↓
{baseURL}/v1/videos/{id}
    ↓
更新任务状态和结果
```

## 开发指南

如需修改自定义视频渠道的行为，请编辑以下文件：

- `relay/channel/task/customvideo/adaptor.go` - 核心适配器逻辑
- `relay/channel/task/customvideo/constants.go` - 常量定义
- `constant/channel.go` - 渠道类型定义
- `router/video-router.go` - 路由配置

## 许可证

遵循项目主许可证。
