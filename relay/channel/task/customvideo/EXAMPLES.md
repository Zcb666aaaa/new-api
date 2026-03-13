# 自定义视频渠道使用示例

## 快速开始

### 1. 配置渠道

在管理后台添加自定义视频渠道：

```yaml
渠道名称: 我的视频服务
渠道类型: CustomVideo
Base URL: https://api.myvideo.com
API Key: sk-xxxxxxxxxxxxx
模型列表: custom-video-model
状态: 启用
```

### 2. 配置模型价格

```yaml
模型名称: custom-video-model
基础价格: 0.1 (每次任务)
其他倍率:
  - seconds: 0.02 (每秒额外费用)
```

### 3. 调用API

#### Python 示例

```python
import requests
import time

# 创建视频任务
response = requests.post(
    'https://your-api.com/v1/videos/create',
    headers={
        'Authorization': 'Bearer YOUR_TOKEN',
        'Content-Type': 'application/json'
    },
    json={
        'model': 'custom-video-model',
        'prompt': '一只可爱的猫咪在草地上玩耍',
        'size': '1920x1080',
        'duration': 5
    }
)

task = response.json()
task_id = task['id']
print(f"任务已创建: {task_id}")

# 轮询任务状态
while True:
    response = requests.get(
        f'https://your-api.com/v1/videos/{task_id}',
        headers={'Authorization': 'Bearer YOUR_TOKEN'}
    )
    
    task = response.json()
    status = task['status']
    print(f"任务状态: {status}")
    
    if status == 'completed':
        video_url = task.get('metadata', {}).get('url')
        print(f"视频已生成: {video_url}")
        break
    elif status == 'failed':
        error = task.get('error', {})
        print(f"任务失败: {error.get('message')}")
        break
    
    time.sleep(5)
```

#### JavaScript 示例

```javascript
const axios = require('axios');

async function createVideo() {
    // 创建视频任务
    const createResponse = await axios.post(
        'https://your-api.com/v1/videos/create',
        {
            model: 'custom-video-model',
            prompt: '一只可爱的猫咪在草地上玩耍',
            size: '1920x1080',
            duration: 5
        },
        {
            headers: {
                'Authorization': 'Bearer YOUR_TOKEN',
                'Content-Type': 'application/json'
            }
        }
    );

    const taskId = createResponse.data.id;
    console.log(`任务已创建: ${taskId}`);

    // 轮询任务状态
    while (true) {
        const statusResponse = await axios.get(
            `https://your-api.com/v1/videos/${taskId}`,
            {
                headers: {
                    'Authorization': 'Bearer YOUR_TOKEN'
                }
            }
        );

        const task = statusResponse.data;
        const status = task.status;
        console.log(`任务状态: ${status}`);

        if (status === 'completed') {
            const videoUrl = task.metadata?.url;
            console.log(`视频已生成: ${videoUrl}`);
            break;
        } else if (status === 'failed') {
            const error = task.error;
            console.log(`任务失败: ${error?.message}`);
            break;
        }

        await new Promise(resolve => setTimeout(resolve, 5000));
    }
}

createVideo();
```

#### cURL 示例

```bash
# 创建视频任务
TASK_ID=$(curl -X POST https://your-api.com/v1/videos/create \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "custom-video-model",
    "prompt": "一只可爱的猫咪在草地上玩耍",
    "size": "1920x1080",
    "duration": 5
  }' | jq -r '.id')

echo "任务ID: $TASK_ID"

# 查询任务状态
while true; do
  STATUS=$(curl -s https://your-api.com/v1/videos/$TASK_ID \
    -H "Authorization: Bearer YOUR_TOKEN" | jq -r '.status')
  
  echo "任务状态: $STATUS"
  
  if [ "$STATUS" = "completed" ]; then
    VIDEO_URL=$(curl -s https://your-api.com/v1/videos/$TASK_ID \
      -H "Authorization: Bearer YOUR_TOKEN" | jq -r '.metadata.url')
    echo "视频已生成: $VIDEO_URL"
    break
  elif [ "$STATUS" = "failed" ]; then
    echo "任务失败"
    break
  fi
  
  sleep 5
done
```

## 上游服务实现示例

### Flask 示例（Python）

```python
from flask import Flask, request, jsonify
import uuid
import time

app = Flask(__name__)

# 模拟任务存储
tasks = {}

@app.route('/v1/videos/create', methods=['POST'])
def create_video():
    # 验证 API Key
    auth_header = request.headers.get('Authorization')
    if not auth_header or not auth_header.startswith('Bearer '):
        return jsonify({'error': 'Unauthorized'}), 401
    
    api_key = auth_header.replace('Bearer ', '')
    if api_key != 'sk-your-secret-key':
        return jsonify({'error': 'Invalid API key'}), 401
    
    # 解析请求
    data = request.json
    model = data.get('model')
    prompt = data.get('prompt')
    size = data.get('size', '1920x1080')
    duration = data.get('duration', 5)
    
    # 创建任务
    task_id = str(uuid.uuid4())
    tasks[task_id] = {
        'id': task_id,
        'status': 'queued',
        'model': model,
        'prompt': prompt,
        'size': size,
        'duration': duration,
        'created_at': int(time.time()),
        'video_url': None,
        'progress': '0%'
    }
    
    # 模拟异步处理（实际应该使用队列）
    import threading
    threading.Thread(target=process_video, args=(task_id,)).start()
    
    return jsonify({
        'id': task_id,
        'status': 'queued',
        'created_at': tasks[task_id]['created_at']
    })

