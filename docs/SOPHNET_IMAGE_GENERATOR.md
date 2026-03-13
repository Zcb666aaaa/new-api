# 算能云异步图片生成接口文档

## 概述

算能云（Sophnet）异步图片生成功能已集成到 new-api 中，支持通过 `/imagegenerator/task` 路径提交和查询图片生成任务。

## 支持的模型

- `Qwen-Image` - 通义千问图像生成（开源）
- `Qwen-Image-Plus` - 通义千问图像生成Plus版
- `Qwen-Image-Edit-2509` - 通义千问图像编辑（开源，图生图）
- `Z-Image-Turbo` - Z-Image高速版（开源）
- `Wan2.6-T2I` - 万相2.6文生图

## API 接口

### 1. 创建图片生成任务

**请求方法**: `POST`

**请求路径**: `/imagegenerator/task`

**请求头**:
```
Authorization: Bearer {your_api_key}
Content-Type: application/json
```

**请求体参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| model | string | 是 | 模型名称，见支持的模型列表 |
| input | object | 是 | 输入对象 |
| input.prompt | string | 是 | 正向提示词 |
| input.images | array(string) | 否 | 输入图像的URL或Base64编码（Qwen-Image-Edit-2509必填） |
| input.negative_prompt | string | 否 | 反向提示词 |
| parameters | object | 否 | 生成参数 |
| parameters.size | string | 否 | 生成图像的分辨率 |
| parameters.seed | int | 否 | 图片生成的种子值 |
| parameters.prompt_extend | bool | 否 | 是否开启prompt智能改写，默认true |
| parameters.watermark | bool | 否 | 是否添加水印，默认false |
| parameters.save_to_jpeg | bool | 否 | 是否输出jpg格式，默认false |

**请求示例**:

```bash
curl -X POST "http://your-api-domain/imagegenerator/task" \
  -H "Authorization: Bearer sk-xxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "Qwen-Image",
    "input": {
      "prompt": "一只可爱的猫咪在弹钢琴",
      "negative_prompt": "模糊，低质量"
    },
    "parameters": {
      "size": "1328*1328",
      "prompt_extend": true,
      "watermark": false
    }
  }'
```

**响应示例**:

```json
{
  "code": "success",
  "message": "success",
  "data": "task_abc123xyz"
}
```

### 2. 查询图片生成任务

**请求方法**: `GET`

**请求路径**: `/imagegenerator/task/{taskId}`

**请求头**:
```
Authorization: Bearer {your_api_key}
```

**路径参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| taskId | string | 是 | 任务ID（创建任务时返回的ID） |

**请求示例**:

```bash
curl -X GET "http://your-api-domain/imagegenerator/task/task_abc123xyz" \
  -H "Authorization: Bearer sk-xxxx"
```

**响应示例（任务进行中）**:

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

**响应示例（任务成功）**:

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

**响应示例（任务失败）**:

```json
{
  "code": "success",
  "message": "success",
  "data": {
    "task_id": "task_abc123xyz",
    "status": "failure",
    "progress": "100%",
    "reason": "Invalid prompt"
  }
}
```

## 任务状态说明

| 状态 | 说明 |
|------|------|
| queued | 任务已排队，等待处理 |
| in_progress | 任务正在执行中 |
| success | 任务执行成功 |
| failure | 任务执行失败 |
| unknown | 任务状态未知 |

## 渠道配置

### 1. 创建算能云渠道

在管理后台创建渠道时：

- **渠道类型**: 选择 `Sophnet`
- **Base URL**: `https://www.sophnet.com`
- **API Key**: 填入你的算能云 API Key
- **模型**: 添加以下实际的模型名称（重要：必须使用实际模型名称，不是 sophnet_generate）
  - `Qwen-Image`
  - `Qwen-Image-Plus`
  - `Qwen-Image-Edit-2509`
  - `Z-Image-Turbo`
  - `Wan2.6-T2I`

**注意**: 在请求时，`model` 参数必须使用上述实际的模型名称（如 `Qwen-Image`），系统会自动路由到正确的算能云渠道。

### 2. 模型映射（可选）

如果需要将客户端请求的模型名称映射到算能云的实际模型名称，可以在渠道配置中设置模型映射。

例如：
```
my-image-model -> Qwen-Image
my-image-edit -> Qwen-Image-Edit-2509
```

## 计费说明

- 异步图片生成任务按照模型配置的价格计费
- 任务提交时预扣费，任务完成后根据实际使用情况结算
- 如果任务失败，会自动退款

## 注意事项

