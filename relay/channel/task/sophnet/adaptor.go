package sophnet

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
)

type TaskAdaptor struct {
	ChannelType int
	IsVideo     bool // 标识是否为视频任务
}

func (a *TaskAdaptor) Init(info *relaycommon.RelayInfo) {
	a.ChannelType = info.ChannelType
}

// isVideoModel 判断是否为视频模型
func isVideoModel(model string) bool {
	for _, m := range UpstreamVideoModels {
		if m == model {
			return true
		}
	}
	return false
}

// ValidateRequestAndSetAction 验证请求并设置 Action
func (a *TaskAdaptor) ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.RelayInfo) (taskErr *dto.TaskError) {
	action := c.Param("action")
	if action == "" {
		action = "generate" // 默认为生成动作
	}

	// 判断是否为 OpenAI 格式的请求
	isOpenAIImageRequest := strings.HasPrefix(c.Request.URL.Path, "/v1/images/generations")
	isOpenAIVideoRequest := strings.HasPrefix(c.Request.URL.Path, "/v1/videos") || 
		strings.HasPrefix(c.Request.URL.Path, "/v1/video/generations")

	if isOpenAIImageRequest {
		a.IsVideo = false
		return a.validateOpenAIImageRequest(c, info, action)
	} else if isOpenAIVideoRequest {
		a.IsVideo = true
		return a.validateOpenAIVideoRequest(c, info, action)
	}

	// 根据请求路径判断是图片还是视频（算能原生格式）
	if strings.Contains(c.Request.URL.Path, "/videogenerator/") {
		a.IsVideo = true
		return a.validateVideoRequest(c, info, action)
	} else {
		a.IsVideo = false
		return a.validateImageRequest(c, info, action)
	}
}

// validateImageRequest 验证图片请求
func (a *TaskAdaptor) validateImageRequest(c *gin.Context, info *relaycommon.RelayInfo, action string) (taskErr *dto.TaskError) {
	var request *CreateTaskRequest
	err := common.UnmarshalBodyReusable(c, &request)
	if err != nil {
		taskErr = service.TaskErrorWrapperLocal(err, "invalid_request", http.StatusBadRequest)
		return
	}

	// 验证必填参数
	if request.Model == "" {
		taskErr = service.TaskErrorWrapperLocal(fmt.Errorf("model is required"), "invalid_request", http.StatusBadRequest)
		return
	}

	if request.Input.Prompt == "" {
		taskErr = service.TaskErrorWrapperLocal(fmt.Errorf("input.prompt is required"), "invalid_request", http.StatusBadRequest)
		return
	}

	// 图编辑模型需要输入图片
	if request.Model == "Qwen-Image-Edit-2509" && len(request.Input.Images) == 0 {
		taskErr = service.TaskErrorWrapperLocal(fmt.Errorf("input.images is required for Qwen-Image-Edit-2509"), "invalid_request", http.StatusBadRequest)
		return
	}

	info.Action = action
	info.UpstreamModelName = request.Model
	c.Set("task_request", request)
	return nil
}

// validateVideoRequest 验证视频请求
func (a *TaskAdaptor) validateVideoRequest(c *gin.Context, info *relaycommon.RelayInfo, action string) (taskErr *dto.TaskError) {
	var request *CreateVideoTaskRequest
	err := common.UnmarshalBodyReusable(c, &request)
	if err != nil {
		taskErr = service.TaskErrorWrapperLocal(err, "invalid_request", http.StatusBadRequest)
		return
	}

	// 验证必填参数
	if request.Model == "" {
		taskErr = service.TaskErrorWrapperLocal(fmt.Errorf("model is required"), "invalid_request", http.StatusBadRequest)
		return
	}

	if len(request.Content) == 0 {
		taskErr = service.TaskErrorWrapperLocal(fmt.Errorf("content is required"), "invalid_request", http.StatusBadRequest)
		return
	}

	info.Action = action
	info.UpstreamModelName = request.Model
	c.Set("task_request", request)
	return nil
}

