package api

import "encoding/json"

// ModelsResponse 模型列表响应
type ModelsResponse struct {
	Object string       `json:"object"`
	Data   []ModelInfo  `json:"data"`
}

// ModelInfo 模型信息
type ModelInfo struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// ChatCompletionRequest OpenAI 兼容的聊天请求
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature *float64      `json:"temperature,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"`
}

// ChatCompletionResponse OpenAI 兼容的聊天响应
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// ChatChoice 聊天选择
type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// Usage 使用量统计
type Usage struct {
	PromptTokens     int32 `json:"prompt_tokens"`
	CompletionTokens int32 `json:"completion_tokens"`
	TotalTokens      int32 `json:"total_tokens"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    int    `json:"code"` // HTTP 状态码
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(message string, errorType string, code int) *ErrorResponse {
	return &ErrorResponse{
		Error: ErrorDetail{
			Message: message,
			Type:    errorType,
			Code:    code,
		},
	}
}

// ToJSON 转换为 JSON
func (e *ErrorResponse) ToJSON() []byte {
	json, _ := json.Marshal(e)
	return json
}

// ========== Images API Types ==========

// ImageRequest 图片生成请求
type ImageRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	N              int    `json:"n,omitempty"`               // Number of images
	Size           string `json:"size,omitempty"`            // 256x256, 512x512, 1024x1024
	ResponseFormat string `json:"response_format,omitempty"` // url, b64_json
	Quality        string `json:"quality,omitempty"`         // standard, hd
	Style          string `json:"style,omitempty"`           // vivid, natural
}

// ImageResponse 图片生成响应
type ImageResponse struct {
	Created int64      `json:"created"`
	Data    []ImageItem `json:"data"`
}

// ImageItem 图片项
type ImageItem struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// ImageEditRequest 图片编辑请求
type ImageEditRequest struct {
	Model          string `json:"model"`
	Image          string `json:"image"`           // Base64 or file
	Mask           string `json:"mask,omitempty"`  // Optional
	Prompt         string `json:"prompt"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
}

// ImageVariationRequest 图片变体请求
type ImageVariationRequest struct {
	Model          string `json:"model"`
	Image          string `json:"image"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
}

// ========== Audio API Types ==========

// SpeechRequest 语音合成请求
type SpeechRequest struct {
	Model string  `json:"model"`
	Input string  `json:"input"`
	Voice string  `json:"voice"`           // alloy, echo, fable, onyx, nova, shimmer
	Speed float64 `json:"speed,omitempty"` // 0.25 to 4.0
}

// TranscriptionRequest 转录请求
type TranscriptionRequest struct {
	Model          string   `json:"model"`
	File           string   `json:"file"`             // Base64 or file upload
	Language       string   `json:"language,omitempty"`
	Prompt         string   `json:"prompt,omitempty"`
	ResponseFormat string   `json:"response_format,omitempty"` // json, text, srt, verbose_json, vtt
	Temperature    *float64 `json:"temperature,omitempty"`
}

// TranscriptionResponse 转录响应
type TranscriptionResponse struct {
	Text     string `json:"text"`
	Task     string `json:"task,omitempty"`
	Language string `json:"language,omitempty"`
	Duration float64 `json:"duration,omitempty"`
	Words    []Word `json:"words,omitempty"`
}

// Word 转录词
type Word struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}
