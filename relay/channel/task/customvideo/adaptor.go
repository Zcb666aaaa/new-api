package customvideo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay/channel"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// ============================
// Request / Response structures
// ============================

// CustomVideoRequest 自定义视频请求（透传格式）
type CustomVideoRequest struct {
	Model    string         `json:"model"`
	Prompt   string         `json:"prompt"`
	Size     string         `json:"size,omitempty"`
	Duration *int           `json:"duration,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// CustomVideoResponse 自定义视频响应
type CustomVideoResponse struct {
	ID        string         `json:"id"`
	Status    string         `json:"status"`
	Message   string         `json:"message,omitempty"`
	VideoURL  string         `json:"video_url,omitempty"`
	Progress  string         `json:"progress,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt int64          `json:"created_at,omitempty"`
}

// ============================
// Adaptor implementation
// ============================

type TaskAdaptor struct {
	ChannelType int
	apiKey      string
	baseURL     string
}

func (a *TaskAdaptor) Init(info *relaycommon.RelayInfo) {
	a.ChannelType = info.ChannelType
	a.baseURL = info.ChannelBaseUrl
	a.apiKey = info.ApiKey
}

func (a *TaskAdaptor) ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.RelayInfo) (taskErr *dto.TaskError) {
	return relaycommon.ValidateMultipartDirect(c, info)
}

func (a *TaskAdaptor) BuildRequestURL(info *relaycommon.RelayInfo) (string, error) {
	// 直接使用 baseURL + /v1/videos/create
	baseURL := strings.TrimRight(a.baseURL, "/")
	return fmt.Sprintf("%s", baseURL), nil
}

func (a *TaskAdaptor) BuildRequestHeader(c *gin.Context, req *http.Request, info *relaycommon.RelayInfo) error {
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")
	return nil
}

func (a *TaskAdaptor) BuildRequestBody(c *gin.Context, info *relaycommon.RelayInfo) (io.Reader, error) {
	// 自定义视频渠道：完全透传原始请求体，不做任何字段过滤
	bodyStorage, err := common.GetBodyStorage(c)
	if err != nil {
		return nil, errors.Wrap(err, "get_body_storage_failed")
	}

	// 读取原始请求体
	bodyBytes, err := io.ReadAll(bodyStorage)
	if err != nil {
		return nil, errors.Wrap(err, "read_body_failed")
	}

	// 记录日志（可选）
	var logBody map[string]interface{}
	if err := common.Unmarshal(bodyBytes, &logBody); err == nil {
		logger.LogJson(c, "custom video request body (passthrough)", logBody)
	}

	return bytes.NewReader(bodyBytes), nil
}

// EstimateBilling 根据用户请求参数计算 OtherRatios
func (a *TaskAdaptor) EstimateBilling(c *gin.Context, info *relaycommon.RelayInfo) map[string]float64 {
	taskReq, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return nil
	}

	otherRatios := make(map[string]float64)
	
	// 根据时长计费
	var duration float64
	if taskReq.Duration > 0 {
		duration = float64(taskReq.Duration)
		otherRatios["seconds"] = duration
	} else {
		duration = 5.0 // 默认5秒
		otherRatios["seconds"] = duration
	}

	logger.LogInfo(c, fmt.Sprintf("CustomVideo EstimateBilling - duration: %.1f seconds, otherRatios: %+v", duration, otherRatios))

	return otherRatios
}

// AdjustBillingOnSubmit 提交后调整计费（可选）
func (a *TaskAdaptor) AdjustBillingOnSubmit(info *relaycommon.RelayInfo, taskData []byte) map[string]float64 {
	return nil
}

// AdjustBillingOnComplete 任务完成后调整计费（可选）
func (a *TaskAdaptor) AdjustBillingOnComplete(task *model.Task, taskResult *relaycommon.TaskInfo) int {
	return 0
}

// DoRequest delegates to common helper
func (a *TaskAdaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	return channel.DoTaskApiRequest(a, c, info, requestBody)
}

// DoResponse handles upstream response
func (a *TaskAdaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (taskID string, taskData []byte, taskErr *dto.TaskError) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
		return
	}
	_ = resp.Body.Close()

	// 解析自定义响应
	var customResp CustomVideoResponse
	if err := common.Unmarshal(responseBody, &customResp); err != nil {
		taskErr = service.TaskErrorWrapper(errors.Wrapf(err, "body: %s", responseBody), "unmarshal_response_body_failed", http.StatusInternalServerError)
		return
	}

	// 检查错误
	if customResp.ID == "" {
		taskErr = service.TaskErrorWrapper(fmt.Errorf("task_id is empty"), "invalid_response", http.StatusInternalServerError)
		return
	}

	// 转换为 OpenAI 格式响应
	openAIResp := dto.NewOpenAIVideo()
	openAIResp.ID = info.PublicTaskID
	openAIResp.TaskID = info.PublicTaskID
	openAIResp.Model = c.GetString("model")
	if openAIResp.Model == "" && info != nil {
		openAIResp.Model = info.OriginModelName
	}
	openAIResp.Status = convertCustomStatus(customResp.Status)
	openAIResp.CreatedAt = common.GetTimestamp()

	// 返回 OpenAI 格式
	c.JSON(http.StatusOK, openAIResp)

	return customResp.ID, responseBody, nil
}