// validateOpenAIImageRequest 验证 OpenAI 格式的图片请求
func (a *TaskAdaptor) validateOpenAIImageRequest(c *gin.Context, info *relaycommon.RelayInfo, action string) (taskErr *dto.TaskError) {
	var openaiRequest dto.ImageRequest
	err := common.UnmarshalBodyReusable(c, &openaiRequest)
	if err != nil {
		taskErr = service.TaskErrorWrapperLocal(err, "invalid_request", http.StatusBadRequest)
		return
	}

	// 验证必填参数
	if openaiRequest.Model == "" {
		taskErr = service.TaskErrorWrapperLocal(fmt.Errorf("model is required"), "invalid_request", http.StatusBadRequest)
		return
	}

	if openaiRequest.Prompt == "" {
		taskErr = service.TaskErrorWrapperLocal(fmt.Errorf("prompt is required"), "invalid_request", http.StatusBadRequest)
		return
	}

	// 转换为算能格式
	sophnetRequest, err := ConvertOpenAIImageRequest(c, &openaiRequest)
	if err != nil {
		taskErr = service.TaskErrorWrapperLocal(err, "conversion_failed", http.StatusBadRequest)
		return
	}

	info.Action = action
	info.UpstreamModelName = sophnetRequest.Model
	c.Set("task_request", sophnetRequest)
	c.Set("openai_format", true) // 标记为 OpenAI 格式
	return nil
}

// validateOpenAIVideoRequest 验证 OpenAI 格式的视频请求
func (a *TaskAdaptor) validateOpenAIVideoRequest(c *gin.Context, info *relaycommon.RelayInfo, action string) (taskErr *dto.TaskError) {
	var openaiRequest dto.VideoRequest
	err := common.UnmarshalBodyReusable(c, &openaiRequest)
	if err != nil {
		taskErr = service.TaskErrorWrapperLocal(err, "invalid_request", http.StatusBadRequest)
		return
	}

	// 验证必填参数
	if openaiRequest.Model == "" {
		taskErr = service.TaskErrorWrapperLocal(fmt.Errorf("model is required"), "invalid_request", http.StatusBadRequest)
		return
	}

	// 转换为算能格式
	sophnetRequest, err := ConvertOpenAIVideoRequest(c, &openaiRequest)
	if err != nil {
		taskErr = service.TaskErrorWrapperLocal(err, "conversion_failed", http.StatusBadRequest)
		return
	}

	info.Action = action
	info.UpstreamModelName = sophnetRequest.Model
	c.Set("task_request", sophnetRequest)
	c.Set("openai_format", true) // 标记为 OpenAI 格式
	return nil
}

// BuildRequestURL 构建上游请求 URL
func (a *TaskAdaptor) BuildRequestURL(info *relaycommon.RelayInfo) (string, error) {
	baseURL := info.ChannelBaseUrl
	if a.IsVideo {
		return fmt.Sprintf("%s/api/open-apis/projects/easyllms/videogenerator/generate", baseURL), nil
	}
	return fmt.Sprintf("%s/api/open-apis/projects/easyllms/imagegenerator/task", baseURL), nil
}

// BuildRequestHeader 构建请求头
func (a *TaskAdaptor) BuildRequestHeader(c *gin.Context, req *http.Request, info *relaycommon.RelayInfo) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+info.ApiKey)
	return nil
}

// BuildRequestBody 构建请求体
func (a *TaskAdaptor) BuildRequestBody(c *gin.Context, info *relaycommon.RelayInfo) (io.Reader, error) {
	request, ok := c.Get("task_request")
	if !ok {
		return nil, fmt.Errorf("task_request not found in context")
	}

	// 应用模型映射：如果模型被重定向，使用映射后的模型名
	if info.UpstreamModelName != "" {
		if a.IsVideo {
			if videoReq, ok := request.(*CreateVideoTaskRequest); ok {
				videoReq.Model = info.UpstreamModelName
			}
		} else {
			if imageReq, ok := request.(*CreateTaskRequest); ok {
				imageReq.Model = info.UpstreamModelName
			}
		}
	}

	data, err := common.Marshal(request)
	if err != nil {
		return nil, err
	}
	
	// 打印算能渠道转发的请求体
	logger.LogInfo(c, fmt.Sprintf("算能渠道转发请求体: %s", string(data)))
	
	return bytes.NewReader(data), nil
}

// DoRequest 发送请求
func (a *TaskAdaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	return channel.DoTaskApiRequest(a, c, info, requestBody)
}

// DoResponse 处理创建任务的响应
func (a *TaskAdaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (taskID string, taskData []byte, taskErr *dto.TaskError) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
		return
	}

	if a.IsVideo {
		return a.doVideoResponse(c, responseBody, info)
	}
	return a.doImageResponse(c, responseBody, info)
}

