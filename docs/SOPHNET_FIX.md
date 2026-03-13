# 算能云渠道配置快速修复

## 错误信息
```json
{
    "error": {
        "code": "model_not_found",
        "message": "No available channel for model sophnet_generate"
    }
}
```

## 解决方案

### 方法 1: 在管理后台添加模型（推荐）

1. 登录管理后台
2. 进入**渠道管理**
3. 找到您创建的算能云渠道，点击**编辑**
4. 在**模型**字段中，添加：
   ```
   sophnet_generate
   ```
5. 点击**提交**保存

### 方法 2: 直接在数据库中添加（高级用户）

```sql
-- 假设您的渠道 ID 是 1，更新模型字段
UPDATE channels 
SET models = 'sophnet_generate' 
WHERE id = 1 AND type = 58;
```

### 方法 3: 通过 API 添加模型

```bash
curl -X PUT "https://your-domain.com/api/channel/1" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "models": "sophnet_generate"
  }'
```

## 配置说明

### 正确的配置方式

**渠道配置（后台设置）：**
- 模型字段: `sophnet_generate`

**API 请求（调用时）：**
```json
{
  "model": "Qwen-Image"
}
```

### 为什么需要这样配置？

- `sophnet_generate` 是系统内部的模型标识符，用于路由和计费
- `Qwen-Image` 等是上游真实的模型名称，会直接传递给算能云 API
- 系统会自动处理这两者之间的映射

## 验证配置

配置完成后，重新发送请求：

```bash
curl -X POST "https://your-domain.com/sophnet/submit/generate" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "Qwen-Image",
    "input": {
      "prompt": "测试图片"
    }
  }'
```

应该返回：
```json
{
  "code": "success",
  "message": "success",
  "data": "task_xxxxxxxx"
}
```

## 常见问题

### Q: 为什么不能直接使用 Qwen-Image 作为模型名？
A: 系统使用 `平台_action` 的命名规则（如 `sophnet_generate`）来标识不同平台的不同操作，这样可以：
- 统一管理多个平台的模型
- 分别配置每个操作的价格
- 更灵活的路由控制

### Q: 我可以同时支持多个上游模型吗？
A: 可以。只需在渠道中配置 `sophnet_generate`，然后在请求时通过 `model` 字段指定具体的上游模型（`Qwen-Image`、`Z-Image-Turbo` 等）。

### Q: 如何查看我的渠道是否配置正确？
A: 
1. 进入管理后台 -> 渠道管理
2. 查看算能云渠道的"模型"列，应显示 `sophnet_generate`
3. 确保渠道状态为"启用"

## 完整配置检查清单

- [ ] 渠道类型选择了"算能"（值为 58）
- [ ] Base URL 填写了 `https://www.sophnet.com` 或留空
- [ ] 密钥填写了有效的算能云 API Key
- [ ] 模型字段包含 `sophnet_generate`
- [ ] 渠道状态为"启用"
- [ ] 分组设置正确（如 `default`）

配置完成后，即可正常使用！
