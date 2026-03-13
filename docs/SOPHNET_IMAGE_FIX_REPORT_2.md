# 算能云异步图片生成 - 任务查询修复报告

## 修复日期
2026年3月4日

## 问题描述

用户在创建任务后，使用返回的 task_id 查询任务状态时，收到以下错误：

```json
{
  "code": "task_not_exist",
  "message": "task_not_exist",
  "data": null
}
```

**请求示例:**
```bash
# 创建任务成功，返回 task_id
POST /imagegenerator/task
Response: {"code":"success","data":"task_abc123xyz"}

# 查询任务失败
GET /imagegenerator/task/task_abc123xyz
Response: {"code":"task_not_exist","message":"task_not_exist","data":null}
```

## 根本原因

在 `relay/relay_task.go` 文件的 `videoFetchByIDRespBodyBuilder` 函数中，获取任务 ID 的逻辑存在问题：

**问题代码:**
```go
func videoFetchByIDRespBodyBuilder(c *gin.Context) (respBody []byte, taskResp *dto.TaskError) {
    taskId := c.Param("task_id")  // ❌ 错误：路由参数名是 "id"，不是 "task_id"
    if taskId == "" {
        taskId = c.GetString("task_id")
    }
    userId := c.GetInt("id")
    // ...
}
```

**路由配置:**
```go
relayImageGeneratorRouter.GET("/task/:id", controller.RelayTaskFetch)
//                                    ^^^ 参数名是 "id"
```

**问题分析:**
1. 路由定义使用 `:id` 作为参数名
2. 但函数尝试获取 `task_id` 参数
3. 导致 `taskId` 始终为空字符串
4. 查询数据库时找不到任务记录

## 修复方案

修改 `videoFetchByIDRespBodyBuilder` 函数，优先尝试获取 `:id` 参数：

**修复代码:**
```go
func videoFetchByIDRespBodyBuilder(c *gin.Context) (respBody []byte, taskResp *dto.TaskError) {
    taskId := c.Param("id")           // ✅ 首先尝试获取 "id" 参数
    if taskId == "" {
        taskId = c.Param("task_id")   // ✅ 兼容其他路由的 "task_id" 参数
    }
    if taskId == "" {
        taskId = c.GetString("task_id")  // ✅ 最后尝试从 context 获取
    }
    userId := c.GetInt("id")
    // ...
}
```

**修复说明:**
1. 首先尝试获取 `:id` 参数（用于 `/imagegenerator/task/:id` 路由）
2. 如果为空，尝试获取 `:task_id` 参数（兼容其他路由，如 `/v1/videos/:task_id`）
3. 如果还是为空，从 context 中获取（兼容某些特殊情况）
4. 这样可以兼容所有的路由配置

## 影响范围

### 受影响的路由
- ✅ `/imagegenerator/task/:id` - 算能云异步图片生成（已修复）
- ✅ `/v1/videos/:task_id` - OpenAI 视频 API（保持兼容）
- ✅ `/sophnet/fetch/:id` - 算能云旧路由（保持兼容）

### 不受影响的功能
- ✅ 任务创建功能正常
- ✅ 其他查询路由正常
- ✅ Suno 任务查询正常

## 测试验证

### 测试步骤

1. **创建任务**
```bash
curl -X POST "http://localhost:3000/imagegenerator/task" \
  -H "Authorization: Bearer sk-xxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "Qwen-Image",
    "input": {
      "prompt": "一只可爱的猫咪在弹钢琴"
    }
  }'
```

**期望响应:**
```json
{
  "code": "success",
  "message": "success",
  "data": "task_abc123xyz"
}
```

2. **查询任务**
```bash
curl -X GET "http://localhost:3000/imagegenerator/task/task_abc123xyz" \
  -H "Authorization: Bearer sk-xxxx"
```

**期望响应（任务进行中）:**
```json
{
  "code": "success",
  "message": "success",
  "data": {
    "task_id": "task_abc123xyz",
    "status": "in_progress",
    "progress": "50%"
  }
}
```

**期望响应（任务成功）:**
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

### 测试结果

- ✅ 任务创建成功
- ✅ 任务查询成功
- ✅ 任务状态正确返回
- ✅ 兼容其他路由

## 相关文件

### 修改的文件
- `relay/relay_task.go` - 修复任务 ID 获取逻辑

### 更新的文档
- `docs/SOPHNET_IMAGE_TROUBLESHOOTING.md` - 添加此问题的说明

## 完整的工作流程

### 1. 创建任务
```bash
POST /imagegenerator/task
Body: {"model":"Qwen-Image","input":{"prompt":"..."}}
Response: {"code":"success","data":"task_abc123xyz"}
```

### 2. 查询任务（轮询）
```bash
GET /imagegenerator/task/task_abc123xyz
Response: {"code":"success","data":{"status":"in_progress",...}}
```

### 3. 获取结果
```bash
GET /imagegenerator/task/task_abc123xyz
Response: {"code":"success","data":{"status":"success","url":"..."}}
```

## 注意事项

### 对用户的影响
- ✅ 无需修改客户端代码
- ✅ 现有的请求格式保持不变
- ✅ 向后兼容

### 路由参数命名
不同的路由使用不同的参数名：
- `/imagegenerator/task/:id` - 使用 `:id`
- `/v1/videos/:task_id` - 使用 `:task_id`
- `/sophnet/fetch/:id` - 使用 `:id`

修复后的代码能够自动适配所有这些路由。

## 其他相关修复

本次修复是继以下修复之后的第三个修复：

### 修复 1: 平台标识问题
- 将 `c.Set("platform", ...)` 改为 `c.Set("channel_type", ...)`
- 解决了 `invalid_api_platform` 错误

### 修复 2: 模型名称问题
- 从请求体中动态提取模型名称
- 解决了 `model_not_found: sophnet_generate` 错误

### 修复 3: 任务查询问题（本次）
- 修复任务 ID 参数获取逻辑
- 解决了 `task_not_exist` 错误

## 总结

通过修改 `videoFetchByIDRespBodyBuilder` 函数，优先尝试获取 `:id` 参数，解决了任务查询失败的问题。

修复后的代码：
- ✅ 支持 `/imagegenerator/task/:id` 路由
- ✅ 兼容 `/v1/videos/:task_id` 路由
- ✅ 兼容其他使用 `:id` 的路由
- ✅ 向后兼容，不影响现有功能

用户现在可以正常创建和查询算能云异步图片生成任务了！

---

**状态**: ✅ 已修复
**测试**: ✅ 已验证
**文档**: ✅ 已更新
