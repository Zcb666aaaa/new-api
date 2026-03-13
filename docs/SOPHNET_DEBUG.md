# 算能云渠道调试指南

## 问题排查步骤

### 第一步：检查渠道配置

1. **登录管理后台**，进入渠道管理页面

2. **检查算能云渠道配置**：
   - 类型：必须是 "算能"（type = 58）
   - 状态：必须是"启用"
   - Base URL：`https://www.sophnet.com`（或留空使用默认值）
   - 密钥：您的算能云 API Key
   - 模型：必须包含 `sophnet_generate`

3. **SQL 查询验证**（如果有数据库访问权限）：
```sql
SELECT id, name, type, status, base_url, models 
FROM channels 
WHERE type = 58;
```

应该看到类似输出：
```
id | name      | type | status | base_url                  | models
1  | 算能云    | 58   | 1      | https://www.sophnet.com   | sophnet_generate
```

### 第二步：检查请求路径

算能云使用以下路径结构：

**提交任务：**
```
POST /sophnet/submit/generate
```

**查询任务：**
```
GET /sophnet/fetch/{task_id}
或
POST /sophnet/fetch
```

### 第三步：验证请求格式

**正确的请求示例：**

```bash
curl -X POST "http://your-domain/sophnet/submit/generate" \
  -H "Authorization: Bearer sk-your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "Qwen-Image",
    "input": {
      "prompt": "一只可爱的小猫"
    }
  }'
```

**注意事项：**
- URL 路径：`/sophnet/submit/generate`（不是 `/api/sophnet/...`）
- 请求体中的 `model` 字段：使用上游模型名（如 `Qwen-Image`）
- 渠道配置中的模型：使用 `sophnet_generate`

### 第四步：查看日志

#### 后端日志位置
检查以下日志文件（根据您的部署方式）：
- Docker: `docker logs new-api`
- 系统服务: `/var/log/new-api/`
- 开发模式: 控制台输出

#### 关键日志关键词
搜索以下关键词：
```
sophnet
task_id
BuildRequestURL
model_not_found
channel not found
```

#### 常见错误日志及解决方案

**错误1: model_not_found**
```json
{
  "error": {
    "code": "model_not_found",
    "message": "No available channel for model sophnet_generate"
  }
}
```
**解决方案：** 在渠道的"模型"字段中添加 `sophnet_generate`

**错误2: channel_not_found**
```json
{
  "error": {
    "code": "channel_not_found",
    "message": "No available channel"
  }
}
```
**解决方案：** 
- 检查渠道是否启用
- 检查用户分组是否匹配
- 检查渠道优先级和权重

**错误3: invalid_api_key**
```json
{
  "error": {
    "code": "invalid_api_key",
    "message": "Invalid API key"
  }
}
```
**解决方案：** 检查算能云 API Key 是否正确

### 第五步：使用测试脚本

运行提供的测试脚本：

```bash
# 修改脚本中的配置
vim test_sophnet.sh

# 添加执行权限
chmod +x test_sophnet.sh

# 运行测试
./test_sophnet.sh
```

### 第六步：手动测试完整流程

#### 1. 获取渠道列表
```bash
curl -X GET "http://your-domain/api/channel/" \
  -H "Authorization: Bearer your-admin-token"
```

查找 type=58 的渠道，确认：
- status = 1（启用）
- models 包含 "sophnet_generate"
- base_url 正确

#### 2. 提交任务
```bash
curl -v -X POST "http://your-domain/sophnet/submit/generate" \
  -H "Authorization: Bearer your-user-token" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "Qwen-Image",
    "input": {
      "prompt": "测试图片"
    }
  }'
```

**预期响应：**
```json
{
  "code": "success",
  "message": "success",
  "data": "task_abc123xyz..."
}
```

#### 3. 查询任务
```bash
curl -X GET "http://your-domain/sophnet/fetch/task_abc123xyz..." \
  -H "Authorization: Bearer your-user-token"
```

**预期响应（进行中）：**
```json
{
  "code": 200,
  "data": {
    "task_id": "task_abc123xyz...",
    "status": "IN_PROGRESS",
    "progress": "50%"
  }
}
```

**预期响应（成功）：**
```json
{
  "code": 200,
  "data": {
    "task_id": "task_abc123xyz...",
    "status": "SUCCESS",
    "progress": "100%",
    "result_url": "https://example.com/image.jpg"
  }
}
```

### 第七步：检查上游 API

直接测试算能云 API 是否正常：

```bash
curl -X POST "https://www.sophnet.com/api/open-apis/projects/easyllms/imagegenerator/task" \
  -H "Authorization: Bearer your-sophnet-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "Qwen-Image",
    "input": {
      "prompt": "测试"
    }
  }'
```

如果上游 API 返回错误，说明问题在算能云侧。

### 常见问题汇总

#### Q1: 提示 "model_not_found"
**原因：** 渠道未配置 `sophnet_generate` 模型
**解决：** 编辑渠道，在模型字段添加 `sophnet_generate`

#### Q2: 提示 "channel_not_found"
**原因：** 
- 渠道未启用
- 用户分组不匹配
- 渠道类型错误

**解决：**
1. 确认渠道状态为"启用"
2. 检查用户所属分组与渠道分组是否匹配
3. 确认渠道类型为 58（算能）

#### Q3: 提示 "invalid_request"
**原因：** 请求参数错误
**解决：** 检查请求体格式，确保包含必填字段：
- `model`: 上游模型名称
- `input.prompt`: 提示词

#### Q4: 任务一直处于 PENDING 状态
**原因：** 轮询服务未启动或上游处理慢
**解决：**
1. 检查后端日志，确认轮询服务正在运行
2. 等待更长时间（图片生成通常需要 10-60 秒）
3. 检查上游 API 是否正常

#### Q5: 请求超时
**原因：** 网络问题或上游服务慢
**解决：**
1. 检查服务器到算能云的网络连接
2. 增加超时时间配置
3. 检查是否需要配置代理

### 调试技巧

#### 1. 启用详细日志
在环境变量中设置：
```bash
export LOG_LEVEL=debug
```

#### 2. 使用 curl -v 查看详细请求
```bash
curl -v -X POST "..." \
  -H "..." \
  -d '...'
```

#### 3. 检查网络连通性
```bash
# 测试到算能云的连接
curl -I https://www.sophnet.com

# 测试 DNS 解析
nslookup www.sophnet.com
```

#### 4. 查看数据库任务记录
```sql
SELECT * FROM tasks 
WHERE platform = 'sophnet' 
ORDER BY created_at DESC 
LIMIT 10;
```

### 获取帮助

如果以上步骤都无法解决问题，请提供以下信息：

1. **错误信息**：完整的错误响应
2. **渠道配置**：渠道的完整配置（隐藏敏感信息）
3. **请求示例**：您发送的完整请求（隐藏 token）
4. **后端日志**：相关的后端日志片段
5. **环境信息**：部署方式、版本号等

---

## 快速检查清单

- [ ] 渠道类型是 58（算能）
- [ ] 渠道状态是"启用"
- [ ] Base URL 是 `https://www.sophnet.com`
- [ ] 模型字段包含 `sophnet_generate`
- [ ] API Key 正确且有效
- [ ] 用户分组匹配
- [ ] 请求路径是 `/sophnet/submit/generate`
- [ ] 请求体包含 `model` 和 `input.prompt`
- [ ] 后端服务正常运行
- [ ] 网络连接正常

全部勾选后，应该可以正常使用！
