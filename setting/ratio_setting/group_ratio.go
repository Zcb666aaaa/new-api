package ratio_setting

import (
	"encoding/json"
	"errors"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/QuantumNous/new-api/types"
)

var defaultGroupRatio = map[string]float64{
	"default": 1,
	"vip":     1,
	"svip":    1,
}

var groupRatioMap = types.NewRWMap[string, float64]()

var defaultGroupGroupRatio = map[string]map[string]float64{
	"vip": {
		"edit_this": 0.9,
	},
}

var groupGroupRatioMap = types.NewRWMap[string, map[string]float64]()

// GroupModelEntry 分组模型定价条目
// 支持三种模式（通过字段是否为非零值区分）：
//   - 阶梯计费覆盖：Tiers 非空，直接替换该模型的全局阶梯配置
//   - 按量计费覆盖：Input/Output 非零（$/1M tokens），同时可设置 CompletionRatio
//   - 按次/秒计费覆盖：Price 非零（直接替换 modelPrice）
//   - 倍率覆盖（兜底）：Ratio 非零（替换 modelRatio）
type GroupModelEntry struct {
	// 阶梯计费覆盖：非空时替换全局阶梯配置
	Tiers []PriceTier `json:"tiers,omitempty"`
	// 按量计费：输入价格 $/1M tokens（优先于 Ratio）
	Input float64 `json:"input,omitempty"`
	// 按量计费：输出价格 $/1M tokens
	Output float64 `json:"output,omitempty"`
	// 按次/秒计费：直接替换 modelPrice
	Price float64 `json:"price,omitempty"`
	// 按量计费倍率覆盖（当 Input/Output 未设置时生效）
	Ratio float64 `json:"ratio,omitempty"`
	// 补全倍率（仅 Ratio 模式时使用）
	CompletionRatio float64 `json:"completion_ratio,omitempty"`
}

// GroupModelRatio: map[userGroup]map[modelName]GroupModelEntry
// 用于给不同用户分组设置特定模型的独立倍率，覆盖全局模型倍率
var defaultGroupModelRatio = map[string]map[string]GroupModelEntry{}

var groupModelRatioMap = types.NewRWMap[string, map[string]GroupModelEntry]()

var defaultGroupSpecialUsableGroup = map[string]map[string]string{
	"vip": {
		"append_1":   "vip_special_group_1",
		"-:remove_1": "vip_removed_group_1",
	},
}

type GroupRatioSetting struct {
	GroupRatio              *types.RWMap[string, float64]            `json:"group_ratio"`
	GroupGroupRatio         *types.RWMap[string, map[string]float64]       `json:"group_group_ratio"`
	GroupSpecialUsableGroup *types.RWMap[string, map[string]string]        `json:"group_special_usable_group"`
	GroupModelRatio         *types.RWMap[string, map[string]GroupModelEntry] `json:"group_model_ratio"`
}

var groupRatioSetting GroupRatioSetting

func init() {
	groupSpecialUsableGroup := types.NewRWMap[string, map[string]string]()
	groupSpecialUsableGroup.AddAll(defaultGroupSpecialUsableGroup)

	groupRatioMap.AddAll(defaultGroupRatio)
	groupGroupRatioMap.AddAll(defaultGroupGroupRatio)
	groupModelRatioMap.AddAll(defaultGroupModelRatio)

	groupRatioSetting = GroupRatioSetting{
		GroupSpecialUsableGroup: groupSpecialUsableGroup,
		GroupRatio:              groupRatioMap,
		GroupGroupRatio:         groupGroupRatioMap,
		GroupModelRatio:         groupModelRatioMap,
	}

	config.GlobalConfig.Register("group_ratio_setting", &groupRatioSetting)
}

func GetGroupRatioSetting() *GroupRatioSetting {
	if groupRatioSetting.GroupSpecialUsableGroup == nil {
		groupRatioSetting.GroupSpecialUsableGroup = types.NewRWMap[string, map[string]string]()
		groupRatioSetting.GroupSpecialUsableGroup.AddAll(defaultGroupSpecialUsableGroup)
	}
	return &groupRatioSetting
}

