# 自定义视频渠道实现总结

## 实现概述

已成功实现自定义透传视频渠道（CustomVideo），支持异步任务形式的视频生成服务接入。

## 核心特性

✅ **请求透传**：直接转发客户端请求到上游服务，无需格式转换  
✅ **异步任务**：完整的异步任务生命周期管理（创建、轮询、完成）  
✅ **完善扣费**：基于时长的灵活计费机制，支持自定义倍率  
✅ **标准接口**：符合 OpenAI 视频 API 规范  
✅ **简单配置**：只需配置 baseURL 和 API Key 即可使用  

## 文件修改清单

### 1. 常量定义
**文件**: `constant/channel.go`
- 新增 `ChannelTypeCustomVideo = 59` 渠道类型常量
- 添加到 `ChannelBaseURLs` 数组（索引59，空字符串）
- 添加到 `ChannelTypeNames` 映射（"CustomVideo"）

### 2. 核心适配器
**文件**: `relay/channel/task/customvideo/adaptor.go`
- 实现 `TaskAdaptor` 接口的所有方法
- 支持请求透传和响应解析
- 实现计费逻辑（基于时长）
- 支持任务状态映射和轮询

**主要方法**:
- `Init()` - 初始化适配器
- `ValidateRequestAndSetAction()` - 验证请求
- `BuildRequestURL()` - 构建请求URL（baseURL + /v1/videos/create）
- `BuildRequestHeader()` - 设置请求头（Authorization）
- `BuildRequestBody()` - 构建请求体（透传格式）
- `EstimateBilling()` - 预估计费（基于时长）
- `DoRequest()` - 发送请求
- `DoResponse()` - 处理响应
- `FetchTask()` - 查询任务（baseURL + /v1/videos/{id}）
- `ParseTaskResult()` - 解析任务结果
- `ConvertToOpenAIVideo()` - 转换为 OpenAI 格式

### 3. 常量配置
**文件**: `relay/channel/task/customvideo/constants.go`
- 定义模型列表：`custom-video-model`
- 定义渠道名称：`customvideo`

### 4. 适配器注册
**文件**: `relay/relay_adaptor.go`
- 导入自定义视频适配器包
- 在 `GetTaskAdaptor()` 中注册 `ChannelTypeCustomVideo` 分支

### 5. 路由配置
**文件**: `router/video-router.go`
- 添加 `/v1/videos/create` 创建任务路由
- 添加 `/v1/videos/:task_id` 查询任务路由
- 使用标准中间件（TokenAuth, Distribute）

### 6. 测试排除
**文件**: `controller/channel-test.go`
- 将 `ChannelTypeCustomVideo` 添加到不支持测试的渠道列表

### 7. 视频代理
**文件**: `controller/video_proxy.go`
- 添加 `ChannelTypeCustomVideo` 分支
- 直接使用存储的视频URL（不需要特殊处理）

### 8. 文档
**文件**: `relay/channel/task/customvideo/README.md`
- 完整的功能说明文档
- API 接口规范
- 上游服务要求
- 配置指南

**文件**: `relay/channel/task/customvideo/EXAMPLES.md`
- 详细的使用示例（Python, JavaScript, cURL）
- 上游服务实现示例（Flask, Express）
- 高级用法和故障排查

## API 端点

### 对外接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/v1/videos/create` | 创建视频任务 |
| GET | `/v1/videos/{task_id}` | 查询任务状态 |

### 转发规则

| 对外路径 | 转发路径 | 说明 |
|---------|---------|------|
| `/v1/videos/create` | `{baseURL}/v1/videos/create` | 创建任务 |
| `/v1/videos/{id}` | `{baseURL}/v1/videos/{id}` | 查询任务，{id}直接拼接 |

## 请求/响应格式

### 创建任务请求
```json
{
  "model": "custom-video-model",
  "prompt": "视频描述",
  "size": "1920x1080",
  "duration": 5,
  "metadata": {}
}
```

### 创建任务响应
```json
{
  "id": "task_xxx",
  "task_id": "task_xxx",
  "model": "custom-video-model",
  "status": "queued",
  "created_at": 1234567890
}
```

### 查询任务响应
```json
{
  "id": "task_xxx",
  "status": "completed",
  "video_url": "https://...",
  "progress": "100%",
  "metadata": {
    "url": "https://..."
  }
}
```

## 状态映射

| 上游状态 | 标准状态 | 进度 |
|---------|---------|------|
| queued, pending | queued | 10% |
| processing, running, in_progress | in_progress | 50% |
| completed, succeeded, success | completed | 100% |
| failed, error, canceled, cancelled | failed | 100% |

## 计费机制

### 预估计费（EstimateBilling）
在任务创建时根据请求参数计算：
```go
otherRatios := map[string]float64{
    "seconds": float64(duration), // 默认5秒
}
```

### 计费公式
```
总费用 = 基础价格 × seconds倍率 × 时长(秒)
```

