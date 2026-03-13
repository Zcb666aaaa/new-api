package sophnet

// 图片生成状态映射
const (
	StatusPending   = "PENDING"
	StatusRunning   = "RUNNING"
	StatusSucceeded = "SUCCEEDED"
	StatusFailed    = "FAILED"
	StatusUnknown   = "UNKNOWN"
	StatusCanceled  = "CANCELED"
)

// 视频生成状态映射
const (
	VideoStatusQueued    = "queued"
	VideoStatusRunning   = "running"
	VideoStatusSucceeded = "succeeded"
	VideoStatusFailed    = "failed"
	VideoStatusCancelled = "cancelled"
)

// 上游支持的文本生成模型（Chat Completions）
var UpstreamChatModels = []string{
	// 通义千问系列
	"Qwen-Max",
	"Qwen-Plus",
	"Qwen-Turbo",
	"Qwen-Long",
	"Qwen2.5-72B-Instruct",
	"Qwen2.5-32B-Instruct",
	"Qwen2.5-14B-Instruct",
	"Qwen2.5-7B-Instruct",
	"Qwen2.5-3B-Instruct",
	"Qwen2.5-1.5B-Instruct",
	"Qwen2.5-0.5B-Instruct",
	"Qwen2.5-Coder-32B-Instruct",
	"Qwen2.5-Math-72B-Instruct",
	// 通义千问视觉系列
	"Qwen-VL-Max",
	"Qwen-VL-Plus",
	"Qwen2-VL-72B-Instruct",
	"Qwen2-VL-7B-Instruct",
	"Qwen2-VL-2B-Instruct",
	// DeepSeek 系列
	"DeepSeek-V3",
	"DeepSeek-Chat",
	"DeepSeek-Reasoner",
	// GLM 系列
	"GLM-4-Plus",
	"GLM-4-0520",
	"GLM-4-Air",
	"GLM-4-AirX",
	"GLM-4-Flash",
	"GLM-4-Long",
	"GLM-4V-Plus",
	"GLM-4V",
	// 其他模型
	"Yi-Lightning",
	"Yi-Large",
	"Yi-Medium",
	"Yi-Vision",
	"Doubao-Pro-32k",
	"Doubao-Pro-128k",
	"Doubao-Lite-32k",
	"Doubao-Lite-128k",
}

// 上游支持的图片生成模型
var UpstreamImageModels = []string{
	"Qwen-Image",
	"Qwen-Image-Plus",
	"Qwen-Image-Edit-2509",
	"Z-Image-Turbo",
	"Wan2.6-T2I",
}

// 上游支持的视频生成模型
var UpstreamVideoModels = []string{
	// 万相系列
	"Wan2.2-T2V-Plus",
	"Wan2.2-I2V-Plus",
	"Wan2.5-T2V-Preview",
	"Wan2.5-I2V-Preview",
	"Wan2.6-T2V",
	"Wan2.6-I2V",
	"Wan2.2-T2V-A14B",
	"Wan2.2-I2V-A14B",
	// 字节跳动系列
	"Seedance-1.5-Pro",
	"Seedance-1.0-Pro",
	"Seedance-1.0-Pro-Fast",
	"Seedance-1.0-Lite-T2V",
	"Seedance-1.0-Lite-I2V",
	"Doubao-Seed3D",
	// 生数系列
	"ViduQ2",
	"ViduQ2-turbo",
	"ViduQ2-pro",
	"ViduQ2-pro-fast",
	"ViduQ1",
	"ViduQ1-classic",
	"Vidu2.0",
	"Vidu1.5",
}