// doImageResponse 处理图片任务响应
func (a *TaskAdaptor) doImageResponse(c *gin.Context, responseBody []byte, info *relaycommon.RelayInfo) (taskID string, taskData []byte, taskErr *dto.TaskError) {
	var sophnetResponse CreateTaskResponse
	err := common.Unmarshal(responseBody, &sophnetResponse)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError)
		return
	}

	// 检查是否有错误
	if sophnetResponse.Code != "" {
		taskErr = service.TaskErrorWrapper(
			fmt.Errorf("%s: %s", sophnetResponse.Code, sophnetResponse.Message),
			sophnetResponse.Code,
			http.StatusInternalServerError,
		)
		return
	}

	if sophnetResponse.Output.TaskId == "" {
		taskErr = service.TaskErrorWrapper(fmt.Errorf("taskId is empty"), "empty_task_id", http.StatusInternalServerError)
		return
	}

	// 只返回 task_id
	c.JSON(http.StatusOK, gin.H{"task_id": info.PublicTaskID})

	// 返回上游真实 task ID 用于后续查询
	return sophnetResponse.Output.TaskId, nil, nil
}

// doVideoResponse 处理视频任务响应
func (a *TaskAdaptor) doVideoResponse(c *gin.Context, responseBody []byte, info *relaycommon.RelayInfo) (taskID string, taskData []byte, taskErr *dto.TaskError) {
	var sophnetResponse CreateVideoTaskResponse
	err := common.Unmarshal(responseBody, &sophnetResponse)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError)
		return
	}

	// 检查是否有错误
	if sophnetResponse.Status != 0 {
		taskErr = service.TaskErrorWrapper(
			fmt.Errorf("status %d: %s", sophnetResponse.Status, sophnetResponse.Message),
			"video_task_failed",
			http.StatusInternalServerError,
		)
		return
	}

	if sophnetResponse.Result.TaskID == "" {
		taskErr = service.TaskErrorWrapper(fmt.Errorf("taskId is empty"), "empty_task_id", http.StatusInternalServerError)
		return
	}

	// 只返回 task_id
	c.JSON(http.StatusOK, gin.H{"task_id": info.PublicTaskID})

	// 返回上游真实 task ID 用于后续查询
	return sophnetResponse.Result.TaskID, nil, nil
}

