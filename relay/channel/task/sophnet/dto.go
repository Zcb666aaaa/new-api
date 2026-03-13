package sophnet

// ==================== 图片生成相关 ====================

// CreateTaskRequest 创建图片任务请求
type CreateTaskRequest struct {
	Model      string            `json:"model"`
	Input      InputObject       `json:"input"`
	Parameters *ParametersObject `json:"parameters,omitempty"`
}

// InputObject 输入对象
type InputObject struct {
	Prompt         string   `json:"prompt"`
	Images         []string `json:"images,omitempty"`
	NegativePrompt string   `json:"negative_prompt,omitempty"`
}

// ParametersObject 参数对象
type ParametersObject struct {
	Size         string `json:"size,omitempty"`
	Seed         *int   `json:"seed,omitempty"`
	PromptExtend *bool  `json:"prompt_extend,omitempty"`
	Watermark    *bool  `json:"watermark,omitempty"`
	SaveToJpeg   *bool  `json:"save_to_jpeg,omitempty"`
}

// CreateTaskResponse 创建任务响应
type CreateTaskResponse struct {
	RequestId string       `json:"requestId"`
	Output    OutputObject `json:"output"`
	Usage     UsageObject  `json:"usage"`
	Code      string       `json:"code,omitempty"`
	Message   string       `json:"message,omitempty"`
}

// OutputObject 输出对象
type OutputObject struct {
	TaskId     string         `json:"taskId"`
	TaskStatus string         `json:"taskStatus"`
	Results    []ResultObject `json:"results,omitempty"`
	Code       string         `json:"code,omitempty"`
	Message    string         `json:"message,omitempty"`
}

// ResultObject 结果对象
type ResultObject struct {
	URL string `json:"url"`
}

// UsageObject 使用量对象
type UsageObject struct {
	ImageCount int `json:"imageCount"`
}

// QueryTaskResponse 查询任务响应（结构与创建响应相同）
type QueryTaskResponse = CreateTaskResponse

// ==================== 视频生成相关 ====================

// CreateVideoTaskRequest 创建视频任务请求
type CreateVideoTaskRequest struct {
	Model                  string                  `json:"model"`
	Content                []VideoContentObject    `json:"content"`
	Parameters             *VideoParametersObject  `json:"parameters,omitempty"`
	CallbackURL            string                  `json:"callback_url,omitempty"`
	ReturnLastFrame        *bool                   `json:"return_last_frame,omitempty"`
	ServiceTier            string                  `json:"service_tier,omitempty"`
	ExecutionExpiresAfter  *int                    `json:"execution_expires_after,omitempty"`
	GenerateAudio          *bool                   `json:"generate_audio,omitempty"`
	Draft                  *bool                   `json:"draft,omitempty"`
	Subjects               []VideoSubjectObject    `json:"subjects,omitempty"`
}

// VideoContentObject 视频内容对象
type VideoContentObject struct {
	Type           string                `json:"type"` // text, image_url, draft_task
	Text           string                `json:"text,omitempty"`
	NegativePrompt string                `json:"negative_prompt,omitempty"`
	ImageUrl       *VideoImageObject     `json:"image_url,omitempty"`
	Image          *VideoImageObject     `json:"image,omitempty"`
	AudioURL       string                `json:"audio_url,omitempty"`
	Role           string                `json:"role,omitempty"` // first_frame, last_frame, reference_image
	DraftTask      *VideoDraftTaskObject `json:"draft_task,omitempty"`
}

// VideoImageObject 视频图片对象
type VideoImageObject struct {
	URL string `json:"url"`
}

// VideoDraftTaskObject 样片任务对象
type VideoDraftTaskObject struct {
	ID string `json:"id"`
}

// VideoParametersObject 视频参数对象
type VideoParametersObject struct {
	Size             string `json:"size,omitempty"`
	Duration         *int   `json:"duration,omitempty"`
	Seed             string `json:"seed,omitempty"`
	SubdivisionLevel string `json:"subdivisionlevel,omitempty"`
	FileFormat       string `json:"fileformat,omitempty"`
}

// VideoSubjectObject 视频主体对象
type VideoSubjectObject struct {
	ID      string   `json:"id"`
	Images  []string `json:"images"`
	VoiceID string   `json:"voice_id,omitempty"`
}

// CreateVideoTaskResponse 创建视频任务响应
type CreateVideoTaskResponse struct {
	Status    int                `json:"status"`
	Message   string             `json:"message"`
	Result    VideoResultObject  `json:"result"`
	Timestamp int64              `json:"timestamp"`
}

// VideoResultObject 视频结果对象
type VideoResultObject struct {
	TaskID string `json:"task_id"`
}

// QueryVideoTaskResponse 查询视频任务响应
type QueryVideoTaskResponse struct {
	Status    int                      `json:"status"`
	Message   string                   `json:"message"`
	Result    VideoTaskDetailObject    `json:"result"`
	Timestamp int64                    `json:"timestamp"`
}

// VideoTaskDetailObject 视频任务详情对象
type VideoTaskDetailObject struct {
	ID        string                 `json:"id"`
	Model     string                 `json:"model"`
	Status    string                 `json:"status"` // queued, running, succeeded, failed, cancelled
	Error     *VideoErrorObject      `json:"error,omitempty"`
	Content   *VideoTaskContent      `json:"content,omitempty"`
	Usage     *VideoUsageObject      `json:"usage,omitempty"`
	CreatedAt int64                  `json:"created_at"`
	UpdatedAt int64                  `json:"updated_at"`
}

// VideoTaskContent 视频任务内容
type VideoTaskContent struct {
	VideoURL      string `json:"video_url,omitempty"`
	FileURL       string `json:"file_url,omitempty"`
	LastFrameURL  string `json:"last_frame_url,omitempty"`
}

// VideoErrorObject 视频错误对象
type VideoErrorObject struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// VideoUsageObject 视频使用量对象
type VideoUsageObject struct {
	Duration          *int   `json:"duration,omitempty"`
	Resolution        string `json:"resolution,omitempty"`
	VideoCount        *int   `json:"video_count,omitempty"`
	Ratio             string `json:"ratio,omitempty"`
	Frames            *int   `json:"frames,omitempty"`
	FPS               *int   `json:"fps,omitempty"`
	CompletionTokens  *int   `json:"completion_tokens,omitempty"`
	SubdivisionLevel  string `json:"subdivisionlevel,omitempty"`
	FileFormat        string `json:"fileformat,omitempty"`
}

// ==================== 通用错误响应 ====================

// SophnetErrorResponse 算能通用错误响应
type SophnetErrorResponse struct {
	Status    int    `json:"status"`
	Message   string `json:"message"`
	Result    any    `json:"result"`
	Timestamp int64  `json:"timestamp"`
}
