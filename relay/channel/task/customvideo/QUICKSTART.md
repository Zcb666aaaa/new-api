# 自定义视频渠道 - 快速入门

## 5分钟快速开始

### 步骤 1: 添加渠道

在管理后台添加新渠道：

1. 进入 **渠道管理** 页面
2. 点击 **添加渠道**
3. 填写以下信息：
   - **渠道类型**: `CustomVideo`
   - **渠道名称**: `我的视频服务`
   - **Base URL**: `https://api.example.com` (你的上游服务地址)
   - **API Key**: `sk-xxxxxxxxxxxxx` (你的API密钥)
   - **模型列表**: `custom-video-model`
   - **状态**: `启用`
4. 点击 **保存**

### 步骤 2: 配置价格

1. 进入 **模型价格** 页面
2. 添加新模型价格：
   - **模型名称**: `custom-video-model`
   - **基础价格**: `0.1` (每次任务基础费用)
   - **其他倍率**: 
     - `seconds`: `0.02` (每秒额外费用)
3. 点击 **保存**

### 步骤 3: 测试调用

使用 cURL 测试：

```bash
# 创建视频任务
curl -X POST https://your-api.com/v1/videos/create \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "custom-video-model",
    "prompt": "一只可爱的猫咪在草地上玩耍",
    "size": "1920x1080",
    "duration": 5
  }'

# 响应示例
{
  "id": "task_abc123",
  "task_id": "task_abc123",
  "model": "custom-video-model",
  "status": "queued",
  "created_at": 1234567890
}

# 查询任务状态
curl -X GET https://your-api.com/v1/videos/task_abc123 \
  -H "Authorization: Bearer YOUR_TOKEN"

# 响应示例
{
  "id": "task_abc123",
  "status": "completed",
  "progress": "100%",
  "metadata": {
    "url": "https://cdn.example.com/video.mp4"
  },
  "created_at": 1234567890,
  "completed_at": 1234567900
}
```

## 上游服务要求

你的上游服务需要实现两个接口：

### 1. 创建任务接口

```
POST {baseURL}/v1/videos/create
Authorization: Bearer {api_key}
Content-Type: application/json

请求体:
{
  "model": "string",
  "prompt": "string",
  "size": "string",
  "duration": integer,
  "metadata": object
}

响应体:
{
  "id": "string (必需)",
  "status": "string (必需)",
  "created_at": integer
}
```

### 2. 查询任务接口

```
GET {baseURL}/v1/videos/{id}
Authorization: Bearer {api_key}

响应体:
{
  "id": "string (必需)",
  "status": "string (必需)",
  "video_url": "string (可选)",
  "progress": "string (可选)",
  "message": "string (可选)"
}
```

### 状态值

上游服务应返回以下状态之一：

- `queued` 或 `pending` - 排队中
- `processing`, `running`, 或 `in_progress` - 处理中
- `completed`, `succeeded`, 或 `success` - 已完成
- `failed`, `error`, `canceled`, 或 `cancelled` - 失败

## 最简单的上游服务示例

使用 Python Flask 实现：

```python
from flask import Flask, request, jsonify
import uuid
import time
import threading

app = Flask(__name__)
tasks = {}

@app.route('/v1/videos/create', methods=['POST'])
def create_video():
    # 验证 API Key
    if request.headers.get('Authorization') != 'Bearer sk-your-secret-key':
        return jsonify({'error': 'Unauthorized'}), 401
    
    # 创建任务
    task_id = str(uuid.uuid4())
    tasks[task_id] = {
        'id': task_id,
        'status': 'queued',
        'created_at': int(time.time())
    }
    
    # 异步处理
    threading.Thread(target=process_video, args=(task_id,)).start()
    
    return jsonify(tasks[task_id])

@app.route('/v1/videos/<task_id>', methods=['GET'])
def get_video(task_id):
    # 验证 API Key
    if request.headers.get('Authorization') != 'Bearer sk-your-secret-key':
        return jsonify({'error': 'Unauthorized'}), 401
    
    task = tasks.get(task_id)
    if not task:
        return jsonify({'error': 'Task not found'}), 404
    
    return jsonify(task)

def process_video(task_id):
    time.sleep(5)  # 模拟处理
    tasks[task_id]['status'] = 'completed'
    tasks[task_id]['video_url'] = f'https://cdn.example.com/{task_id}.mp4'

if __name__ == '__main__':
    app.run(port=8000)
```

运行服务：
```bash
pip install flask
python app.py
```

然后在 new-api 中配置 Base URL 为 `http://localhost:8000`

## 计费说明

### 计费公式

```
总费用 = 基础价格 × seconds倍率 × 时长(秒)
```

### 示例

如果配置：
- 基础价格: 0.1 元
- seconds倍率: 0.02
- 用户请求时长: 5秒

则费用计算：
```
总费用 = 0.1 × 0.02 × 5 = 0.01 元
```

### 自定义计费

你可以添加更多倍率参数：

```yaml
模型价格配置:
  基础价格: 0.1
  其他倍率:
    seconds: 0.02      # 时长倍率
    resolution_1080p: 2.0  # 1080p分辨率倍率
    resolution_720p: 1.5   # 720p分辨率倍率
```

## 常见问题

### Q: 如何测试渠道是否配置正确？

A: 使用 cURL 直接测试上游服务：

```bash
# 测试创建接口
curl -X POST https://api.example.com/v1/videos/create \
  -H "Authorization: Bearer sk-xxxxx" \
  -H "Content-Type: application/json" \
  -d '{"model":"test","prompt":"test"}' \
  -v

# 测试查询接口
curl -X GET https://api.example.com/v1/videos/test_id \
  -H "Authorization: Bearer sk-xxxxx" \
  -v
```

### Q: 任务一直处于 queued 状态怎么办？

A: 检查以下几点：
1. 上游服务是否正常处理任务
2. 查询接口是否返回正确的状态
3. 查看 new-api 日志是否有错误

### Q: 如何查看详细日志？

A: 启用调试模式：

```bash
export LOG_LEVEL=debug
./new-api
```

### Q: 支持哪些视频格式？

A: 自定义视频渠道不限制视频格式，由上游服务决定。建议使用 MP4 格式以获得最佳兼容性。

### Q: 如何处理大文件视频？

A: 建议上游服务返回 CDN URL，而不是直接返回视频内容。new-api 会存储 URL 并提供代理访问。

## 下一步

- 📖 阅读 [完整文档](README.md)
- 💡 查看 [使用示例](EXAMPLES.md)
- 🔧 了解 [实现细节](IMPLEMENTATION.md)

## 获取帮助

如有问题，请：
1. 查看日志文件
2. 检查上游服务状态
3. 参考文档和示例
4. 提交 Issue

---

**提示**: 确保上游服务使用 HTTPS 并妥善保管 API Key！
