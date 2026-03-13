# 算能云异步图片生成 - 故障排除

## 常见错误及解决方法

### 错误 1: `invalid_api_platform: sophnet`

**错误信息:**
```json
{
  "code": "invalid_api_platform",
  "message": "invalid api platform: sophnet",
  "data": null
}
```

**原因:** 系统无法识别平台标识

**解决方法:** 
- 确保使用正确的请求路径 `/imagegenerator/task`
- 检查代码是否正确设置了 `channel_type` 而不是 `platform`

---

### 错误 2: `model_not_found: sophnet_generate`

**错误信息:**
```json
{
  "error": {
    "code": "model_not_found",
    "message": "No available channel for model sophnet_generate under group default",
    "type": "new_api_error"
  }
}
```

**原因:** 请求中使用了错误的模型名称

**解决方法:**
1. **在请求中使用实际的模型名称**，而不是 `sophnet_generate`：
   ```json
   {
     "model": "Qwen-Image",  // ✅ 正确
     "input": {
       "prompt": "一只可爱的猫咪在弹钢琴"
     }
   }
   ```

2. **在渠道配置中添加实际的模型名称**：
   - 进入管理后台
   - 找到算能云渠道
   - 在模型列表中添加：
     - `Qwen-Image`
     - `Qwen-Image-Plus`
     - `Qwen-Image-Edit-2509`
     - `Z-Image-Turbo`
     - `Wan2.6-T2I`

---

### 错误 3: `model is required`

**错误信息:**
```json
{
  "code": "invalid_request",
  "message": "model is required"
}
```

**原因:** 请求体中缺少 `model` 参数

**解决方法:**
确保请求体包含 `model` 字段：
```json
{
  "model": "Qwen-Image",
  "input": {
    "prompt": "你的提示词"
  }
}
```

---

### 错误 4: `input.prompt is required`

**错误信息:**
```json
{
  "code": "invalid_request",
  "message": "input.prompt is required"
}
```

**原因:** 请求体中缺少 `input.prompt` 参数

**解决方法:**
确保请求体包含 `input.prompt` 字段：
```json
{
  "model": "Qwen-Image",
  "input": {
    "prompt": "一只可爱的猫咪在弹钢琴"
  }
}
```

---

### 错误 5: `input.images is required for Qwen-Image-Edit-2509`

**错误信息:**
```json
{
  "code": "invalid_request",
  "message": "input.images is required for Qwen-Image-Edit-2509"
}
```

**原因:** 图编辑模型需要输入参考图片

**解决方法:**
使用 `Qwen-Image-Edit-2509` 模型时，必须提供 `input.images`：
```json
{
  "model": "Qwen-Image-Edit-2509",
  "input": {
    "prompt": "将这张图片转换为油画风格",
    "images": [
      "https://example.com/image.jpg"
    ]
  }
}
```

---

### 错误 6: `channel_not_found`

**错误信息:**
```json
{
  "code": "channel_not_found",
  "message": "No available channel"
}
```

**原因:** 没有配置算能云渠道或渠道已禁用

**解决方法:**
1. 在管理后台创建算能云渠道
2. 确保渠道状态为"启用"
3. 检查渠道的模型列表是否包含你请求的模型

---

### 错误 7: `insufficient_quota`

**错误信息:**
```json
{
  "code": "insufficient_quota",
  "message": "Insufficient quota"
}
```

**原因:** 账户余额不足

**解决方法:**
1. 检查账户余额
2. 充值或调整配额
3. 检查模型的价格配置

---

### 错误 8: `task_not_found` 或 `task_not_exist`

**错误信息:**
```json
{
  "code": "task_not_exist",
  "message": "task_not_exist",
  "data": null
}
```

**原因**: 查询任务时无法找到任务记录

**可能的原因和解决方法:**

1. **任务 ID 不正确**
   - 确保使用创建任务时返回的完整 task_id
   - 检查是否有多余的空格或特殊字符

2. **URL 参数问题（已修复）**
   - 早期版本中，路由参数名称不匹配导致无法获取 task_id
   - 确保使用最新版本的代码

3. **用户权限问题**
   - 只能查询自己创建的任务
   - 确认使用的 API Key 与创建任务时相同

4. **任务已过期**
   - 任务结果可能有保留期限
   - 建议及时获取任务结果

**解决方法:**
```bash
# 确保 URL 格式正确
curl -X GET "http://your-domain/imagegenerator/task/task_abc123xyz" \
  -H "Authorization: Bearer sk-xxxx"

# 注意：task_id 前面不要有多余的斜杠
```

---

### 错误 9: `get_task_failed`

## 完整的正确示例

### 创建任务（文生图）

```bash
curl -X POST "http://your-domain/imagegenerator/task" \
  -H "Authorization: Bearer sk-xxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "Qwen-Image",
    "input": {
      "prompt": "一只可爱的猫咪在弹钢琴，高清，细节丰富",
      "negative_prompt": "模糊，低质量，变形"
    },
    "parameters": {
      "size": "1328*1328",
      "prompt_extend": true,
      "watermark": false
    }
  }'
```

### 创建任务（图生图）

```bash
curl -X POST "http://your-domain/imagegenerator/task" \
  -H "Authorization: Bearer sk-xxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "Qwen-Image-Edit-2509",
    "input": {
      "prompt": "将这张图片转换为油画风格",
      "images": [
        "https://example.com/input-image.jpg"
      ],
      "negative_prompt": "模糊，低质量"
    },
    "parameters": {
      "size": "1280*1280"
    }
  }'
```

### 查询任务

```bash
curl -X GET "http://your-domain/imagegenerator/task/task_abc123xyz" \
  -H "Authorization: Bearer sk-xxxx"
```

---

## 检查清单

在提交问题前，请检查：

- [ ] 使用的是正确的请求路径 `/imagegenerator/task`
- [ ] 请求体包含 `model` 字段，且使用实际的模型名称（如 `Qwen-Image`）
- [ ] 请求体包含 `input.prompt` 字段
- [ ] 如果使用图编辑模型，包含 `input.images` 字段
- [ ] 渠道已创建并启用
- [ ] 渠道的模型列表包含你请求的模型
- [ ] API Key 正确且有效
- [ ] 账户余额充足
- [ ] 请求头包含正确的 `Authorization` 和 `Content-Type`

---

## 调试技巧

### 1. 查看日志

检查 new-api 的日志文件，查找详细的错误信息：
```bash
tail -f logs/new-api.log
```

### 2. 测试渠道连接

在管理后台使用"测试"功能验证渠道配置是否正确。

注意：算能云渠道不支持标准的渠道测试，需要手动测试。

### 3. 验证请求格式

使用 JSON 验证工具确保请求体格式正确：
```bash
echo '{"model":"Qwen-Image","input":{"prompt":"test"}}' | jq .
```

### 4. 检查模型配置

确认渠道中配置的模型名称与请求中使用的模型名称完全一致（区分大小写）。

---

## 获取帮助

如果以上方法都无法解决问题，请提供以下信息：

1. 完整的错误信息
2. 请求的 curl 命令或代码
3. 渠道配置截图
4. 相关日志片段

---

## 相关文档

- [API 使用文档](SOPHNET_IMAGE_GENERATOR.md)
- [快速开始指南](SOPHNET_IMAGE_QUICKSTART.md)
- [集成说明](SOPHNET_IMAGE_INTEGRATION.md)
