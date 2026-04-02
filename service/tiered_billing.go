package service

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/shopspring/decimal"
)

// CalculateTieredQuota 计算阶梯计费的配额
// inputTokens: 输入token数量
// outputTokens: 输出token数量
// modelName: 模型名称
// groupRatio: 分组倍率
// userGroup: 用户分组（用于查找分组阶梯覆盖）
func CalculateTieredQuota(inputTokens, outputTokens int, modelName string, groupRatio float64, userGroup ...string) int {
	quota, _, _ := CalculateTieredQuotaWithInfo(inputTokens, outputTokens, modelName, groupRatio, userGroup...)
	return quota
}

// TieredCacheInfo 阶梯计费时的缓存 token 信息
// 用于 Claude (Anthropic) 格式下分离计算缓存 token 费用
// Claude 的 input_tokens 不包含缓存 token，需要额外传入缓存信息
type TieredCacheInfo struct {
	CacheReadTokens      int     // 缓存读取 token 数
	CacheCreationTokens  int     // 缓存创建 token 数
	CacheRatio           float64 // 缓存读取倍率（相对于输入价格）
	CacheCreationRatio   float64 // 缓存创建倍率（相对于输入价格）
}

// CalculateTieredQuotaWithInfo 计算阶梯计费的配额，同时返回命中的阶梯输入/输出价格（$/1M tokens）
// userGroup 可选，传入时优先使用分组阶梯覆盖配置
func CalculateTieredQuotaWithInfo(inputTokens, outputTokens int, modelName string, groupRatio float64, userGroup ...string) (quota int, inputPrice float64, outputPrice float64) {
	return CalculateTieredQuotaWithCacheInfo(inputTokens, outputTokens, modelName, groupRatio, nil, userGroup...)
}

// CalculateTieredQuotaWithCacheInfo 计算阶梯计费的配额，支持缓存 token 分离计费
// 当 cacheInfo 不为 nil 时（Claude 格式），inputTokens 不包含缓存 token：
//   - 用 (inputTokens + cache tokens) 作为总输入来匹配阶梯区间
//   - 基础输入按阶梯价格计费，缓存 token 按阶梯价格 * 对应倍率计费
// 当 cacheInfo 为 nil 时（OpenAI 格式），行为与原来完全一致
func CalculateTieredQuotaWithCacheInfo(inputTokens, outputTokens int, modelName string, groupRatio float64, cacheInfo *TieredCacheInfo, userGroup ...string) (quota int, inputPrice float64, outputPrice float64) {
	// 优先查找分组阶梯覆盖
	var config ratio_setting.TieredPriceConfig
	var ok bool
	if len(userGroup) > 0 && userGroup[0] != "" {
		config, ok = ratio_setting.GetGroupModelTieredPrice(userGroup[0], modelName)
	}
	if !ok {
		config, ok = ratio_setting.GetModelTieredPrice(modelName, false)
	}
	if !ok {
		return 0, 0, 0
	}

	// 计算用于阶梯匹配的总输入 token 数
	// Claude 格式: input_tokens 不含缓存，需要加上缓存 token 来匹配正确的阶梯
	// OpenAI 格式: prompt_tokens 已包含缓存，直接使用
	totalInputForTier := inputTokens
	if cacheInfo != nil {
		totalInputForTier += cacheInfo.CacheReadTokens + cacheInfo.CacheCreationTokens
	}

	// 根据总输入/输出 token 数量获取对应阶梯（支持二维分段）
	inP, outP, found := ratio_setting.GetTierForRequest(config, totalInputForTier, outputTokens)
	if !found {
		return 0, 0, 0
	}

	// 计算实际花费（美元）
	// 价格单位：美元/百万token
	dInputTokens := decimal.NewFromInt(int64(inputTokens))
	dOutputTokens := decimal.NewFromInt(int64(outputTokens))
	dInputPrice := decimal.NewFromFloat(inP)
	dOutputPrice := decimal.NewFromFloat(outP)
	dGroupRatio := decimal.NewFromFloat(groupRatio)
	dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
	dMillion := decimal.NewFromInt(1000000)

	// 计算输入费用
	// 基础输入 token 按阶梯价格计费
	inputCost := dInputTokens.Mul(dInputPrice).Div(dMillion)

	// 如果有缓存信息（Claude 格式），缓存 token 按阶梯价格 * 对应倍率单独计费
	if cacheInfo != nil {
		if cacheInfo.CacheReadTokens > 0 {
			dCacheReadTokens := decimal.NewFromInt(int64(cacheInfo.CacheReadTokens))
			dCacheRatio := decimal.NewFromFloat(cacheInfo.CacheRatio)
			cacheReadCost := dCacheReadTokens.Mul(dInputPrice).Mul(dCacheRatio).Div(dMillion)
			inputCost = inputCost.Add(cacheReadCost)
		}
		if cacheInfo.CacheCreationTokens > 0 {
			dCacheCreationTokens := decimal.NewFromInt(int64(cacheInfo.CacheCreationTokens))
			dCacheCreationRatio := decimal.NewFromFloat(cacheInfo.CacheCreationRatio)
			cacheCreationCost := dCacheCreationTokens.Mul(dInputPrice).Mul(dCacheCreationRatio).Div(dMillion)
			inputCost = inputCost.Add(cacheCreationCost)
		}
	}

	outputCost := dOutputTokens.Mul(dOutputPrice).Div(dMillion)

	// 总费用 * QuotaPerUnit * 分组倍率
	totalCost := inputCost.Add(outputCost)
	q := totalCost.Mul(dQuotaPerUnit).Mul(dGroupRatio)

	// 如果费用不为零但quota<=0，设置为最小值1
	if !totalCost.IsZero() && q.LessThanOrEqual(decimal.Zero) {
		q = decimal.NewFromInt(1)
	}

	return int(q.Round(0).IntPart()), inP, outP
}

// IsTieredPriceModel 检查模型是否使用阶梯计费
func IsTieredPriceModel(modelName string) bool {
	return ratio_setting.IsModelTieredPrice(modelName)
}
