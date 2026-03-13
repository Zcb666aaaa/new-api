# 算能云异步图片生成功能 - 变更总结

## 变更日期
2026年3月4日

## 变更概述
为 new-api 项目添加了算能云（Sophnet）异步图片生成功能的完整支持，允许用户通过 `/imagegenerator/task` 路径提交和查询图片生成任务。

## 修改的文件

### 1. 路由配置
**文件**: `router/relay-router.go`
- 添加了 `/imagegenerator` 路由组
- 支持 `POST /imagegenerator/task` 创建任务
- 支持 `GET /imagegenerator/task/:id` 查询任务

### 2. 请求分发中间件
**文件**: `middleware/distributor.go`
- 添加了对 `/imagegenerator/task` 路径的识别
- 自动设置平台标识为 `sophnet`
- 根据 HTTP 方法设置正确的中继模式

### 3. 算能云适配器
**文件**: `relay/channel/task/sophnet/adaptor.go`
- 更新 `ValidateRequestAndSetAction` 方法
- 保存上游模型名称到 `info.UpstreamModelName`
- 支持从请求体中提取实际的模型名称

### 4. 模型列表
**文件**: `relay/channel/task/sophnet/models.go`
- 添加了 5 个新的图片生成模型：
  - `Qwen-Image`
  - `Qwen-Image-Plus`
  - `Qwen-Image-Edit-2509`
  - `Z-Image-Turbo`
  - `Wan2.6-T2I`

### 5. 渠道测试
**文件**: `controller/channel-test.go`
- 将 `ChannelTypeSophnet` 添加到不支持测试的渠道列表

## 新增的文件

### 1. API 使用文档
**文件**: `docs/SOPHNET_IMAGE_GENERATOR.md`
- 完整的 API 接口文档
- 包含请求/响应示例
- Python 和 JavaScript 示例代码

### 2. 集成说明文档
**文件**: `docs/SOPHNET_IMAGE_INTEGRATION.md`
- 详细的技术实现说明
- 变更列表和文件清单
- 后续优化建议

### 3. 快速开始文档
**文件**: `docs/SOPHNET_IMAGE_QUICKSTART.md`
- 简洁的快速入门指南
- 常用参数说明
- 测试脚本使用方法

### 4. Linux/Mac 测试脚本
**文件**: `test_sophnet_image.sh`
- Bash 脚本
- 自动创建任务并轮询查询状态
- 支持环境变量配置

### 5. Windows 测试脚本
**文件**: `test_sophnet_image.bat`
- 批处理脚本
- 创建任务并显示响应
- UTF-8 编码支持中文

## 功能特性

### 支持的功能
1. **文生图（Text-to-Image）**
   - 支持正向和反向提示词
   - 自定义分辨率和种子值
   - 智能提示词改写

2. **图生图（Image-to-Image）**
   - 支持 1-3 张参考图片
   - URL 或 Base64 格式
   - 图像编辑和风格转换

3. **高级参数**
   - 提示词智能改写
   - 水印添加
   - JPEG 格式输出
   - 固定随机种子

### API 端点
- `POST /imagegenerator/task` - 创建任务
- `GET /imagegenerator/task/:id` - 查询任务

### 兼容性
保持对旧路径的支持：
- `/sophnet/submit/generate`
- `/sophnet/fetch/{taskId}`

## 使用示例

### 创建任务
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

### 查询任务
```bash
curl -X GET "http://localhost:3000/imagegenerator/task/task_abc123xyz" \
  -H "Authorization: Bearer sk-xxxx"
```

## 测试方法

### 自动测试
```bash
# Linux/Mac
export API_BASE="http://localhost:3000"
export API_KEY="sk-your-key"
bash test_sophnet_image.sh

# Windows
set API_BASE=http://localhost:3000
set API_KEY=sk-your-key
test_sophnet_image.bat
```

### 手动测试
1. 在管理后台创建算能云渠道
2. 配置 Base URL 和 API Key
3. 选择需要的模型
4. 使用 curl 或 Postman 测试接口

## 技术细节

### 状态映射
- `PENDING` → `queued`
- `RUNNING` → `in_progress`
- `SUCCEEDED` → `success`
- `FAILED` / `CANCELED` → `failure`

### 计费逻辑
- 提交时预扣费
- 完成后结算
- 失败时退款

### 错误处理
- 统一的 TaskError 格式
- 详细的错误信息
- HTTP 状态码映射

## 注意事项

1. **必填参数**: 图编辑模型必须提供 `input.images`
2. **轮询频率**: 建议 2-5 秒查询一次
3. **图片保存**: URL 有时效性，请及时下载
4. **渠道测试**: 异步任务不支持标准渠道测试

## 后续优化

1. 批量任务提交
2. 任务取消功能
3. WebSocket 推送
4. 图片缓存
5. 更多生成参数

## 验证清单

- [x] 路由配置正确
- [x] 中间件识别路径
- [x] 适配器处理请求
- [x] 模型列表完整
- [x] 文档齐全
- [x] 测试脚本可用
- [x] 代码无编译错误
- [x] 兼容旧路径

## 相关链接

- 算能云官方文档: https://www.sophnet.com/docs
- new-api 项目: https://github.com/QuantumNous/new-api

## 贡献者

本次功能由 AI 助手根据用户需求实现。

---

**状态**: ✅ 已完成
**测试**: ⏳ 待测试
**部署**: ⏳ 待部署
