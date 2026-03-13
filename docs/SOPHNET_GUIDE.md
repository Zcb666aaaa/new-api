# 算能云渠道注册与调用完整指南

## 一、在管理后台注册算能云渠道

### 1. 登录管理后台

访问您的 new-api 管理后台，使用管理员账号登录。

### 2. 添加渠道

1. 进入**渠道管理**页面
2. 点击**添加渠道**按钮
3. 填写以下信息：

| 字段 | 填写内容 | 说明 |
|------|---------|------|
| **类型** | 选择 "算能" | 渠道类型下拉选择 |
| **名称** | 算能云图片生成 | 自定义渠道名称 |
| **分组** | default | 或选择您的用户分组 |
| **Base URL** | `https://www.sophnet.com` | 留空使用默认值 |
| **密钥** | 您的算能云 API Key | 从算能云官网获取 |
| **模型** | 选择支持的模型（见下方） | 支持多选 |
| **优先级** | 0 | 数字越大优先级越高 |
| **权重** | 1 | 负载均衡权重 |

### 3. 支持的模型

**重要**: 在渠道的"模型"字段中，需要添加以下模型名称：

```
sophnet_generate
```

这是系统内部使用的模型标识。实际调用时，您可以在请求体的 `model` 字段中指定上游真实的模型名称：

**上游支持的模型（在请求体中使用）：**
- `Qwen-Image` - 通义千问图像生成（文生图）
- `Qwen-Image-Plus` - 通义千问图像生成 Plus 版（文生图）
- `Qwen-Image-Edit-2509` - 通义千问图像编辑（图生图，需要输入图片）
- `Z-Image-Turbo` - Z-Image 高速版（文生图）
- `Wan2.6-T2I` - 万相 2.6 文生图

**配置示例：**
- 渠道模型配置: `sophnet_generate`
- 请求体中的 model: `Qwen-Image`（或其他上游模型）

### 4. 获取算能云 API Key

1. 访问算能云官网: https://www.sophnet.com
2. 注册/登录账号
3. 进入控制台
4. 创建 API Key
5. 复制 API Key（格式可能为 `sk-xxxxxx`）

### 5. 保存渠道

点击**提交**按钮保存渠道配置。

---

## 二、调用算能云接口

### 方式 1: 使用 API 直接调用

#### 提交图片生成任务

**请求示例（文生图）：**

```bash
curl -X POST "https://your-new-api-domain.com/sophnet/submit/generate" \
  -H "Authorization: Bearer sk-your-new-api-token" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "Qwen-Image",
    "input": {
      "prompt": "一只可爱的小猫在花园里玩耍，阳光明媚，4K高清",
      "negative_prompt": "模糊,低质量,变形"
    },
    "parameters": {
      "size": "1328*1328",
      "prompt_extend": true,
      "watermark": false
    }
  }'
```

**响应示例：**

```json
{
  "code": "success",
  "message": "success",
  "data": "task_abc123xyz456789"
}
```

#### 查询任务状态

**方式 A - GET 请求：**

```bash
curl -X GET "https://your-new-api-domain.com/sophnet/fetch/task_abc123xyz456789" \
  -H "Authorization: Bearer sk-your-new-api-token"
```

**方式 B - POST 请求：**

```bash
curl -X POST "https://your-new-api-domain.com/sophnet/fetch" \
  -H "Authorization: Bearer sk-your-new-api-token" \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": "task_abc123xyz456789"
  }'
```

**响应示例（进行中）：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 12345,
    "task_id": "task_abc123xyz456789",
    "status": "IN_PROGRESS",
    "progress": "50%",
    "created_at": 1709539200
  }
}
```

**响应示例（成功）：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 12345,
    "task_id": "task_abc123xyz456789",
    "status": "SUCCESS",
    "progress": "100%",
    "result_url": "https://example.com/generated-image.jpg",
    "created_at": 1709539200,
    "finish_time": 1709539230
  }
}
```

### 方式 2: 使用 Python SDK

```python
import requests
import time

# 配置
API_BASE = "https://your-new-api-domain.com"
API_KEY = "sk-your-new-api-token"
headers = {
    "Authorization": f"Bearer {API_KEY}",
    "Content-Type": "application/json"
}

# 1. 提交任务
def submit_task(prompt, model="Qwen-Image"):
    url = f"{API_BASE}/sophnet/submit/generate"
    data = {
        "model": model,
        "input": {
            "prompt": prompt,
            "negative_prompt": "模糊,低质量"
        },
        "parameters": {
            "size": "1328*1328",
            "prompt_extend": True,
            "watermark": False
        }
    }
    
    response = requests.post(url, json=data, headers=headers)
    result = response.json()
    
    if result.get("code") == "success":
        return result["data"]  # 返回 task_id
    else:
        raise Exception(f"提交失败: {result}")

# 2. 查询任务
def query_task(task_id):
    url = f"{API_BASE}/sophnet/fetch/{task_id}"
    response = requests.get(url, headers=headers)
    return response.json()

# 3. 等待任务完成
def wait_for_task(task_id, timeout=300):
    start_time = time.time()
    
    while time.time() - start_time < timeout:
        result = query_task(task_id)
        data = result.get("data", {})
        status = data.get("status")
        
        print(f"任务状态: {status}, 进度: {data.get('progress')}")
        
        if status == "SUCCESS":
            return data.get("result_url")
        elif status == "FAILURE":
            raise Exception(f"任务失败: {data.get('fail_reason')}")
        
        time.sleep(5)  # 每 5 秒查询一次
    
    raise Exception("任务超时")

# 使用示例
if __name__ == "__main__":
    # 提交任务
    prompt = "一只可爱的小猫在花园里玩耍，阳光明媚，4K高清"
    task_id = submit_task(prompt)
    print(f"任务已提交，ID: {task_id}")
    
    # 等待完成
    image_url = wait_for_task(task_id)
    print(f"图片生成成功: {image_url}")
```

