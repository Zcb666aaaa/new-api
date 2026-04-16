package doubao

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay/channel"
	taskcommon "github.com/QuantumNous/new-api/relay/channel/task/taskcommon"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// ============================
// Request / Response structures
// ============================

type ContentItem struct {
	Type     string          `json:"type"`                // "text", "image_url" or "video"
	Text     string          `json:"text,omitempty"`      // for text type
	ImageURL *ImageURL       `json:"image_url,omitempty"` // for image_url type
	VideoURL *VideoReference `json:"video_url,omitempty"` // for video_url type
	AudioURL *AudioReference `json:"audio_url,omitempty"` // for audio_url type
	Role     string          `json:"role,omitempty"`      // reference_image / first_frame / last_frame / reference_video / reference_audio
}

type ImageURL struct {
	URL string `json:"url"`
}

type VideoReference struct {
	URL string `json:"url"` // Draft video URL
}
type AudioReference struct {
	URL string `json:"url"` // Draft audio URL
}
type requestPayload struct {
	Model                 string         `json:"model"`
	Content               []ContentItem  `json:"content"`
	CallbackURL           string         `json:"callback_url,omitempty"`
	ReturnLastFrame       *dto.BoolValue `json:"return_last_frame,omitempty"`
	ServiceTier           string         `json:"service_tier,omitempty"`
	ExecutionExpiresAfter dto.IntValue   `json:"execution_expires_after,omitempty"`
	GenerateAudio         *dto.BoolValue `json:"generate_audio,omitempty"`
	Draft                 *dto.BoolValue `json:"draft,omitempty"`
	Resolution            string         `json:"resolution,omitempty"`
	Ratio                 string         `json:"ratio,omitempty"`
	Duration              dto.IntValue   `json:"duration,omitempty"`
	Frames                dto.IntValue   `json:"frames,omitempty"`
	Seed                  dto.IntValue   `json:"seed,omitempty"`
	CameraFixed           *dto.BoolValue `json:"camera_fixed,omitempty"`
	Watermark             *dto.BoolValue `json:"watermark,omitempty"`
}

type responsePayload struct {
	ID string `json:"id"` // task_id
}

type responseTask struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Status  string `json:"status"`
	Content struct {
		VideoURL string `json:"video_url"`
	} `json:"content"`
	Seed            int    `json:"seed"`
	Resolution      string `json:"resolution"`
	Duration        int    `json:"duration"`
	Ratio           string `json:"ratio"`
	FramesPerSecond int    `json:"framespersecond"`
	ServiceTier     string `json:"service_tier"`
	Usage           struct {
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
}

// ============================
// Adaptor implementation
// ============================

type TaskAdaptor struct {
	taskcommon.BaseBilling
	ChannelType int
	apiKey      string
	baseURL     string
}

func (a *TaskAdaptor) Init(info *relaycommon.RelayInfo) {
	a.ChannelType = info.ChannelType
	a.baseURL = info.ChannelBaseUrl
	a.apiKey = info.ApiKey
}

// ValidateRequestAndSetAction parses body, validates fields and sets default action.
func (a *TaskAdaptor) ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.RelayInfo) (taskErr *dto.TaskError) {
	// Accept only POST /v1/video/generations as "generate" action.
	return relaycommon.ValidateBasicTaskRequest(c, info, constant.TaskActionGenerate)
}

// EstimateBilling 将预扣费额度强制固定为 2.5 × QuotaPerUnit（即 2.5 美元当量），
// 忽略模型单价计算结果，不返回额外 OtherRatios 倍率。
// 任务完成后由 ParseTaskResult 返回的 TotalTokens 进行差额结算。
func (a *TaskAdaptor) EstimateBilling(_ *gin.Context, info *relaycommon.RelayInfo) map[string]float64 {
	info.PriceData.Quota = int(2.5 * common.QuotaPerUnit)
	return nil
}

// BuildRequestURL constructs the upstream URL.
func (a *TaskAdaptor) BuildRequestURL(info *relaycommon.RelayInfo) (string, error) {
	return fmt.Sprintf("%s/tasks", a.baseURL), nil
}

// BuildRequestHeader sets required headers.
func (a *TaskAdaptor) BuildRequestHeader(c *gin.Context, req *http.Request, info *relaycommon.RelayInfo) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	return nil
}

// BuildRequestBody converts request into Doubao specific format.
func (a *TaskAdaptor) BuildRequestBody(c *gin.Context, info *relaycommon.RelayInfo) (io.Reader, error) {
	req, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return nil, err
	}

	body, err := a.convertToRequestPayload(&req)
	if err != nil {
		return nil, errors.Wrap(err, "convert request payload failed")
	}
	if info.IsModelMapped {
		body.Model = info.UpstreamModelName
	} else {
		info.UpstreamModelName = body.Model
	}
	data, err := common.Marshal(body)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

