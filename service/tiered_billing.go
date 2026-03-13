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
func CalculateTieredQuota(inputTokens, outputTokens int, modelName string, groupRatio float64) int {
	config, ok := ratio_setting.GetModelTieredPrice(modelName, false)
	if !ok {
		// 如果没有阶梯配置，返回0（这不应该发生）
		return 0
	}

	// 根据输入token数量获取对应阶梯
	inputPrice, outputPrice, found := ratio_setting.GetTierForTokenCount(config, inputTokens)
	if !found {
		return 0
	}

	// 计算实际花费（美元）
	// 价格单位：美元/百万token
	dInputTokens := decimal.NewFromInt(int64(inputTokens))
	dOutputTokens := decimal.NewFromInt(int64(outputTokens))
	dInputPrice := decimal.NewFromFloat(inputPrice)
	dOutputPrice := decimal.NewFromFloat(outputPrice)
	dGroupRatio := decimal.NewFromFloat(groupRatio)
	dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)

	// 计算输入和输出的费用
	// 价格是 $/1M tokens，所以需要除以 1,000,000
	inputCost := dInputTokens.Mul(dInputPrice).Div(decimal.NewFromInt(1000000))
	outputCost := dOutputTokens.Mul(dOutputPrice).Div(decimal.NewFromInt(1000000))
	
	// 总费用 * QuotaPerUnit * 分组倍率
	totalCost := inputCost.Add(outputCost)
	quota := totalCost.Mul(dQuotaPerUnit).Mul(dGroupRatio)

	// 如果费用不为零但quota<=0，设置为最小值1
	if !totalCost.IsZero() && quota.LessThanOrEqual(decimal.Zero) {
		quota = decimal.NewFromInt(1)
	}

	return int(quota.Round(0).IntPart())
}

// IsTieredPriceModel 检查模型是否使用阶梯计费
func IsTieredPriceModel(modelName string) bool {
	return ratio_setting.IsModelTieredPrice(modelName)
}
