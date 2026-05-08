package benchmarks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/yockii/ai-gateway/pkg/api"
)

const (
	baseURL = "http://localhost:8080"
	testKey = "test-key"
)

var client = &http.Client{
	Timeout: 30 * time.Second,
}

// BenchmarkListModels 基准测试：列出模型
func BenchmarkListModels(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", baseURL+"/v1/models", nil)
		req.Header.Set("Authorization", "Bearer "+testKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			b.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != 200 {
			b.Fatalf("Unexpected status: %d", resp.StatusCode)
		}
	}
}

// BenchmarkChatCompletions 基准测试：聊天完成
func BenchmarkChatCompletions(b *testing.B) {
	req := api.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []api.ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	body, _ := json.Marshal(req)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		httpReq, _ := http.NewRequest("POST", baseURL+"/v1/chat/completions", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+testKey)

		resp, err := client.Do(httpReq)
		if err != nil {
			b.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()

		// 允许非 200 状态（因为上游可能未配置）
	}
}

// BenchmarkHealthCheck 基准测试：健康检查
func BenchmarkHealthCheck(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resp, err := client.Get(baseURL + "/health")
		if err != nil {
			b.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != 200 {
			b.Fatalf("Unexpected status: %d", resp.StatusCode)
		}
	}
}

// BenchmarkConcurrentRequests 并发请求基准测试
func BenchmarkConcurrentRequests(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", baseURL+"/v1/models", nil)
			req.Header.Set("Authorization", "Bearer "+testKey)

			resp, err := client.Do(req)
			if err != nil {
				continue
			}
			resp.Body.Close()
		}
	})
}

// BenchmarkMiddleware 基准测试中间件性能
func BenchmarkMiddleware(b *testing.B) {
	// 测试带认证的请求
	b.Run("WithAuth", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest("GET", baseURL+"/v1/models", nil)
			req.Header.Set("Authorization", "Bearer "+testKey)

			resp, err := client.Do(req)
			if err != nil {
				continue
			}
			resp.Body.Close()
		}
	})

	// 测试不带认证的请求
	b.Run("WithoutAuth", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := client.Get(baseURL + "/health")
			if err != nil {
				continue
			}
			resp.Body.Close()
		}
	})
}

// TestAPIResponseTime 测试 API 响应时间
func TestAPIResponseTime(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		maxLatency time.Duration
	}{
		{"HealthCheck", baseURL + "/health", 10 * time.Millisecond},
		{"ListModels", baseURL + "/v1/models", 20 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()

			req, _ := http.NewRequest("GET", tt.url, nil)
			if tt.url != baseURL+"/health" {
				req.Header.Set("Authorization", "Bearer "+testKey)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			latency := time.Since(start)

			t.Logf("Latency for %s: %v", tt.name, latency)

			if latency > tt.maxLatency {
				t.Errorf("Latency %v exceeds maximum %v", latency, tt.maxLatency)
			}
		})
	}
}

// TestServerAvailable 测试服务器是否可用
func TestServerAvailable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			t.Fatal("Server not available within timeout")
		default:
			resp, err := client.Get(baseURL + "/health")
			if err == nil && resp.StatusCode == 200 {
				resp.Body.Close()
				t.Log("Server is available")
				return
			}
			if resp != nil {
				resp.Body.Close()
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// BenchmarkJSONSerialization 基准测试 JSON 序列化
func BenchmarkJSONSerialization(b *testing.B) {
	chatResp := &api.ChatCompletionResponse{
		ID:      "test-id",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "gpt-3.5-turbo",
		Choices: []api.ChatChoice{
			{
				Index: 0,
				Message: api.ChatMessage{
					Role:    "assistant",
					Content: "Hello, how can I help you today?",
				},
				FinishReason: "stop",
			},
		},
		Usage: api.Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(chatResp)
		if err != nil {
			b.Fatalf("JSON marshal failed: %v", err)
		}
	}
}

// BenchmarkJSONDeserialization 基准测试 JSON 反序列化
func BenchmarkJSONDeserialization(b *testing.B) {
	chatReqJSON := []byte(`{
		"model": "gpt-3.5-turbo",
		"messages": [
			{"role": "user", "content": "Hello"}
		],
		"temperature": 0.7,
		"max_tokens": 100
	}`)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var req api.ChatCompletionRequest
		err := json.Unmarshal(chatReqJSON, &req)
		if err != nil {
			b.Fatalf("JSON unmarshal failed: %v", err)
		}
	}
}

// Helper function to make authenticated requests
func makeAuthenticatedRequest(method, url string, body []byte) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+testKey)
	req.Header.Set("Content-Type", "application/json")

	return client.Do(req)
}

// BenchmarkStreamingResponse 基准测试流式响应
func BenchmarkStreamingResponse(b *testing.B) {
	req := api.ChatCompletionRequest{
		Model:     "gpt-3.5-turbo",
		Stream:    true,
		Messages: []api.ChatMessage{
			{Role: "user", Content: "Say hello"},
		},
	}

	body, _ := json.Marshal(req)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resp, err := makeAuthenticatedRequest("POST", baseURL+"/v1/chat/completions", body)
		if err != nil {
			continue
		}

		// 读取响应体
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// TestQPSLoadTest QPS 负载测试
func TestQPSLoadTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	const (
		concurrency = 100
		duration    = 10 * time.Second
		requests    = 1000
	)

	results := make(chan time.Duration, requests*2)
	start := make(chan struct{})
	var wg sync.WaitGroup

	// 启动并发 workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start // 等待开始信号

			for j := 0; j < requests/concurrency; j++ {
				reqStart := time.Now()
				resp, err := http.Get(baseURL + "/health")
				if err == nil {
					resp.Body.Close()
				}
				results <- time.Since(reqStart)
			}
		}()
	}

	// 开始测试
	testStart := time.Now()
	close(start)

	// 等待完成
	wg.Wait()
	close(results)

	testDuration := time.Since(testStart)

	// 统计结果
	var totalLatency time.Duration
	count := 0
	maxLatency := time.Duration(0)

	for latency := range results {
		totalLatency += latency
		count++
		if latency > maxLatency {
			maxLatency = latency
		}
	}

	if count > 0 {
		avgLatency := totalLatency / time.Duration(count)
		qps := float64(count) / testDuration.Seconds()

		t.Logf("Load Test Results:")
		t.Logf("  Total Requests: %d", count)
		t.Logf("  Duration: %v", testDuration)
		t.Logf("  QPS: %.2f", qps)
		t.Logf("  Avg Latency: %v", avgLatency)
		t.Logf("  Max Latency: %v", maxLatency)

		// 验证性能目标
		if qps < 1000 {
			t.Errorf("QPS %.2f is below target 1000", qps)
		}
		if avgLatency > 20*time.Millisecond {
			t.Errorf("Avg latency %v exceeds target 20ms", avgLatency)
		}
	}
}