@app.route('/v1/videos/<task_id>', methods=['GET'])
def get_video(task_id):
    # 验证 API Key
    auth_header = request.headers.get('Authorization')
    if not auth_header or not auth_header.startswith('Bearer '):
        return jsonify({'error': 'Unauthorized'}), 401
    
    # 查询任务
    task = tasks.get(task_id)
    if not task:
        return jsonify({'error': 'Task not found'}), 404
    
    return jsonify(task)

def process_video(task_id):
    """模拟视频生成过程"""
    task = tasks[task_id]
    
    # 模拟处理中
    time.sleep(2)
    task['status'] = 'processing'
    task['progress'] = '30%'
    
    time.sleep(3)
    task['progress'] = '60%'
    
    time.sleep(2)
    task['progress'] = '90%'
    
    # 模拟完成
    time.sleep(1)
    task['status'] = 'completed'
    task['progress'] = '100%'
    task['video_url'] = f'https://cdn.example.com/videos/{task_id}.mp4'

if __name__ == '__main__':
    app.run(port=8000)
```

### Express 示例（Node.js）

```javascript
const express = require('express');
const { v4: uuidv4 } = require('uuid');

const app = express();
app.use(express.json());

// 模拟任务存储
const tasks = new Map();

// 创建视频任务
app.post('/v1/videos/create', (req, res) => {
    // 验证 API Key
    const authHeader = req.headers.authorization;
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
        return res.status(401).json({ error: 'Unauthorized' });
    }
    
    const apiKey = authHeader.replace('Bearer ', '');
    if (apiKey !== 'sk-your-secret-key') {
        return res.status(401).json({ error: 'Invalid API key' });
    }
    
    // 解析请求
    const { model, prompt, size = '1920x1080', duration = 5 } = req.body;
    
    // 创建任务
    const taskId = uuidv4();
    const task = {
        id: taskId,
        status: 'queued',
        model,
        prompt,
        size,
        duration,
        created_at: Math.floor(Date.now() / 1000),
        video_url: null,
        progress: '0%'
    };
    
    tasks.set(taskId, task);
    
    // 模拟异步处理
    processVideo(taskId);
    
    res.json({
        id: taskId,
        status: 'queued',
        created_at: task.created_at
    });
});

// 查询视频任务
app.get('/v1/videos/:taskId', (req, res) => {
    // 验证 API Key
    const authHeader = req.headers.authorization;
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
        return res.status(401).json({ error: 'Unauthorized' });
    }
    
    // 查询任务
    const task = tasks.get(req.params.taskId);
    if (!task) {
        return res.status(404).json({ error: 'Task not found' });
    }
    
    res.json(task);
});

// 模拟视频生成过程
async function processVideo(taskId) {
    const task = tasks.get(taskId);
    
    // 模拟处理中
    await sleep(2000);
    task.status = 'processing';
    task.progress = '30%';
    
    await sleep(3000);
    task.progress = '60%';
    
    await sleep(2000);
    task.progress = '90%';
    
    // 模拟完成
    await sleep(1000);
    task.status = 'completed';
    task.progress = '100%';
    task.video_url = `https://cdn.example.com/videos/${taskId}.mp4`;
}

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

app.listen(8000, () => {
    console.log('Server running on port 8000');
});
```

## 高级用法

### 自定义元数据

```json
{
  "model": "custom-video-model",
  "prompt": "视频描述",
  "size": "1920x1080",
  "duration": 5,
  "metadata": {
    "style": "anime",
    "fps": 30,
    "quality": "high",
    "watermark": false
  }
}
```

### 批量创建任务

```python
import asyncio
import aiohttp

async def create_video_batch(prompts):
    async with aiohttp.ClientSession() as session:
        tasks = []
        for prompt in prompts:
            task = create_video_task(session, prompt)
            tasks.append(task)
        
        results = await asyncio.gather(*tasks)
        return results

async def create_video_task(session, prompt):
    async with session.post(
        'https://your-api.com/v1/videos/create',
        headers={'Authorization': 'Bearer YOUR_TOKEN'},
        json={
            'model': 'custom-video-model',
            'prompt': prompt,
            'duration': 5
        }
    ) as response:
        return await response.json()

# 使用
prompts = [
    '一只猫在玩耍',
    '一只狗在奔跑',
    '一只鸟在飞翔'
]

results = asyncio.run(create_video_batch(prompts))
```

## 故障排查

### 调试模式

启用详细日志：

```bash
export LOG_LEVEL=debug
./new-api
```

### 测试上游服务

```bash
# 测试创建接口
curl -X POST https://api.myvideo.com/v1/videos/create \
  -H "Authorization: Bearer sk-xxxxx" \
  -H "Content-Type: application/json" \
  -d '{"model":"test","prompt":"test"}' \
  -v

# 测试查询接口
curl -X GET https://api.myvideo.com/v1/videos/test_id \
  -H "Authorization: Bearer sk-xxxxx" \
  -v
```

### 常见错误

1. **401 Unauthorized**
   - 检查 API Key 是否正确
   - 确认 Authorization 头格式正确

2. **404 Not Found**
   - 检查 Base URL 配置
   - 确认上游服务路径正确

3. **任务一直处于 queued 状态**
   - 检查任务轮询是否正常
   - 查看上游服务是否正常处理任务

## 性能优化

### 并发控制

在渠道配置中设置：
- **最大并发数**：限制同时处理的任务数
- **速率限制**：控制请求频率

### 缓存策略

- 启用结果缓存
- 配置合理的过期时间

### 监控告警

- 监控任务成功率
- 监控平均处理时间
- 设置异常告警