// DoRequest delegates to common helper.
func (a *TaskAdaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	return channel.DoTaskApiRequest(a, c, info, requestBody)
}

// DoResponse handles upstream response, returns taskID etc.
func (a *TaskAdaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (taskID string, taskData []byte, taskErr *dto.TaskError) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
		return
	}
	_ = resp.Body.Close()


	// Parse Doubao response
	var dResp responsePayload
	if err := common.Unmarshal(responseBody, &dResp); err != nil {
		taskErr = service.TaskErrorWrapper(errors.Wrapf(err, "body: %s", responseBody), "unmarshal_response_body_failed", http.StatusInternalServerError)
		return
	}

	if dResp.ID == "" {
		taskErr = service.TaskErrorWrapper(fmt.Errorf("task_id is empty"), "invalid_response", http.StatusInternalServerError)
		return
	}

	ov := dto.NewOpenAIVideo()
	ov.ID = info.PublicTaskID
	ov.TaskID = info.PublicTaskID
	ov.CreatedAt = time.Now().Unix()
	ov.Model = info.OriginModelName

	c.JSON(http.StatusOK, ov)
	return dResp.ID, responseBody, nil
}

// FetchTask fetch task status
func (a *TaskAdaptor) FetchTask(baseUrl, key string, body map[string]any, proxy string) (*http.Response, error) {
	taskID, ok := body["task_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid task_id")
	}

	uri := fmt.Sprintf("%s/tasks/%s", baseUrl, taskID)

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	client, err := service.GetHttpClientWithProxy(proxy)
	if err != nil {
		return nil, fmt.Errorf("new proxy http client failed: %w", err)
	}
	return client.Do(req)
}

func (a *TaskAdaptor) GetModelList() []string {
	return ModelList
}

func (a *TaskAdaptor) GetChannelName() string {
	return ChannelName
}

func (a *TaskAdaptor) convertToRequestPayload(req *relaycommon.TaskSubmitReq) (*requestPayload, error) {
	r := requestPayload{
		Model:   req.Model,
		Content: []ContentItem{},
	}

	// Add text prompt
	if req.Prompt != "" {
		r.Content = append(r.Content, ContentItem{
			Type: "text",
			Text: req.Prompt,
		})
	}

	// Add images if present
	if req.HasImage() {
		// 尝试从 metadata 获取图片角色配置
		var imageRoles []string
		if req.Metadata != nil {
			if roles, ok := req.Metadata["image_roles"].([]interface{}); ok {
				for _, role := range roles {
					if roleStr, ok := role.(string); ok {
						imageRoles = append(imageRoles, roleStr)
					}
				}
			}
		}

		for i, imgURL := range req.Images {
			item := ContentItem{
				Type: "image_url",
				ImageURL: &ImageURL{
					URL: imgURL,
				},
			}
			// 如果有对应的角色配置，设置 Role 字段
			// 支持的角色: first_frame, last_frame, reference_image
			if i < len(imageRoles) && imageRoles[i] != "" {
				item.Role = imageRoles[i]
			}
			r.Content = append(r.Content, item)
		}
	}

	// 从 metadata["references"] 中读取媒体资源，自动判断类型并拼接到 content
	// 每个元素可以是 URL（通过扩展名判断）或 base64（通过 data: 前缀判断）
	// 角色由类型自动决定：image_url→reference_image，video_url→reference_video，audio_url→reference_audio
	if req.Metadata != nil {
		if refs, ok := req.Metadata["references"].([]interface{}); ok {
			for _, v := range refs {
				ref, ok := v.(string)
				if !ok || ref == "" {
					continue
				}
				r.Content = append(r.Content, detectAndBuildContentItem(ref, ""))
			}
		}
	}

	// 处理顶层字段 - 这些值可以被 metadata 覆盖
	if req.Duration > 0 {
		r.Duration = dto.IntValue(req.Duration)
	}
	// Size 字段可以映射到 resolution 或 ratio
	if req.Size != "" {
		// 如果是 "1080p", "720p" 格式，设置为 resolution
		// 如果是 "16:9", "9:16" 格式，设置为 ratio
		if strings.Contains(req.Size, ":") {
			r.Ratio = req.Size
		} else {
			r.Resolution = req.Size
		}
	}

	// 从 metadata 中解析其他参数（会覆盖上面的默认值）
	if err := taskcommon.UnmarshalMetadata(req.Metadata, &r); err != nil {
		return nil, errors.Wrap(err, "unmarshal metadata failed")
	}

	return &r, nil
}

