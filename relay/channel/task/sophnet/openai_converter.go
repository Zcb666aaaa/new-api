package sophnet

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/gin-gonic/gin"
)

// ConvertOpenAIImageRequest 将 OpenAI 图片生成请求转换为算能格式
func ConvertOpenAIImageRequest(c *gin.Context, request *dto.ImageRequest) (*CreateTaskRequest, error) {
	sophnetRequest := &CreateTaskRequest{
		Model: request.Model,
		Input: InputObject{
			Prompt: request.Prompt,
		},
		Parameters: &ParametersObject{},
	}

	// 转换 size 参数
	if request.Size != "" {
		sophnetRequest.Parameters.Size = request.Size
	}

	// 转换 watermark 参数
	if request.Watermark != nil {
		sophnetRequest.Parameters.Watermark = request.Watermark
	}

	// 从 Extra 中提取算能特有参数
	if request.Extra != nil {
		if negativePrompt, ok := request.Extra["negative_prompt"]; ok {
			var np string
			if err := unmarshalRawMessage(negativePrompt, &np); err == nil {
				sophnetRequest.Input.NegativePrompt = np
			}
		}

		if seed, ok := request.Extra["seed"]; ok {
			var s int
			if err := unmarshalRawMessage(seed, &s); err == nil {
				sophnetRequest.Parameters.Seed = &s
			}
		}

		if promptExtend, ok := request.Extra["prompt_extend"]; ok {
			var pe bool
			if err := unmarshalRawMessage(promptExtend, &pe); err == nil {
				sophnetRequest.Parameters.PromptExtend = &pe
			}
		}

		if saveToJpeg, ok := request.Extra["save_to_jpeg"]; ok {
			var stj bool
			if err := unmarshalRawMessage(saveToJpeg, &stj); err == nil {
				sophnetRequest.Parameters.SaveToJpeg = &stj
			}
		}

		// 支持图片编辑模型的 images 参数
		if images, ok := request.Extra["images"]; ok {
			var imgs []string
			if err := unmarshalRawMessage(images, &imgs); err == nil {
				sophnetRequest.Input.Images = imgs
			}
		}
	}

	return sophnetRequest, nil
}

