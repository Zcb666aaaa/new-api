package sophnet

// ModelList 导出所有支持的模型列表（图片+视频+文本）
var ModelList = []string{}

// ChatModelList 导出文本生成模型列表
var ChatModelList = []string{}

var ChannelName = "sophnet"

func init() {
	// 合并图片、视频和文本模型
	ModelList = append(ModelList, UpstreamImageModels...)
	ModelList = append(ModelList, UpstreamVideoModels...)
	ModelList = append(ModelList, UpstreamChatModels...)
	
	// 单独导出文本模型列表
	ChatModelList = append(ChatModelList, UpstreamChatModels...)
}
