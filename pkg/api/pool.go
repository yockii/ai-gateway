package api

import (
	"sync"
)

// 对象池用于减少内存分配

var (
	// ChatCompletionResponsePool 聊天完成响应对象池
	chatCompletionResponsePool = sync.Pool{
		New: func() interface{} {
			return &ChatCompletionResponse{}
		},
	}

	// ChatCompletionRequestPool 聊天完成请求对象池
	chatCompletionRequestPool = sync.Pool{
		New: func() interface{} {
			return &ChatCompletionRequest{}
		},
	}

	// ImageResponsePool 图片响应对象池
	imageResponsePool = sync.Pool{
		New: func() interface{} {
			return &ImageResponse{}
		},
	}

	// EmbeddingResponsePool 嵌入响应对象池
	embeddingResponsePool = sync.Pool{
		New: func() interface{} {
			return &EmbeddingResponse{}
		},
	}
)

// AcquireChatCompletionResponse 从池中获取 ChatCompletionResponse
func AcquireChatCompletionResponse() *ChatCompletionResponse {
	return chatCompletionResponsePool.Get().(*ChatCompletionResponse)
}

// ReleaseChatCompletionResponse 将 ChatCompletionResponse 归还到池
func ReleaseChatCompletionResponse(resp *ChatCompletionResponse) {
	// 重置对象以避免数据泄漏
	*resp = ChatCompletionResponse{}
	chatCompletionResponsePool.Put(resp)
}

// AcquireChatCompletionRequest 从池中获取 ChatCompletionRequest
func AcquireChatCompletionRequest() *ChatCompletionRequest {
	return chatCompletionRequestPool.Get().(*ChatCompletionRequest)
}

// ReleaseChatCompletionRequest 将 ChatCompletionRequest 归还到池
func ReleaseChatCompletionRequest(req *ChatCompletionRequest) {
	*req = ChatCompletionRequest{}
	chatCompletionRequestPool.Put(req)
}

// AcquireImageResponse 从池中获取 ImageResponse
func AcquireImageResponse() *ImageResponse {
	return imageResponsePool.Get().(*ImageResponse)
}

// ReleaseImageResponse 将 ImageResponse 归还到池
func ReleaseImageResponse(resp *ImageResponse) {
	*resp = ImageResponse{}
	imageResponsePool.Put(resp)
}

// AcquireEmbeddingResponse 从池中获取 EmbeddingResponse
func AcquireEmbeddingResponse() *EmbeddingResponse {
	return embeddingResponsePool.Get().(*EmbeddingResponse)
}

// ReleaseEmbeddingResponse 将 EmbeddingResponse 归还到池
func ReleaseEmbeddingResponse(resp *EmbeddingResponse) {
	*resp = EmbeddingResponse{}
	embeddingResponsePool.Put(resp)
}
