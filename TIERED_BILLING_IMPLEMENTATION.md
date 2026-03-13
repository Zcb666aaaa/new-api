# 阶梯计费功能实现总结

## 功能概述
实现了基于输入token数量的阶梯计费功能,支持不同token范围使用不同的输入/输出价格。

## 实现文件清单

### 后端实现

1. **setting/ratio_setting/tiered_price.go** (新建)
   - `TieredPriceConfig` 结构体：阶梯计费配置
   - `PriceTier` 结构体：单个阶梯定义
   - `GetModelTieredPrice()`: 获取模型阶梯配置
   - `IsModelTieredPrice()`: 判断是否使用阶梯计费
   - `GetTierForTokenCount()`: 根据token数量获取对应阶梯

2. **setting/ratio_setting/model_ratio.go** (修改)
   - 在 `InitRatioSettings()` 中添加 `InitTieredPriceSettings()`

3. **setting/ratio_setting/exposed_cache.go** (修改)
   - 在暴露数据中添加 `model_tiered_price`

4. **model/option.go** (修改)
   - 添加 `ModelTieredPrice` 到配置项

5. **model/pricing.go** (修改)
   - 添加 QuotaType = 3 表示阶梯计费
   - 在 updatePricing() 中检测阶梯计费模型

6. **service/tiered_billing.go** (新建)
   - `CalculateTieredQuota()`: 计算阶梯计费的配额
   - `IsTieredPriceModel()`: 检查是否为阶梯计费模型

7. **relay/compatible_handler.go** (修改)
   - 在 `postConsumeQuota()` 中添加阶梯计费分支

### 前端实现

1. **web/src/pages/Setting/Ratio/TieredPricingEditor.jsx** (新建)
   - 阶梯计费可视化编辑器
   - 支持添加/编辑/删除阶梯配置
   - 每个阶梯包含: min_tokens, max_tokens, input价格, output价格

2. **web/src/components/settings/RatioSetting.jsx** (修改)
   - 导入 TieredPricingEditor
   - 添加"阶梯计费设置"标签页
   - 在 inputs 中添加 ModelTieredPrice

3. **web/src/i18n/locales/zh.json** (修改)
   - 添加阶梯计费相关中文翻译

4. **web/src/i18n/locales/en.json** (修改)
   - 添加阶梯计费相关英文翻译

## QuotaType 类型定义

- **0**: 按量付费 (基于倍率ratio)
- **1**: 按次付费 (固定价格ModelPrice)
- **2**: 按秒付费 (ModelPricePerSecond)
- **3**: 阶梯计费 (ModelTieredPrice) ← **新增**

## 阶梯计费数据结构

```json
{
  "模型名称": {
    "tiers": [
      {
        "min_tokens": 0,
        "max_tokens": 32000,
        "input": 4.0,
        "output": 18.0
      },
      {
        "min_tokens": 32000,
        "max_tokens": 64000,
        "input": 6.0,
        "output": 22.0
      },
      {
        "min_tokens": 64000,
        "max_tokens": -1,
        "input": 8.0,
        "output": 26.0
      }
    ]
  }
}
```

- `min_tokens`: 最小token数（包含）,0表示从0开始
- `max_tokens`: 最大token数（不包含）,-1表示无限制
- `input`: 输入价格（美元/百万token）
- `output`: 输出价格（美元/百万token）

## 计费逻辑

1. 检测模型是否使用阶梯计费 (`IsTieredPriceModel`)
2. 根据输入token数量查找匹配的阶梯 (`GetTierForTokenCount`)
3. 使用该阶梯的输入/输出价格计算费用
4. 公式: `quota = (inputTokens * inputPrice / 1000000 + outputTokens * outputPrice / 1000000) * QuotaPerUnit * GroupRatio`

## 使用方式

### 管理员配置

1. 进入"设置" → "比率与模型倍率设置"
2. 选择"阶梯计费设置"标签页
3. 点击"添加阶梯计费模型"
4. 输入模型名称
5. 添加多个阶梯配置
6. 保存更改

### 计费示例

假设配置如上JSON所示:

- 输入 20000 tokens, 输出 5000 tokens
  - 匹配第1阶梯 (0 <= 20000 < 32000)
  - 费用 = (20000 * 4 / 1000000 + 5000 * 18 / 1000000) * QuotaPerUnit * GroupRatio

- 输入 50000 tokens, 输出 5000 tokens  
  - 匹配第2阶梯 (32000 <= 50000 < 64000)
  - 费用 = (50000 * 6 / 1000000 + 5000 * 22 / 1000000) * QuotaPerUnit * GroupRatio

- 输入 100000 tokens, 输出 5000 tokens
  - 匹配第3阶梯 (64000 <= 100000)
  - 费用 = (100000 * 8 / 1000000 + 5000 * 26 / 1000000) * QuotaPerUnit * GroupRatio

## 注意事项

1. 阶梯计费与按量计费、按次计费互斥,同一模型只能使用一种计费方式
2. 阶梯边界遵循左闭右开原则: [min_tokens, max_tokens)
3. 建议最后一个阶梯的 max_tokens 设置为 -1 表示无限制
4. 价格单位统一为美元/百万token
5. 系统会根据分组倍率进一步调整最终费用
