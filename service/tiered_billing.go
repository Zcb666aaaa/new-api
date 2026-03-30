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

// CalculateTieredQuotaWithInfo 计算阶梯计费的配额，同时返回命中的阶梯输入/输出价格（$/1M tokens）
// userGroup 可选，传入时优先使用分组阶梯覆盖配置
func CalculateTieredQuotaWithInfo(inputTokens, outputTokens int, modelName string, groupRatio float64, userGroup ...string) (quota int, inputPrice float64, outputPrice float64) {
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

	// 根据输入/输出token数量获取对应阶梯（支持二维分段）
	inP, outP, found := ratio_setting.GetTierForRequest(config, inputTokens, outputTokens)
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

	// 计算输入和输出的费用
	// 价格是 $/1M tokens，所以需要除以 1,000,000
	inputCost := dInputTokens.Mul(dInputPrice).Div(decimal.NewFromInt(1000000))
	outputCost := dOutputTokens.Mul(dOutputPrice).Div(decimal.NewFromInt(1000000))

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
