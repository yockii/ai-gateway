package services

import (
	"context"
	"fmt"
	"log"

	bifrost "github.com/maximhq/bifrost/core"
	"github.com/yockii/ai-gateway/internal/config"
)

// BifrostClient Bifrost 集成客户端 (per D-01: 库集成模式)
type BifrostClient struct {
	bifrost *bifrost.Bifrost
	config  *config.Config
}

// NewBifrostClient 创建 Bifrost 客户端
func NewBifrostClient(cfg *config.Config) (*BifrostClient, error) {
	// TODO: 研究 Bifrost 初始化参数
	// 参考 Bifrost 项目路径: D:/projects/github.com/maximhq/bifrost/
	b := &bifrost.Bifrost{} // Placeholder - needs actual initialization

	client := &BifrostClient{
		bifrost: b,
		config:  cfg,
	}

	log.Println("✅ Bifrost 客户端初始化成功 (库集成模式)")
	return client, nil
}

// ChatCompletion 聊天完成接口 (非流式)
func (bc *BifrostClient) ChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	// TODO: 实现 Bifrost ChatCompletion 调用
	// per D-02: Bifrost 管理底层供应商路由，业务层管理定价
	return &ChatCompletionResponse{}, fmt.Errorf("not yet implemented")
}

// StreamChatCompletion 流式聊天完成 (per D-04)
func (bc *BifrostClient) StreamChatCompletion(ctx context.Context, req *ChatCompletionRequest) (<-chan StreamChunk, error) {
	// TODO: 实现 SSE 流式响应
	return nil, fmt.Errorf("not yet implemented")
}

// SelectProvider 选择最优供应商 (per D-05: 协同故障转移)
func (bc *BifrostClient) SelectProvider(ctx context.Context, modelID string, candidates []ProviderCandidate) (*ProviderCandidate, error) {
	// TODO: 实现 Bifrost 原生故障转移逻辑
	return nil, fmt.Errorf("not yet implemented")
}

// Request types (compatible with OpenAI format per D-03)
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature *float64      `json:"temperature,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Response types
type ChatCompletionResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   Usage        `json:"usage"`
}

type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int32 `json:"prompt_tokens"`
	CompletionTokens int32 `json:"completion_tokens"`
	TotalTokens      int32 `json:"total_tokens"`
}

type StreamChunk struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

type StreamChoice struct {
	Index        int          `json:"index"`
	Delta        MessageDelta `json:"delta"`
	FinishReason *string      `json:"finish_reason"`
}

type MessageDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type ProviderCandidate struct {
	ID         string
	Name       string
	ModelName  string
	Priority   int
	IsBackup   bool
}
