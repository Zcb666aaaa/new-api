# 算能云文本生成集成说明

## 概述

算能云的文本生成接口完全兼容 OpenAI 的 `/v1/chat/completions` 标准格式，因此可以直接使用标准的 OpenAI 客户端进行调用。

## API 信息

- **请求地址**: `https://www.sophnet.com/api/open-apis/v1/chat/completions`
- **请求方式**: POST
- **认证方式**: Bearer Token

## 支持的模型

### 通义千问系列
- Qwen-Max
- Qwen-Plus
- Qwen-Turbo
- Qwen-Long
- Qwen2.5-72B-Instruct
- Qwen2.5-32B-Instruct
- Qwen2.5-14B-Instruct
- Qwen2.5-7B-Instruct
- Qwen2.5-3B-Instruct
- Qwen2.5-1.5B-Instruct
- Qwen2.5-0.5B-Instruct
- Qwen2.5-Coder-32B-Instruct
- Qwen2.5-Math-72B-Instruct

### 通义千问视觉系列
- Qwen-VL-Max
- Qwen-VL-Plus
- Qwen2-VL-72B-Instruct
- Qwen2-VL-7B-Instruct
- Qwen2-VL-2B-Instruct

### DeepSeek 系列
- DeepSeek-V3
- DeepSeek-Chat
- DeepSeek-Reasoner

### GLM 系列
- GLM-4-Plus
- GLM-4-0520
- GLM-4-Air
- GLM-4-AirX
- GLM-4-Flash
- GLM-4-Long
- GLM-4V-Plus
- GLM-4V

### 其他模型
- Yi-Lightning
- Yi-Large
- Yi-Medium
- Yi-Vision
- Doubao-Pro-32k
- Doubao-Pro-128k
- Doubao-Lite-32k
- Doubao-Lite-128k

## 使用方式

### 1. 配置渠道

在 new-api 管理后台添加算能云渠道：

- **渠道类型**: Sophnet (58)
- **Base URL**: `https://www.sophnet.com/api/open-apis`
- **API Key**: 你的算能云 API Key
- **模型**: 选择上述支持的文本模型

### 2. 标准 OpenAI 格式调用

```bash
curl https://your-new-api-domain/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "model": "Qwen-Max",
    "messages": [
      {
        "role": "system",
        "content": "You are a helpful assistant."
      },
      {
        "role": "user",
        "content": "Hello!"
      }
    ],
    "stream": false
  }'
```

### 3. 多模态调用（视觉模型）

```bash
curl https://your-new-api-domain/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "model": "Qwen-VL-Max",
    "messages": [
      {
        "role": "user",
        "content": [
          {
            "type": "text",
            "text": "describe the image in 100 words or less"
          },
          {
            "type": "image_url",
            "image_url": {
              "url": "https://example.com/image.jpg",
              "detail": "high"
            }
          }
        ]
      }
    ]
  }'
```

## 支持的参数

算能云支持以下 OpenAI 标准参数：

- `messages`: 聊天上下文信息（必填）
- `model`: 模型名称（必填）
- `stream`: 是否流式返回（默认 false）
- `max_tokens`: 最大回复长度
- `temperature`: 温度参数 [0, 2.0]
- `top_p`: 核采样参数
- `stop`: 停止词列表
- `presence_penalty`: 存在惩罚 [-2.0, 2.0]
- `frequency_penalty`: 频率惩罚 [-2.0, 2.0]
- `logprobs`: 是否返回对数概率
- `top_logprobs`: 返回的 top logprobs 数量 [0, 20]
- `response_format`: 响应格式（支持 JSON 模式）
- `tools`: 工具列表（仅支持 function）
- `tool_choice`: 工具选择策略
- `parallel_tool_calls`: 是否允许并行工具调用

## 特殊功能

### 思考模式

算能云支持思考模式，可以通过以下参数启用：

```json
{
  "model": "DeepSeek-Reasoner",
  "messages": [...],
  "chat_template_kwargs": {
    "enable_thinking": true
  },
  "thinking": {
    "budget_tokens": 1000
  }
}
```

或使用兼容参数：

```json
{
  "model": "DeepSeek-Reasoner",
  "messages": [...],
  "enable_thinking": true
}
```

### JSON 模式

```json
{
  "model": "Qwen-Max",
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful assistant that outputs JSON."
    },
    {
      "role": "user",
      "content": "Generate a user profile"
    }
  ],
  "response_format": {
    "type": "json_object"
  }
}
```

## 实现细节

算能云的文本生成功能通过以下方式集成到 new-api：

1. **API 类型映射**: `ChannelTypeSophnet` (58) → `APITypeOpenAI`
2. **适配器**: 直接使用 OpenAI 适配器（`openai.Adaptor`）
3. **请求转发**: 请求会被转发到 `https://www.sophnet.com/api/open-apis/v1/chat/completions`
4. **响应处理**: 使用标准 OpenAI 响应处理逻辑

## 注意事项

1. 算能云的文本生成接口完全兼容 OpenAI 格式，无需特殊转换
2. 支持流式和非流式两种模式
3. 支持多模态输入（视觉模型）
4. 支持工具调用（Function Calling）
5. 部分模型支持思考模式（如 DeepSeek-Reasoner）

## 相关文档

- [算能云图片生成](./SOPHNET_IMAGE.md)
- [算能云视频生成](./SOPHNET_VIDEO.md)