1. **图编辑模型**: `Qwen-Image-Edit-2509` 模型必须提供 `input.images` 参数
2. **图片格式**: 输入图片支持 URL 或 Base64 编码格式
3. **任务轮询**: 建议每隔 2-5 秒查询一次任务状态，避免频繁请求
4. **任务有效期**: 任务结果会保留一段时间，建议及时获取生成的图片

## 兼容性说明

除了 `/imagegenerator/task` 路径外，还支持以下路径（兼容旧版本）：

- `/sophnet/submit/generate` - 提交任务
- `/sophnet/fetch/{taskId}` - 查询任务

推荐使用 `/imagegenerator/task` 路径，更符合算能云的官方 API 规范。

## 错误处理

常见错误码：

| 错误码 | 说明 | 解决方法 |
|--------|------|----------|
| invalid_request | 请求参数无效 | 检查请求参数是否完整且格式正确 |
| model_not_found | 模型不存在 | 检查模型名称是否正确 |
| insufficient_quota | 余额不足 | 充值或检查账户余额 |
| task_not_found | 任务不存在 | 检查任务ID是否正确 |
| channel_not_found | 渠道不存在 | 检查渠道配置是否正确 |

## 示例代码

### Python 示例

```python
import requests
import time

# 配置
API_BASE = "http://your-api-domain"
API_KEY = "sk-xxxx"

headers = {
    "Authorization": f"Bearer {API_KEY}",
    "Content-Type": "application/json"
}

# 1. 创建任务
create_payload = {
    "model": "Qwen-Image",
    "input": {
        "prompt": "一只可爱的猫咪在弹钢琴",
        "negative_prompt": "模糊，低质量"
    },
    "parameters": {
        "size": "1328*1328",
        "prompt_extend": True
    }
}

response = requests.post(
    f"{API_BASE}/imagegenerator/task",
    headers=headers,
    json=create_payload
)

result = response.json()
task_id = result["data"]
print(f"任务已创建: {task_id}")

# 2. 轮询查询任务状态
while True:
    response = requests.get(
        f"{API_BASE}/imagegenerator/task/{task_id}",
        headers=headers
    )
    
    result = response.json()
    task_data = result["data"]
    status = task_data["status"]
    progress = task_data.get("progress", "0%")
    
    print(f"任务状态: {status}, 进度: {progress}")
    
    if status == "success":
        image_url = task_data["url"]
        print(f"图片生成成功: {image_url}")
        break
    elif status == "failure":
        reason = task_data.get("reason", "未知错误")
        print(f"任务失败: {reason}")
        break
    
    time.sleep(3)  # 等待3秒后再次查询
```

### JavaScript 示例

```javascript
const API_BASE = "http://your-api-domain";
const API_KEY = "sk-xxxx";

const headers = {
  "Authorization": `Bearer ${API_KEY}`,
  "Content-Type": "application/json"
};

// 1. 创建任务
async function createTask() {
  const response = await fetch(`${API_BASE}/imagegenerator/task`, {
    method: "POST",
    headers: headers,
    body: JSON.stringify({
      model: "Qwen-Image",
      input: {
        prompt: "一只可爱的猫咪在弹钢琴",
        negative_prompt: "模糊，低质量"
      },
      parameters: {
        size: "1328*1328",
        prompt_extend: true
      }
    })
  });
  
  const result = await response.json();
  return result.data;
}

// 2. 查询任务状态
async function fetchTask(taskId) {
  const response = await fetch(`${API_BASE}/imagegenerator/task/${taskId}`, {
    method: "GET",
    headers: headers
  });
  
  const result = await response.json();
  return result.data;
}

// 3. 轮询直到完成
async function waitForTask(taskId) {
  while (true) {
    const taskData = await fetchTask(taskId);
    console.log(`任务状态: ${taskData.status}, 进度: ${taskData.progress}`);
    
    if (taskData.status === "success") {
      console.log(`图片生成成功: ${taskData.url}`);
      return taskData.url;
    } else if (taskData.status === "failure") {
      console.error(`任务失败: ${taskData.reason}`);
      throw new Error(taskData.reason);
    }
    
    await new Promise(resolve => setTimeout(resolve, 3000));
  }
}

// 使用示例
(async () => {
  try {
    const taskId = await createTask();
    console.log(`任务已创建: ${taskId}`);
    
    const imageUrl = await waitForTask(taskId);
    console.log(`最终图片URL: ${imageUrl}`);
  } catch (error) {
    console.error("错误:", error);
  }
})();
```

## 技术支持

如有问题，请参考：
- [算能云官方文档](https://www.sophnet.com/docs)
- [new-api 项目文档](https://github.com/QuantumNous/new-api)