// FetchTask 查询任务状态
func (a *TaskAdaptor) FetchTask(baseUrl, key string, body map[string]any, proxy string) (*http.Response, error) {
	taskID, ok := body["task_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid task_id")
	}

	// 转发地址直接采用 baseURL/{id}
	baseUrl = strings.TrimRight(baseUrl, "/")
	uri := fmt.Sprintf("%s/%s", baseUrl, taskID)

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

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

// ParseTaskResult 解析任务结果
func (a *TaskAdaptor) ParseTaskResult(respBody []byte) (*relaycommon.TaskInfo, error) {
	logger.LogInfo(context.Background(), fmt.Sprintf("CustomVideo ParseTaskResult - respBody: %s", string(respBody)))
	
	var customResp CustomVideoResponse
	if err := common.Unmarshal(respBody, &customResp); err != nil {
		return nil, errors.Wrap(err, "unmarshal task result failed")
	}

	logger.LogInfo(context.Background(), fmt.Sprintf("CustomVideo ParseTaskResult - customResp: %+v", customResp))

	taskResult := relaycommon.TaskInfo{
		Code: 0,
	}

	// 状态映射
	switch strings.ToLower(customResp.Status) {
	case "queued", "pending":
		taskResult.Status = model.TaskStatusQueued
		taskResult.Progress = "10%"
	case "processing", "running", "in_progress":
		taskResult.Status = model.TaskStatusInProgress
		if customResp.Progress != "" {
			taskResult.Progress = customResp.Progress
		} else {
			taskResult.Progress = "50%"
		}
	case "completed", "succeeded", "success":
		taskResult.Status = model.TaskStatusSuccess
		taskResult.Progress = "100%"
		taskResult.Url = customResp.VideoURL
		logger.LogInfo(context.Background(), fmt.Sprintf("CustomVideo ParseTaskResult - Setting taskResult.Url to: %s", customResp.VideoURL))
	case "failed", "error", "canceled", "cancelled":
		taskResult.Status = model.TaskStatusFailure
		taskResult.Progress = "100%"
		if customResp.Message != "" {
			taskResult.Reason = customResp.Message
		} else {
			taskResult.Reason = "task failed"
		}
	default:
		taskResult.Status = model.TaskStatusQueued
		taskResult.Progress = "10%"
	}

	logger.LogInfo(context.Background(), fmt.Sprintf("CustomVideo ParseTaskResult - taskResult: %+v", taskResult))

	return &taskResult, nil
}

func (a *TaskAdaptor) ConvertToOpenAIVideo(task *model.Task) ([]byte, error) {
	var customResp CustomVideoResponse
	if err := common.Unmarshal(task.Data, &customResp); err != nil {
		return nil, errors.Wrap(err, "unmarshal custom response failed")
	}

	logger.LogInfo(context.Background(), fmt.Sprintf("CustomVideo ConvertToOpenAIVideo - customResp: %+v", customResp))

	openAIResp := dto.NewOpenAIVideo()
	openAIResp.ID = task.TaskID
	openAIResp.Status = convertCustomStatus(customResp.Status)
	openAIResp.Model = task.Properties.OriginModelName
	openAIResp.SetProgressStr(task.Progress)
	openAIResp.CreatedAt = task.CreatedAt
	openAIResp.CompletedAt = task.UpdatedAt

	// 设置视频URL - 直接使用上游返回的原始URL
	if customResp.VideoURL != "" {
		openAIResp.SetMetadata("url", customResp.VideoURL)
		logger.LogInfo(context.Background(), fmt.Sprintf("CustomVideo ConvertToOpenAIVideo - using upstream VideoURL: %s", customResp.VideoURL))
	}

	// 错误处理
	if customResp.Message != "" && (customResp.Status == "failed" || customResp.Status == "error") {
		openAIResp.Error = &dto.OpenAIVideoError{
			Code:    "custom_video_error",
			Message: customResp.Message,
		}
	}

	return common.Marshal(openAIResp)
}

func convertCustomStatus(customStatus string) string {
	switch strings.ToLower(customStatus) {
	case "queued", "pending":
		return dto.VideoStatusQueued
	case "processing", "running", "in_progress":
		return dto.VideoStatusInProgress
	case "completed", "succeeded", "success":
		return dto.VideoStatusCompleted
	case "failed", "error", "canceled", "cancelled":
		return dto.VideoStatusFailed
	default:
		return dto.VideoStatusUnknown
	}
}