### 方式 3: 使用 Node.js

```javascript
const axios = require('axios');

const API_BASE = 'https://your-new-api-domain.com';
const API_KEY = 'sk-your-new-api-token';

// 提交任务
async function submitTask(prompt, model = 'Qwen-Image') {
  const response = await axios.post(
    `${API_BASE}/sophnet/submit/generate`,
    {
      model: model,
      input: {
        prompt: prompt,
        negative_prompt: '模糊,低质量'
      },
      parameters: {
        size: '1328*1328',
        prompt_extend: true,
        watermark: false
      }
    },
    {
      headers: {
        'Authorization': `Bearer ${API_KEY}`,
        'Content-Type': 'application/json'
      }
    }
  );
  
  return response.data.data; // 返回 task_id
}

// 查询任务
async function queryTask(taskId) {
  const response = await axios.get(
    `${API_BASE}/sophnet/fetch/${taskId}`,
    {
      headers: {
        'Authorization': `Bearer ${API_KEY}`
      }
    }
  );
  
  return response.data;
}

// 等待任务完成
async function waitForTask(taskId, timeout = 300000) {
  const startTime = Date.now();
  
  while (Date.now() - startTime < timeout) {
    const result = await queryTask(taskId);
    const data = result.data;
    const status = data.status;
    
    console.log(`任务状态: ${status}, 进度: ${data.progress}`);
    
    if (status === 'SUCCESS') {
      return data.result_url;
    } else if (status === 'FAILURE') {
      throw new Error(`任务失败: ${data.fail_reason}`);
    }
    
    await new Promise(resolve => setTimeout(resolve, 5000)); // 等待 5 秒
  }
  
  throw new Error('任务超时');
}

// 使用示例
(async () => {
  try {
    const prompt = '一只可爱的小猫在花园里玩耍，阳光明媚，4K高清';
    const taskId = await submitTask(prompt);
    console.log(`任务已提交，ID: ${taskId}`);
    
    const imageUrl = await waitForTask(taskId);
    console.log(`图片生成成功: ${imageUrl}`);
  } catch (error) {
    console.error('错误:', error.message);
  }
})();
```

---

## 三、图生图调用示例

### 使用 `Qwen-Image-Edit-2509` 模型

```bash
curl -X POST "https://your-new-api-domain.com/sophnet/submit/generate" \
  -H "Authorization: Bearer sk-your-new-api-token" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "Qwen-Image-Edit-2509",
    "input": {
      "prompt": "将图片改成油画风格，梵高风格",
      "images": [
        "https://example.com/input-image.jpg"
      ]
    },
    "parameters": {
      "size": "1280*1280",
      "prompt_extend": true
    }
  }'
```

**支持的图片格式：**
- URL: `https://example.com/image.jpg`
- Base64: `data:image/jpeg;base64,/9j/4AAQ...`

---

## 四、参数说明

### Input 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| prompt | string | 是 | 正向提示词 |
| images | array | 否* | 输入图片（图生图时必填） |
| negative_prompt | string | 否 | 反向提示词 |

### Parameters 参数

| 参数 | 类型 | 必填 | 说明 | 默认值 |
|------|------|------|------|--------|
| size | string | 否 | 图片分辨率 | 1328*1328 |
| seed | int | 否 | 随机种子 | 随机 |
| prompt_extend | bool | 否 | 智能改写 prompt | true |
| watermark | bool | 否 | 是否添加水印 | false |
| save_to_jpeg | bool | 否 | 输出 jpg 格式 | false |

### 支持的分辨率

**Qwen-Image:**
- 1328*1328（默认）
- 1664*928
- 更多尺寸参见官方文档

**Z-Image-Turbo:**
- 1024*1024（默认）

**Qwen-Image-Edit-2509:**
- 1280*1280（默认，与输入图片宽高比一致）
- 1024*1024

---

## 五、常见问题

### 1. 如何查看任务列表？

访问管理后台的**任务管理**页面，可以查看所有任务的状态。

### 2. 任务失败如何处理？

任务失败会自动退款。查看任务详情中的 `fail_reason` 字段了解失败原因。

### 3. 是否支持批量生成？

目前需要逐个提交任务。可以使用脚本并发提交多个任务。

### 4. 图片保存在哪里？

生成的图片由算能云托管，返回的是图片 URL。建议下载后保存到自己的存储。

### 5. 计费方式？

- 提交任务时预扣费
- 任务完成后根据实际消耗结算
- 任务失败自动退款

---

## 六、注意事项

1. **API Key 安全**: 不要在客户端代码中暴露 API Key
2. **并发控制**: 建议控制并发数，避免触发限流
3. **超时设置**: 图片生成通常需要 10-60 秒
4. **图片下载**: 及时下载生成的图片，上游可能有过期时间
5. **错误重试**: 遇到网络错误建议重试，但避免频繁重试

---

## 七、技术支持

如有问题，请：
1. 查看管理后台的任务日志
2. 检查渠道是否启用
3. 确认 API Key 是否有效
4. 查看系统日志获取详细错误信息

祝您使用愉快！🎨
