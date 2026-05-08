package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/yockii/ai-gateway/internal/middleware"
	"github.com/yockii/ai-gateway/pkg/api"
)

// CreateSpeech 语音合成 (TTS, per D-04: 流式支持)
func (h *Handler) CreateSpeech(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var req api.SpeechRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Invalid request body",
				Type:    "invalid_request_error",
				Code:    fiber.StatusBadRequest,
			},
		})
	}

	// 验证请求
	if err := h.validateSpeechRequest(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: err.Error(),
				Type:    "invalid_request_error",
				Code:    fiber.StatusBadRequest,
			},
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 调用 Gateway
	audioData, err := h.gateway.TextToSpeech(ctx, userID, &req)
	if err != nil {
		log.Printf("语音合成错误: 用户=%s 模型=%s 错误=%v", userID, req.Model, err)
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Failed to generate speech",
				Type:    "api_error",
				Code:    fiber.StatusInternalServerError,
			},
		})
	}

	// 设置响应头
	c.Set("Content-Type", "audio/mpeg")

	return c.Send(audioData)
}

// CreateTranscription 语音识别 (STT)
func (h *Handler) CreateTranscription(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	// 获取上传的音频文件
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: api.ErrorDetail{
				Message: "Audio file is required",
				Type:    "invalid_request_error",
				Code:    fiber.StatusBadRequest,
			},
		})
	}

	model := c.FormValue("model", "whisper-1")
	language := c.FormValue("language")
	prompt := c.FormValue("prompt")
	responseFormat := c.FormValue("response_format", "json")
	temperature := parseFloatOrDefault(c.FormValue("temperature"), 0)

	req := api.TranscriptionRequest{
		Model:          model,
		Language:       language,
		Prompt:         prompt,
		ResponseFormat: responseFormat,
		Temperature:    &temperature,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// TODO: 处理音频文件并转录
	_ = userID
	_ = file
	_ = ctx
	_ = req

	return c.Status(fiber.StatusNotImplemented).JSON(api.ErrorResponse{
		Error: api.ErrorDetail{
			Message: "Transcription not yet implemented",
			Type:    "not_implemented",
			Code:    fiber.StatusNotImplemented,
		},
	})
}

// CreateTranslation 翻译
func (h *Handler) CreateTranslation(c fiber.Ctx) error {
	// TODO: 实现翻译功能
	return c.Status(fiber.StatusNotImplemented).JSON(api.ErrorResponse{
		Error: api.ErrorDetail{
			Message: "Translation not yet implemented",
			Type:    "not_implemented",
			Code:    fiber.StatusNotImplemented,
		},
	})
}

// validateSpeechRequest 验证语音请求
func (h *Handler) validateSpeechRequest(req *api.SpeechRequest) error {
	if req.Model == "" {
		return fmt.Errorf("model is required")
	}

	validModels := map[string]bool{
		"tts-1":   true,
		"tts-1-hd": true,
	}

	if !validModels[req.Model] {
		return fmt.Errorf("invalid model: %s", req.Model)
	}

	if req.Input == "" {
		return fmt.Errorf("input is required")
	}

	if len(req.Input) > 4096 {
		return fmt.Errorf("input is too long (max 4096 characters)")
	}

	validVoices := map[string]bool{
		"alloy":   true,
		"echo":    true,
		"fable":   true,
		"onyx":    true,
		"nova":    true,
		"shimmer": true,
	}

	if req.Voice != "" && !validVoices[req.Voice] {
		return fmt.Errorf("invalid voice: %s", req.Voice)
	}

	return nil
}

func parseFloatOrDefault(s string, defaultVal float64) float64 {
	result := defaultVal
	if s != "" {
		fmt.Sscanf(s, "%f", &result)
	}
	return result
}