func GetGroupRatioCopy() map[string]float64 {
	return groupRatioMap.ReadAll()
}

func ContainsGroupRatio(name string) bool {
	_, ok := groupRatioMap.Get(name)
	return ok
}

func GroupRatio2JSONString() string {
	return groupRatioMap.MarshalJSONString()
}

func UpdateGroupRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonString(groupRatioMap, jsonStr)
}

func GetGroupRatio(name string) float64 {
	ratio, ok := groupRatioMap.Get(name)
	if !ok {
		common.SysLog("group ratio not found: " + name)
		return 1
	}
	return ratio
}

func GetGroupGroupRatio(userGroup, usingGroup string) (float64, bool) {
	gp, ok := groupGroupRatioMap.Get(userGroup)
	if !ok {
		return -1, false
	}
	ratio, ok := gp[usingGroup]
	if !ok {
		return -1, false
	}
	return ratio, true
}

func GroupGroupRatio2JSONString() string {
	return groupGroupRatioMap.MarshalJSONString()
}

func UpdateGroupGroupRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonString(groupGroupRatioMap, jsonStr)
}

// GetGroupModelEntry 获取用户分组对某个模型的专属定价条目
// 返回条目和是否命中
func GetGroupModelEntry(userGroup, modelName string) (GroupModelEntry, bool) {
	modelMap, ok := groupModelRatioMap.Get(userGroup)
	if !ok {
		return GroupModelEntry{}, false
	}
	entry, ok := modelMap[modelName]
	if !ok {
		return GroupModelEntry{}, false
	}
	return entry, true
}

// GetGroupModelRatio 兼容旧接口：获取用户分组对某个模型的专属倍率（单值）
// 优先返回 Input（按量），其次 Price（按次），最后 Ratio
func GetGroupModelRatio(userGroup, modelName string) (float64, bool) {
	entry, ok := GetGroupModelEntry(userGroup, modelName)
	if !ok {
		return 0, false
	}
	if entry.Input != 0 {
		return entry.Input, true
	}
	if entry.Price != 0 {
		return entry.Price, true
	}
	if entry.Ratio != 0 {
		return entry.Ratio, true
	}
	return 0, false
}

func GroupModelRatio2JSONString() string {
	return groupModelRatioMap.MarshalJSONString()
}

func UpdateGroupModelRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonString(groupModelRatioMap, jsonStr)
}

// GetGroupModelCompletionRatio 获取用户分组对某个模型的专属补全倍率（仅 Ratio 模式有效）
func GetGroupModelCompletionRatio(userGroup, modelName string) (float64, bool) {
	entry, ok := GetGroupModelEntry(userGroup, modelName)
	if !ok || entry.CompletionRatio == 0 {
		return 0, false
	}
	return entry.CompletionRatio, true
}

// GetGroupModelInputOutputPrice 获取用户分组对某个模型的按量计费价格覆盖
// 返回 (inputPrice, outputPrice, found)，单位 $/1M tokens
func GetGroupModelInputOutputPrice(userGroup, modelName string) (float64, float64, bool) {
	entry, ok := GetGroupModelEntry(userGroup, modelName)
	if !ok || (entry.Input == 0 && entry.Output == 0) {
		return 0, 0, false
	}
	return entry.Input, entry.Output, true
}

// GetGroupModelTieredPrice 获取用户分组对某个阶梯计费模型的覆盖阶梯配置
// 返回 TieredPriceConfig 和是否命中
func GetGroupModelTieredPrice(userGroup, modelName string) (TieredPriceConfig, bool) {
	entry, ok := GetGroupModelEntry(userGroup, modelName)
	if !ok || len(entry.Tiers) == 0 {
		return TieredPriceConfig{}, false
	}
	return TieredPriceConfig{Tiers: entry.Tiers}, true
}

func CheckGroupRatio(jsonStr string) error {
	checkGroupRatio := make(map[string]float64)
	err := json.Unmarshal([]byte(jsonStr), &checkGroupRatio)
	if err != nil {
		return err
	}
	for name, ratio := range checkGroupRatio {
		if ratio < 0 {
			return errors.New("group ratio must be not less than 0: " + name)
		}
	}
	return nil
}