func (a *TaskAdaptor) ParseTaskResult(respBody []byte) (*relaycommon.TaskInfo, error) {

	resTask := responseTask{}
	if err := common.Unmarshal(respBody, &resTask); err != nil {
		return nil, errors.Wrap(err, "unmarshal task result failed")
	}

	taskResult := relaycommon.TaskInfo{
		Code: 0,
	}

	// Map Doubao status to internal status
	switch resTask.Status {
	case "pending", "queued":
		taskResult.Status = model.TaskStatusQueued
		taskResult.Progress = "10%"
	case "processing", "running":
		taskResult.Status = model.TaskStatusInProgress
		taskResult.Progress = "50%"
	case "succeeded":
		taskResult.Status = model.TaskStatusSuccess
		taskResult.Progress = "100%"
		taskResult.Url = resTask.Content.VideoURL
		// 解析 usage 信息用于按倍率计费
		taskResult.CompletionTokens = resTask.Usage.CompletionTokens
		taskResult.TotalTokens = resTask.Usage.TotalTokens
	case "failed":
		taskResult.Status = model.TaskStatusFailure
		taskResult.Progress = "100%"
		taskResult.Reason = "task failed"
	default:
		// Unknown status, treat as processing
		taskResult.Status = model.TaskStatusInProgress
		taskResult.Progress = "30%"
	}

	return &taskResult, nil
}

func (a *TaskAdaptor) ConvertToOpenAIVideo(originTask *model.Task) ([]byte, error) {
	var dResp responseTask
	if err := common.Unmarshal(originTask.Data, &dResp); err != nil {
		return nil, errors.Wrap(err, "unmarshal doubao task data failed")
	}

	openAIVideo := dto.NewOpenAIVideo()
	openAIVideo.ID = originTask.TaskID
	openAIVideo.TaskID = originTask.TaskID
	openAIVideo.Status = originTask.Status.ToVideoStatus()
	openAIVideo.SetProgressStr(originTask.Progress)
	openAIVideo.SetMetadata("url", dResp.Content.VideoURL)
	openAIVideo.CreatedAt = originTask.CreatedAt
	openAIVideo.CompletedAt = originTask.UpdatedAt
	openAIVideo.Model = originTask.Properties.OriginModelName

	if dResp.Status == "failed" {
		openAIVideo.Error = &dto.OpenAIVideoError{
			Message: "task failed",
			Code:    "failed",
		}
	}

	return common.Marshal(openAIVideo)
}

// detectAndBuildContentItem 根据 URL 扩展名或 base64 的 MIME 前缀自动判断媒体类型，
// 构建对应的 ContentItem（image_url / video_url / audio_url）。
//
// 检测优先级：
//  1. base64 字符串（以 "data:" 开头）→ 取 MIME 类型前缀判断
//  2. URL 扩展名判断
//  3. 无法判断时默认当作图片处理
func detectAndBuildContentItem(ref, role string) ContentItem {
	mediaType := detectMediaType(ref)
	switch mediaType {
	case "video":
		if role == "" {
			role = "reference_video"
		}
		return ContentItem{
			Type:     "video_url",
			VideoURL: &VideoReference{URL: ref},
			Role:     role,
		}
	case "audio":
		if role == "" {
			role = "reference_audio"
		}
		return ContentItem{
			Type:     "audio_url",
			AudioURL: &AudioReference{URL: ref},
			Role:     role,
		}
	default: // image
		if role == "" {
			role = "reference_image"
		}
		return ContentItem{
			Type:     "image_url",
			ImageURL: &ImageURL{URL: ref},
			Role:     role,
		}
	}
}

// detectMediaType 根据 data URI 前缀或 URL 扩展名判断媒体类型，
// 返回 "image"、"video" 或 "audio"。
func detectMediaType(ref string) string {
	// base64 data URI：data:<mime>;base64,<data>
	if strings.HasPrefix(ref, "data:") {
		// 只取 MIME 类型部分（分号前），避免对整个 base64 数据做 ToLower
		mimeStr := ref[5:] // 去掉 "data:"
		if semi := strings.IndexByte(mimeStr, ';'); semi != -1 {
			mimeStr = mimeStr[:semi]
		}
		mime := strings.ToLower(mimeStr)
		var mediaType string
		switch {
		case strings.HasPrefix(mime, "video/"):
			mediaType = "video"
		case strings.HasPrefix(mime, "audio/"):
			mediaType = "audio"
		default:
			mediaType = "image"
		}
		common.SysLog(fmt.Sprintf("[detectMediaType] mime=%s, mediaType=%s", mime, mediaType))
		return mediaType
	}

	// 通过 URL 扩展名判断（忽略查询参数）
	u := ref
	if idx := strings.Index(u, "?"); idx != -1 {
		u = u[:idx]
	}
	ext := strings.ToLower(strings.TrimPrefix(path.Ext(u), "."))
	var mediaType string
	switch ext {
	case "mp4", "mov", "webm", "avi", "mkv", "flv", "wmv", "m4v":
		mediaType = "video"
	case "mp3", "wav", "aac", "m4a", "ogg", "flac", "opus", "wma":
		mediaType = "audio"
	default:
		mediaType = "image"
	}
	common.SysLog(fmt.Sprintf("[detectMediaType] ext=%s, mediaType=%s", ext, mediaType))
	return mediaType
}