// FetchTask 轮询查询任务状态
func (a *TaskAdaptor) FetchTask(baseUrl, key string, body map[string]any, proxy string) (*http.Response, error) {
	taskId, ok := body["task_id"].(string)
	if !ok {
		return nil, fmt.Errorf("task_id not found in body")
	}

	var requestUrl string
	
	// 从 body 中获取模型名称来判断任务类型
	model, _ := body["model"].(string)
	isVideo := isVideoModel(model)
	
	if isVideo {
		requestUrl = fmt.Sprintf("%s/api/open-apis/projects/easyllms/videogenerator/generate/%s", baseUrl, taskId)
	} else {
		requestUrl = fmt.Sprintf("%s/api/open-apis/projects/easyllms/imagegenerator/task/%s", baseUrl, taskId)
	}

	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+key)

	client, err := service.GetHttpClientWithProxy(proxy)
	if err != nil {
		return nil, fmt.Errorf("new proxy http client failed: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// ParseTaskResult 解析任务查询结果
func (a *TaskAdaptor) ParseTaskResult(body []byte) (*relaycommon.TaskInfo, error) {
	// 先尝试解析为图片响应（检查是否有 requestId 字段）
	var imageResult QueryTaskResponse
	if err := common.Unmarshal(body, &imageResult); err == nil && imageResult.RequestId != "" {
		logger.LogInfo(context.Background(), fmt.Sprintf("Sophnet ParseTaskResult: detected as image response, requestId=%s, full body=%s", imageResult.RequestId, string(body)))
		return a.parseImageTaskResult(&imageResult)
	}

	// 尝试解析为视频响应（检查是否有 result.id 字段）
	var videoResult QueryVideoTaskResponse
	if err := common.Unmarshal(body, &videoResult); err == nil && videoResult.Result.ID != "" {
		logger.LogInfo(context.Background(), fmt.Sprintf("Sophnet ParseTaskResult: detected as video response, taskId=%s, full body=%s", videoResult.Result.ID, string(body)))
		return a.parseVideoTaskResult(&videoResult)
	}

	// 检查是否为视频错误响应（status 字段存在且不为 0）
	// 视频错误响应的 result 可能是 null、空对象或包含部分信息
	var errorResp SophnetErrorResponse
	if err := common.Unmarshal(body, &errorResp); err == nil && errorResp.Status != 0 {
		// 检查 result 是否为空或无效（nil、空对象、或没有有效 ID）
		isEmptyResult := errorResp.Result == nil
		if !isEmptyResult {
			// 尝试检查 result 是否为空对象或无效的视频结果
			if resultMap, ok := errorResp.Result.(map[string]any); ok {
				// 如果是 map 但没有 id 字段或 id 为空，视为错误响应
				if id, exists := resultMap["id"]; !exists || id == nil || id == "" {
					isEmptyResult = true
				}
			}
		}
		
		if isEmptyResult {
			logger.LogError(context.Background(), fmt.Sprintf("Sophnet ParseTaskResult: error response, status=%d, message=%s, full body=%s", errorResp.Status, errorResp.Message, string(body)))
			return &relaycommon.TaskInfo{
				Status: string(model.TaskStatusFailure),
				Reason: fmt.Sprintf("status %d: %s", errorResp.Status, errorResp.Message),
			}, nil
		}
	}

	// 都不匹配，返回错误
	logger.LogError(context.Background(), fmt.Sprintf("Sophnet ParseTaskResult: unable to parse, body=%s", string(body)))
	return nil, fmt.Errorf("unable to parse task result")
}

// parseImageTaskResult 解析图片任务结果
func (a *TaskAdaptor) parseImageTaskResult(result *QueryTaskResponse) (*relaycommon.TaskInfo, error) {
	// 先检查顶层是否有错误码（某些失败场景下错误信息在顶层）
	if result.Code != "" {
		logger.LogInfo(context.Background(), fmt.Sprintf("Sophnet parseImageTaskResult: top-level error, code=%s, message=%s", result.Code, result.Message))
		return &relaycommon.TaskInfo{
			Status:   string(model.TaskStatusFailure),
			Progress: "100%",
			Reason:   fmt.Sprintf("%s: %s", result.Code, result.Message),
		}, nil
	}

	convertedStatus := convertStatus(result.Output.TaskStatus)
	logger.LogInfo(context.Background(), fmt.Sprintf("Sophnet parseImageTaskResult: original=%s, converted=%s", result.Output.TaskStatus, convertedStatus))
	
	taskInfo := &relaycommon.TaskInfo{
		Status:   convertedStatus,
		Progress: calculateProgress(result.Output.TaskStatus),
	}

	// 处理失败情况
	if result.Output.TaskStatus == StatusFailed {
		// 优先使用 Output 中的错误信息
		if result.Output.Message != "" {
			taskInfo.Reason = result.Output.Message
		} else if result.Output.Code != "" {
			taskInfo.Reason = fmt.Sprintf("%s: %s", result.Output.Code, result.Output.Message)
		} else if result.Message != "" {
			// 回退到顶层 Message
			taskInfo.Reason = result.Message
		} else {
			taskInfo.Reason = "任务失败"
		}
	}

	// 处理成功情况 - 提取图片 URL（仅取第一张作为 Url 字段）
	if result.Output.TaskStatus == StatusSucceeded && len(result.Output.Results) > 0 {
		taskInfo.Url = result.Output.Results[0].URL
	}

	return taskInfo, nil
}

// parseVideoTaskResult 解析视频任务结果
func (a *TaskAdaptor) parseVideoTaskResult(result *QueryVideoTaskResponse) (*relaycommon.TaskInfo, error) {
	// 先检查顶层 status 是否表示错误
	if result.Status != 0 {
		logger.LogInfo(context.Background(), fmt.Sprintf("Sophnet parseVideoTaskResult: top-level error, status=%d, message=%s", result.Status, result.Message))
		return &relaycommon.TaskInfo{
			Status:   string(model.TaskStatusFailure),
			Progress: "100%",
			Reason:   fmt.Sprintf("status %d: %s", result.Status, result.Message),
		}, nil
	}

	convertedStatus := convertVideoStatus(result.Result.Status)
	logger.LogInfo(context.Background(), fmt.Sprintf("Sophnet parseVideoTaskResult: original=%s, converted=%s", result.Result.Status, convertedStatus))
	
	taskInfo := &relaycommon.TaskInfo{
		Status:   convertedStatus,
		Progress: calculateVideoProgress(result.Result.Status),
	}

	// 提取时间信息（算能视频接口返回毫秒时间戳，转换为秒）
	if result.Result.CreatedAt > 0 {
		// 判断是毫秒还是秒（如果大于 10^12 则为毫秒）
		if result.Result.CreatedAt > 1e12 {
			taskInfo.StartTime = result.Result.CreatedAt / 1000
		} else {
			taskInfo.StartTime = result.Result.CreatedAt
		}
	}
	if result.Result.UpdatedAt > 0 {
		// 判断是毫秒还是秒
		if result.Result.UpdatedAt > 1e12 {
			taskInfo.FinishTime = result.Result.UpdatedAt / 1000
		} else {
			taskInfo.FinishTime = result.Result.UpdatedAt
		}
	}

	// 处理失败情况
	if result.Result.Status == VideoStatusFailed {
		// 优先使用 result.error 中的错误信息
		if result.Result.Error != nil && result.Result.Error.Message != "" {
			if result.Result.Error.Code != "" {
				taskInfo.Reason = fmt.Sprintf("%s: %s", result.Result.Error.Code, result.Result.Error.Message)
			} else {
				taskInfo.Reason = result.Result.Error.Message
			}
		} else if result.Message != "" && result.Message != "请求成功" {
			// 回退到顶层 message（但排除"请求成功"这种非错误信息）
			taskInfo.Reason = result.Message
		} else {
			taskInfo.Reason = "视频生成失败"
		}
	}

	// 处理成功情况 - 提取视频 URL 和 Usage 信息
	if result.Result.Status == VideoStatusSucceeded && result.Result.Content != nil {
		if result.Result.Content.VideoURL != "" {
			taskInfo.Url = result.Result.Content.VideoURL
		} else if result.Result.Content.FileURL != "" {
			taskInfo.Url = result.Result.Content.FileURL
		}

		// 提取 Usage 信息用于计费（支持 token 和时长两种模式）
		if result.Result.Usage != nil {
			if result.Result.Usage.CompletionTokens != nil {
				taskInfo.CompletionTokens = *result.Result.Usage.CompletionTokens
			}
			// TotalTokens 通常等于 CompletionTokens（视频生成没有 prompt tokens）
			if taskInfo.CompletionTokens > 0 {
				taskInfo.TotalTokens = taskInfo.CompletionTokens
			}
		}
	}

	return taskInfo, nil
}

// GetModelList 获取支持的模型列表
func (a *TaskAdaptor) GetModelList() []string {
	return ModelList
}

// GetChannelName 获取渠道名称
func (a *TaskAdaptor) GetChannelName() string {
	return ChannelName
}

// EstimateBilling 根据请求参数预估费用（视频任务需要根据时长计费）
func (a *TaskAdaptor) EstimateBilling(c *gin.Context, info *relaycommon.RelayInfo) map[string]float64 {
	// 图片任务不需要额外倍率
	if !a.IsVideo {
		return nil
	}

	// 视频任务：根据时长计费
	request, ok := c.Get("task_request")
	if !ok {
		return nil
	}

	videoReq, ok := request.(*CreateVideoTaskRequest)
	if !ok {
		return nil
	}

	// 提取时长参数（默认 5 秒）
	duration := 5
	if videoReq.Parameters != nil && videoReq.Parameters.Duration != nil {
		duration = *videoReq.Parameters.Duration
	}

	return map[string]float64{
		"seconds": float64(duration),
	}
}

// AdjustBillingOnSubmit 提交后不需要调整（算能返回的响应中没有实际参数）
func (a *TaskAdaptor) AdjustBillingOnSubmit(_ *relaycommon.RelayInfo, _ []byte) map[string]float64 {
	return nil
}

// AdjustBillingOnComplete 任务完成后根据实际消耗进行差额结算
// 支持两种计费模式：
// 1. 按 token 计费：如果 taskResult.TotalTokens > 0，返回 0 让系统自动按 token 重算
// 2. 按时长计费：如果没有 token 但有 duration，手动计算实际费用
func (a *TaskAdaptor) AdjustBillingOnComplete(task *model.Task, taskResult *relaycommon.TaskInfo) int {
	// 只有成功的视频任务需要根据实际消耗重新计费
	if taskResult.Status != string(model.TaskStatusSuccess) {
		return 0
	}

	// 检查是否为视频任务
	if task.Properties.UpstreamModelName == "" {
		return 0
	}
	isVideo := isVideoModel(task.Properties.UpstreamModelName)
	if !isVideo {
		return 0
	}

	// 优先使用 token 计费：如果有 TotalTokens，返回 0 让系统自动按 token 重算
	if taskResult.TotalTokens > 0 {
		logger.LogInfo(context.Background(), fmt.Sprintf(
			"算能视频任务 %s 使用 token 计费，TotalTokens=%d，由系统自动重算",
			task.TaskID, taskResult.TotalTokens,
		))
		return 0
	}

	// 如果没有 token，尝试按时长计费
	var videoResult QueryVideoTaskResponse
	if err := common.Unmarshal(task.Data, &videoResult); err != nil {
		logger.LogError(context.Background(), fmt.Sprintf("解析视频任务数据失败: %s", err.Error()))
		return 0
	}

	// 如果有 Usage 信息，使用实际时长重新计算
	if videoResult.Result.Usage != nil && videoResult.Result.Usage.Duration != nil {
		actualDuration := *videoResult.Result.Usage.Duration
		if actualDuration <= 0 {
			return 0
		}

		// 从 BillingContext 获取计费参数
		bc := task.PrivateData.BillingContext
		if bc == nil {
			return 0
		}

		// 重新计算实际费用：modelPrice * groupRatio * modelRatio * actualDuration
		actualQuota := int(bc.ModelPrice * bc.GroupRatio * bc.ModelRatio * float64(actualDuration))
		if actualQuota <= 0 {
			actualQuota = 1
		}

		logger.LogInfo(context.Background(), fmt.Sprintf(
			"算能视频任务 %s 使用时长计费，实际时长 %d 秒，重新计算费用: %d",
			task.TaskID, actualDuration, actualQuota,
		))

		return actualQuota
	}

	// 没有 Usage 信息，保持预扣费不变
	logger.LogInfo(context.Background(), fmt.Sprintf(
		"算能视频任务 %s 无 Usage 信息，保持预扣费不变",
		task.TaskID,
	))
	return 0
}

// convertStatus 将算能云图片状态转换为标准状态
func convertStatus(sophnetStatus string) string {
	switch sophnetStatus {
	case StatusPending:
		return string(model.TaskStatusQueued)
	case StatusRunning:
		return string(model.TaskStatusInProgress)
	case StatusSucceeded:
		return string(model.TaskStatusSuccess)
	case StatusFailed, StatusCanceled:
		return string(model.TaskStatusFailure)
	case "":
		return string(model.TaskStatusQueued) // 空状态视为排队中
	default:
		// 未知状态记录日志并返回 UNKNOWN，避免持续轮询
		logger.LogError(context.Background(), fmt.Sprintf("Sophnet unknown image status: %s", sophnetStatus))
		return string(model.TaskStatusUnknown)
	}
}

// convertVideoStatus 将算能云视频状态转换为标准状态
func convertVideoStatus(sophnetStatus string) string {
	switch sophnetStatus {
	case VideoStatusQueued:
		return string(model.TaskStatusQueued)
	case VideoStatusRunning:
		return string(model.TaskStatusInProgress)
	case VideoStatusSucceeded:
		return string(model.TaskStatusSuccess)
	case VideoStatusFailed, VideoStatusCancelled:
		return string(model.TaskStatusFailure)
	case "":
		return string(model.TaskStatusQueued) // 空状态视为排队中
	default:
		// 未知状态记录日志并返回 UNKNOWN，避免持续轮询
		logger.LogError(context.Background(), fmt.Sprintf("Sophnet unknown video status: %s", sophnetStatus))
		return string(model.TaskStatusUnknown)
	}
}

// calculateProgress 根据图片状态计算进度
func calculateProgress(status string) string {
	switch status {
	case StatusPending:
		return "0%"
	case StatusRunning:
		return "50%"
	case StatusSucceeded, StatusFailed, StatusCanceled:
		return "100%"
	default:
		return "0%"
	}
}

// calculateVideoProgress 根据视频状态计算进度
func calculateVideoProgress(status string) string {
	switch status {
	case VideoStatusQueued:
		return "0%"
	case VideoStatusRunning:
		return "50%"
	case VideoStatusSucceeded, VideoStatusFailed, VideoStatusCancelled:
		return "100%"
	default:
		return "0%"
	}
}
