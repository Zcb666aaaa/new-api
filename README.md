<div align="center">

# KolitePay · New API Fork

🍥 **基于 [new-api](https://github.com/QuantumNous/new-api) 的定制化 AI 网关**

</div>

---

## 📌 上游项目

本项目 Fork 自 [QuantumNous/new-api](https://github.com/QuantumNous/new-api)，在其基础上进行了若干定制化修改。

---

## 🔧 修改内容

### 新增功能

| 功能 | 说明 |
|------|------|
| **算能渠道商支持** | 新增 SophNet（算能）AI 渠道接入，兼容 OpenAI 格式 |
| **按秒计费（QuotaType=2）** | 支持基于响应时长的秒级计费模式 |
| **梯度计费（QuotaType=3）** | 支持按输入 Token 数量分阶梯定价，灵活应对不同用量场景 |

### 优化调整

- 修改部分提供商的默认 API 地址
- 调整支付界面 UI 及交互体验

---

## 💰 计费类型说明

| QuotaType | 模式 | 说明 |
|-----------|------|------|
| `0` | 按量付费 | 基于倍率 ratio 计算 |
| `1` | 按次付费 | 固定 ModelPrice |
| `2` | 按秒付费 | 基于 ModelPricePerSecond |
| `3` | 梯度计费 | 基于输入 Token 数量分阶梯定价 |


---


## 📄 License

遵循上游项目 [AGPL-3.0 License]。