// ConvertOpenAIVideoRequest 将 OpenAI 视频生成请求转换为算能格式
func ConvertOpenAIVideoRequest(c *gin.Context, request *dto.VideoRequest) (*CreateVideoTaskRequest, error) {
	sophnetRequest := &CreateVideoTaskRequest{
		Model:      request.Model,
		Parameters: &VideoParametersObject{},
	}

	// 构建 content 数组
	var content []VideoContentObject

	// 添加文本提示
	if request.Prompt != "" {
		content = append(content, VideoContentObject{
			Type: "text",
			Text: request.Prompt,
		})
	}

	// 添加图片输入（用于 I2V 模型）
	if request.Image != "" {
		imageContent := VideoContentObject{
			Type: "image_url",
		}

		// 判断是 URL 还是 base64
		if strings.HasPrefix(request.Image, "http://") || strings.HasPrefix(request.Image, "https://") {
			imageContent.ImageUrl = &VideoImageObject{
				URL: request.Image,
			}
		} else if strings.HasPrefix(request.Image, "data:") {
			// data URL 格式
			imageContent.ImageUrl = &VideoImageObject{
				URL: request.Image,
			}
		} else {
			// 假设是 base64，添加前缀
			imageContent.ImageUrl = &VideoImageObject{
				URL: "data:image/jpeg;base64," + request.Image,
			}
		}

		content = append(content, imageContent)
	}

	if len(content) == 0 {
		return nil, fmt.Errorf("prompt or image is required")
	}

	sophnetRequest.Content = content

	// 转换参数
	if request.Width > 0 && request.Height > 0 {
		sophnetRequest.Parameters.Size = fmt.Sprintf("%dx%d", request.Width, request.Height)
	}

	if request.Duration > 0 {
		duration := int(request.Duration)
		sophnetRequest.Parameters.Duration = &duration
	}

	if request.Seed > 0 {
		sophnetRequest.Parameters.Seed = fmt.Sprintf("%d", request.Seed)
	}

	// 从 Metadata 中提取算能特有参数
	if request.Metadata != nil {
		if negativePrompt, ok := request.Metadata["negative_prompt"].(string); ok {
			// 添加到第一个文本 content 中
			for i := range content {
				if content[i].Type == "text" {
					content[i].NegativePrompt = negativePrompt
					break
				}
			}
		}

		if subdivisionLevel, ok := request.Metadata["subdivision_level"].(string); ok {
			sophnetRequest.Parameters.SubdivisionLevel = subdivisionLevel
		}

		if fileFormat, ok := request.Metadata["file_format"].(string); ok {
			sophnetRequest.Parameters.FileFormat = fileFormat
		}

		if callbackURL, ok := request.Metadata["callback_url"].(string); ok {
			sophnetRequest.CallbackURL = callbackURL
		}

		if returnLastFrame, ok := request.Metadata["return_last_frame"].(bool); ok {
			sophnetRequest.ReturnLastFrame = &returnLastFrame
		}

		if serviceTier, ok := request.Metadata["service_tier"].(string); ok {
			sophnetRequest.ServiceTier = serviceTier
		}

		if generateAudio, ok := request.Metadata["generate_audio"].(bool); ok {
			sophnetRequest.GenerateAudio = &generateAudio
		}

		if draft, ok := request.Metadata["draft"].(bool); ok {
			sophnetRequest.Draft = &draft
		}

		// 支持 execution_expires_after 参数
		if expiresAfter, ok := request.Metadata["execution_expires_after"].(float64); ok {
			expires := int(expiresAfter)
			sophnetRequest.ExecutionExpiresAfter = &expires
		} else if expiresAfter, ok := request.Metadata["execution_expires_after"].(int); ok {
			sophnetRequest.ExecutionExpiresAfter = &expiresAfter
		}

		// 支持 audio_url 参数（添加到 content 中）
		if audioURL, ok := request.Metadata["audio_url"].(string); ok {
			audioContent := VideoContentObject{
				Type:     "text", // 音频 URL 通常附加到文本 content
				AudioURL: audioURL,
			}
			content = append(content, audioContent)
		}

		// 支持 role 参数（用于指定图片角色：first_frame, last_frame, reference_image）
		if role, ok := request.Metadata["role"].(string); ok {
			// 将 role 应用到最后一个 image_url content
			for i := len(content) - 1; i >= 0; i-- {
				if content[i].Type == "image_url" {
					content[i].Role = role
					break
				}
			}
		}

		// 支持 draft_task 参数（用于样片任务）
		if draftTaskID, ok := request.Metadata["draft_task_id"].(string); ok {
			draftContent := VideoContentObject{
				Type: "draft_task",
				DraftTask: &VideoDraftTaskObject{
					ID: draftTaskID,
				},
			}
			content = append(content, draftContent)
		}

		// 支持 subjects 参数
		if subjects, ok := request.Metadata["subjects"].([]interface{}); ok {
			var videoSubjects []VideoSubjectObject
			for _, subj := range subjects {
				if subjMap, ok := subj.(map[string]interface{}); ok {
					vs := VideoSubjectObject{}
					if id, ok := subjMap["id"].(string); ok {
						vs.ID = id
					}
					if images, ok := subjMap["images"].([]interface{}); ok {
						for _, img := range images {
							if imgStr, ok := img.(string); ok {
								vs.Images = append(vs.Images, imgStr)
							}
						}
					}
					if voiceID, ok := subjMap["voice_id"].(string); ok {
						vs.VoiceID = voiceID
					}
					videoSubjects = append(videoSubjects, vs)
				}
			}
			if len(videoSubjects) > 0 {
				sophnetRequest.Subjects = videoSubjects
			}
		}
	}

	// 更新 content（可能在 metadata 处理中添加了新的 content）
	sophnetRequest.Content = content

	return sophnetRequest, nil
}

// unmarshalRawMessage 辅助函数，用于解析 json.RawMessage
func unmarshalRawMessage(raw interface{}, v interface{}) error {
	if raw == nil {
		return fmt.Errorf("raw is nil")
	}
	
	// 如果是 json.RawMessage，先解析
	if rawMsg, ok := raw.(json.RawMessage); ok {
		return common.Unmarshal(rawMsg, v)
	}
	
	// 如果已经是目标类型，直接赋值
	switch target := v.(type) {
	case *string:
		if str, ok := raw.(string); ok {
			*target = str
			return nil
		}
	case *int:
		if num, ok := raw.(float64); ok {
			*target = int(num)
			return nil
		}
		if num, ok := raw.(int); ok {
			*target = num
			return nil
		}
	case *bool:
		if b, ok := raw.(bool); ok {
			*target = b
			return nil
		}
	case *[]string:
		if arr, ok := raw.([]interface{}); ok {
			var strs []string
			for _, item := range arr {
				if str, ok := item.(string); ok {
					strs = append(strs, str)
				}
			}
			*target = strs
			return nil
		}
	}
	
	return fmt.Errorf("unsupported type conversion")
}