### 示例
- 基础价格：0.1 元/次
- seconds倍率：0.02
- 时长：5秒
- **总费用** = 0.1 × 0.02 × 5 = 0.01 元

## 配置示例

### 渠道配置
```yaml
渠道类型: CustomVideo
Base URL: https://api.example.com
API Key: sk-xxxxxxxxxxxxx
模型列表: custom-video-model
```

### 价格配置
```yaml
模型: custom-video-model
基础价格: 0.1
其他倍率:
  seconds: 0.02
```

## 上游服务要求

### 必需实现的接口

1. **POST /v1/videos/create**
   - 接收：JSON 格式的视频生成请求
   - 返回：包含任务ID的响应
   - 认证：Bearer Token

2. **GET /v1/videos/{id}**
   - 接收：任务ID（URL路径参数）
   - 返回：任务状态和结果
   - 认证：Bearer Token

### 响应字段要求

**必需字段**:
- `id` (string) - 任务ID
- `status` (string) - 任务状态

**可选字段**:
- `video_url` (string) - 视频URL
- `progress` (string) - 进度百分比
- `message` (string) - 错误信息
- `metadata` (object) - 额外元数据
- `created_at` (integer) - 创建时间戳

## 使用流程

```
1. 客户端调用 POST /v1/videos/create
   ↓
2. new-api 验证请求并预扣费
   ↓
3. 透传请求到 {baseURL}/v1/videos/create
   ↓
4. 上游服务返回任务ID
   ↓
5. new-api 返回标准格式响应给客户端
   ↓
6. 后台异步轮询 GET {baseURL}/v1/videos/{id}
   ↓
7. 任务完成后更新状态和视频URL
   ↓
8. 客户端查询 GET /v1/videos/{task_id} 获取结果
```

## 测试建议

### 1. 单元测试
- 测试请求构建逻辑
- 测试响应解析逻辑
- 测试状态映射

### 2. 集成测试
- 测试完整的任务创建流程
- 测试任务轮询机制
- 测试计费逻辑

### 3. 端到端测试
- 使用真实的上游服务
- 测试各种状态转换
- 测试错误处理

## 扩展建议

### 1. 支持更多参数
在 `CustomVideoRequest` 中添加更多字段：
```go
type CustomVideoRequest struct {
    Model      string         `json:"model"`
    Prompt     string         `json:"prompt"`
    Size       string         `json:"size,omitempty"`
    Duration   *int           `json:"duration,omitempty"`
    Style      string         `json:"style,omitempty"`      // 新增
    FPS        *int           `json:"fps,omitempty"`        // 新增
    Quality    string         `json:"quality,omitempty"`    // 新增
    Metadata   map[string]any `json:"metadata,omitempty"`
}
```

### 2. 支持更复杂的计费
在 `EstimateBilling` 中实现：
```go
func (a *TaskAdaptor) EstimateBilling(c *gin.Context, info *relaycommon.RelayInfo) map[string]float64 {
    taskReq, _ := relaycommon.GetTaskRequest(c)
    
    otherRatios := map[string]float64{
        "seconds": float64(taskReq.Duration),
    }
    
    // 根据分辨率调整倍率
    if strings.Contains(taskReq.Size, "1080") {
        otherRatios["resolution"] = 2.0
    } else if strings.Contains(taskReq.Size, "720") {
        otherRatios["resolution"] = 1.5
    }
    
    return otherRatios
}
```

### 3. 支持完成后调整计费
实现 `AdjustBillingOnComplete`：
```go
func (a *TaskAdaptor) AdjustBillingOnComplete(task *model.Task, taskResult *relaycommon.TaskInfo) int {
    // 根据实际生成的视频时长调整费用
    if actualDuration := extractActualDuration(task.Data); actualDuration > 0 {
        return calculateActualQuota(actualDuration)
    }
    return 0
}
```

## 注意事项

1. **安全性**
   - 确保上游服务使用 HTTPS
   - 妥善保管 API Key
   - 验证上游响应的合法性

2. **性能**
   - 合理设置轮询间隔
   - 避免过于频繁的查询
   - 考虑使用 webhook 替代轮询

3. **错误处理**
   - 处理网络超时
   - 处理上游服务错误
   - 提供清晰的错误信息

4. **兼容性**
   - 遵循项目规范（Rule 1-6）
   - 使用 `common.Marshal/Unmarshal`
   - 支持所有数据库（SQLite, MySQL, PostgreSQL）

## 总结

自定义视频渠道已完整实现，具备以下优势：

- ✅ **简单易用**：只需配置 baseURL 和 API Key
- ✅ **灵活透传**：直接转发请求，无需复杂转换
- ✅ **完善计费**：支持基于时长的灵活计费
- ✅ **标准接口**：符合 OpenAI 视频 API 规范
- ✅ **易于扩展**：可轻松添加新功能和参数

可以立即开始使用，接入任何符合规范的视频生成服务！
