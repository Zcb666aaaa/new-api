# 算能云异步图片生成 - 问题修复报告

## 修复日期
2026年3月4日

## 问题描述

用户在使用 `/imagegenerator/task` 接口时遇到两个错误：

### 错误 1: `invalid_api_platform: sophnet`
```json
{
  "code": "invalid_api_platform",
  "message": "invalid api platform: sophnet",
  "data": null
}
```

### 错误 2: `model_not_found: sophnet_generate`
```json
{
  "error": {
    "code": "model_not_found",
    "message": "No available channel for model sophnet_generate under group default",
    "type": "new_api_error"
  }
}
```

## 根本原因

### 原因 1: 平台标识设置错误
在 `middleware/distributor.go` 中，`/imagegenerator/task` 路径使用了：
```go
c.Set("platform", string(constant.TaskPlatformSophnet))  // ❌ 错误
```

这导致 `GetTaskAdaptor` 函数无法识别平台，因为它期望的是渠道类型（整数），而不是平台字符串。

### 原因 2: 模型名称硬编码
在 `middleware/distributor.go` 中，模型名称被硬编码为：
```go
modelName := "sophnet_generate"  // ❌ 错误
```

这导致系统尝试查找名为 `sophnet_generate` 的模型，而实际上用户请求的是 `Qwen-Image`。

## 修复方案

### 修复 1: 使用渠道类型而不是平台字符串

**修改文件**: `middleware/distributor.go`

**修改前**:
```go
c.Set("platform", string(constant.TaskPlatformSophnet))
```

**修改后**:
```go
c.Set("channel_type", constant.ChannelTypeSophnet)
```

**原理**: 
- `GetTaskPlatform` 函数首先检查 `channel_type`（整数）
- 如果存在，将其转换为字符串作为平台标识
- 这样 `GetTaskAdaptor` 就能正确识别并返回 sophnet 适配器

### 修复 2: 从请求体中提取实际的模型名称

**修改文件**: `middleware/distributor.go`

**修改前**:
```go
if c.Request.Method == http.MethodPost {
    relayMode = relayconstant.RelayModeVideoSubmit
    modelName := "sophnet_generate"
    modelRequest.Model = modelName
}
```

**修改后**:
```go
if c.Request.Method == http.MethodPost {
    relayMode = relayconstant.RelayModeVideoSubmit
    // 从请求体中提取模型名称
    req, err := getModelFromRequest(c)
    if err != nil {
        return nil, false, err
    }
    if req != nil {
        modelRequest.Model = req.Model
    }
}
```

**原理**:
- `getModelFromRequest` 函数解析请求体并提取 `model` 字段
- 这样系统就能使用用户实际请求的模型名称（如 `Qwen-Image`）
- 渠道选择器会根据这个模型名称找到正确的算能云渠道

## 完整的修复代码

```go
} else if strings.Contains(c.Request.URL.Path, "/imagegenerator/task") {
    // 算能云异步图片生成路径
    var relayMode int
    if c.Request.Method == http.MethodPost {
        relayMode = relayconstant.RelayModeVideoSubmit
        // 从请求体中提取模型名称
        req, err := getModelFromRequest(c)
        if err != nil {
            return nil, false, err
        }
        if req != nil {
            modelRequest.Model = req.Model
        }
    } else if c.Request.Method == http.MethodGet {
        relayMode = relayconstant.RelayModeVideoFetchByID
        shouldSelectChannel = false
    }
    c.Set("channel_type", constant.ChannelTypeSophnet)
    c.Set("relay_mode", relayMode)
```

## 验证步骤

### 1. 配置渠道
在管理后台创建算能云渠道，添加以下模型：
- `Qwen-Image`
- `Qwen-Image-Plus`
- `Qwen-Image-Edit-2509`
- `Z-Image-Turbo`
- `Wan2.6-T2I`

### 2. 测试请求

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

**期望结果**:
```json
{
  "code": "success",
  "message": "success",
  "data": "task_abc123xyz"
}
```

### 3. 查询任务

```bash
curl -X GET "http://localhost:3000/imagegenerator/task/task_abc123xyz" \
  -H "Authorization: Bearer sk-xxxx"
```

## 相关文件

### 修改的文件
- `middleware/distributor.go` - 修复平台标识和模型名称提取

### 更新的文档
- `docs/SOPHNET_IMAGE_GENERATOR.md` - 更新渠道配置说明
- `docs/SOPHNET_IMAGE_QUICKSTART.md` - 更新快速开始指南
- `docs/SOPHNET_IMAGE_TROUBLESHOOTING.md` - 新增故障排除文档

## 重要提示

### 对用户的要求

1. **必须使用实际的模型名称**
   - ✅ 正确: `"model": "Qwen-Image"`
   - ❌ 错误: `"model": "sophnet_generate"`

2. **必须在渠道中配置实际的模型**
   - 在管理后台的渠道配置中添加 `Qwen-Image` 等实际模型名称
   - 不要添加 `sophnet_generate`

3. **请求路径**
   - 推荐使用: `/imagegenerator/task`
   - 也支持: `/sophnet/submit/generate` 和 `/sophnet/fetch/{id}`

## 测试结果

- ✅ 平台识别正确
- ✅ 模型名称提取正确
- ✅ 渠道选择正确
- ✅ 任务创建成功
- ✅ 任务查询成功

## 后续建议

1. 添加更详细的错误提示，帮助用户理解模型配置要求
2. 在管理后台添加模型配置向导
3. 添加模型名称验证，防止用户使用错误的模型名称

## 总结

通过这两个关键修复：
1. 使用 `channel_type` 而不是 `platform` 字符串
2. 从请求体中提取实际的模型名称

算能云异步图片生成功能现在可以正常工作了。用户只需要：
- 在渠道中配置实际的模型名称（如 `Qwen-Image`）
- 在请求中使用相同的模型名称
- 使用正确的请求路径 `/imagegenerator/task`

---

**状态**: ✅ 已修复
**测试**: ✅ 已验证
**文档**: ✅ 已更新
