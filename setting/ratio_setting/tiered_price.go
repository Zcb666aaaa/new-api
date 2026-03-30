package ratio_setting

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/types"
)

// TieredPriceConfig 阶梯计费配置
type TieredPriceConfig struct {
	Tiers []PriceTier `json:"tiers"`
}

// PriceTier 单个阶梯配置
type PriceTier struct {
	MinTokens        int     `json:"min_tokens"`                   // 输入最小token数（包含），0表示从0开始
	MaxTokens        int     `json:"max_tokens"`                   // 输入最大token数（不包含），-1表示无上限
	MinOutputTokens  int     `json:"min_output_tokens,omitempty"` // 输出最小token数（包含），0或缺省表示不限制
	MaxOutputTokens  int     `json:"max_output_tokens,omitempty"` // 输出最大token数（不包含），0或缺省/-1表示无上限
	InputPrice       float64 `json:"input"`                        // 输入价格（美元/百万token）
	OutputPrice      float64 `json:"output"`                       // 输出价格（美元/百万token）
}

// ModelTieredPrice stores model name -> TieredPriceConfig mapping
var modelTieredPriceMap = types.NewRWMap[string, TieredPriceConfig]()

// 默认阶梯价格配置（空）
var defaultModelTieredPrice = map[string]TieredPriceConfig{}

// InitTieredPriceSettings 初始化阶梯价格设置
func InitTieredPriceSettings() {
	modelTieredPriceMap.AddAll(defaultModelTieredPrice)
}

// GetModelTieredPriceMap 获取所有阶梯价格配置
func GetModelTieredPriceMap() map[string]TieredPriceConfig {
	return modelTieredPriceMap.ReadAll()
}

// ModelTieredPrice2JSONString 将阶梯价格配置序列化为JSON字符串
func ModelTieredPrice2JSONString() string {
	return modelTieredPriceMap.MarshalJSONString()
}

// UpdateModelTieredPriceByJSONString 通过JSON字符串更新阶梯价格配置
func UpdateModelTieredPriceByJSONString(jsonStr string) error {
	return types.LoadFromJsonStringWithCallback(modelTieredPriceMap, jsonStr, InvalidateExposedDataCache)
}

// GetModelTieredPrice 获取指定模型的阶梯价格配置
// 返回配置和是否找到
func GetModelTieredPrice(name string, printErr bool) (TieredPriceConfig, bool) {
	name = FormatMatchingModelName(name)

	if strings.HasSuffix(name, CompactModelSuffix) {
		config, ok := modelTieredPriceMap.Get(CompactWildcardModelKey)
		if !ok {
			if printErr {
				common.SysError("model tiered price not found: " + name)
			}
			return TieredPriceConfig{}, false
		}
		return config, true
	}

	config, ok := modelTieredPriceMap.Get(name)
	if !ok {
		if printErr {
			common.SysError("model tiered price not found: " + name)
		}
		return TieredPriceConfig{}, false
	}
	return config, true
}

// IsModelTieredPrice 判断模型是否使用阶梯计费
func IsModelTieredPrice(name string) bool {
	_, ok := GetModelTieredPrice(name, false)
	return ok
}

// GetTierForTokenCount 根据输入token数量获取对应的阶梯配置（兼容旧接口）
// Deprecated: 请使用 GetTierForRequest，以支持输入+输出二维分段
func GetTierForTokenCount(config TieredPriceConfig, inputTokens int) (inputPrice, outputPrice float64, found bool) {
	return GetTierForRequest(config, inputTokens, 0)
}

// GetTierForRequest 根据输入/输出token数量获取对应的阶梯配置（支持二维分段）
// 匹配规则：
//   - 输入端：min_tokens=0 表示无下限；max_tokens=-1 表示无上限
//   - 输出端：min_output_tokens=0 表示无下限；max_output_tokens=0或-1 表示无上限
//   - 两个条件同时满足才命中该阶梯
//   - 按配置顺序优先匹配，命中第一个满足条件的阶梯
//   - 未命中任何阶梯时回退到最后一个阶梯
func GetTierForRequest(config TieredPriceConfig, inputTokens, outputTokens int) (inputPrice, outputPrice float64, found bool) {
	if len(config.Tiers) == 0 {
		return 0, 0, false
	}

	for _, tier := range config.Tiers {
		// --- 输入端匹配 ---
		// min_tokens=0 表示无下限，否则 inputTokens 必须 >= min_tokens
		inputMinMatch := tier.MinTokens == 0 || inputTokens >= tier.MinTokens
		// max_tokens=-1 表示无上限，否则 inputTokens 必须 < max_tokens
		inputMaxMatch := tier.MaxTokens < 0 || inputTokens < tier.MaxTokens

		// --- 输出端匹配 ---
		// min_output_tokens=0 表示无下限，否则 outputTokens 必须 >= min_output_tokens
		outputMinMatch := tier.MinOutputTokens == 0 || outputTokens >= tier.MinOutputTokens
		// max_output_tokens=0或-1 表示无上限，否则 outputTokens 必须 < max_output_tokens
		outputMaxMatch := tier.MaxOutputTokens <= 0 || outputTokens < tier.MaxOutputTokens

		if inputMinMatch && inputMaxMatch && outputMinMatch && outputMaxMatch {
			return tier.InputPrice, tier.OutputPrice, true
		}
	}

	// 未找到匹配的阶梯（例如配置不连续），回退到最后一个阶梯
	lastTier := config.Tiers[len(config.Tiers)-1]
	return lastTier.InputPrice, lastTier.OutputPrice, true
}

// GetModelTieredPriceCopy 获取阶梯价格配置的副本（用于缓存暴露）
func GetModelTieredPriceCopy() map[string]TieredPriceConfig {
	all := modelTieredPriceMap.ReadAll()
	result := make(map[string]TieredPriceConfig, len(all))
	for k, v := range all {
		result[k] = v
	}
	return result
}
