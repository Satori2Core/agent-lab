// Knowledge: AGENT-MODEL-OPENAI — OpenAI 模型适配
// 测试 OpenAIModel：HTTP 调用、JSON 序列化、流式 SSE 解析。
// Reference: smolagents models.py → OpenAIModel 类
package models

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestOpenAIModelGenerate 验证基本的 Generate 调用。
func TestOpenAIModelGenerate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求格式
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req struct {
			Model    string        `json:"model"`
			Messages []ChatMessage `json:"messages"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		if req.Model != "test-model" {
			t.Errorf("model = %q, want %q", req.Model, "test-model")
		}
		if len(req.Messages) == 0 {
			t.Error("messages should not be empty")
		}

		// 返回 mock 响应
		resp := map[string]any{
			"id":      "chatcmpl-test",
			"object":  "chat.completion",
			"choices": []map[string]any{{"index": 0, "message": map[string]any{"role": "assistant", "content": "Hello from mock!"}, "finish_reason": "stop"}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	model := NewOpenAIModel("test-model", WithBaseURL(server.URL), WithAPIKey("test-key"))
	resp, err := model.Generate(context.Background(), []ChatMessage{
		{Role: RoleUser, Content: "Hi"},
	}, nil)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if resp.Content != "Hello from mock!" {
		t.Errorf("Content = %q", resp.Content)
	}
}

// TestOpenAIModelGenerateStream 验证流式输出。
func TestOpenAIModelGenerateStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 返回 SSE 流
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected http.Flusher")
			return
		}

		chunks := []string{"Hello", " from", " stream!"}
		for _, chunk := range chunks {
			data := map[string]any{
				"choices": []map[string]any{{
					"index": 0,
					"delta": map[string]any{"content": chunk},
				}},
			}
			jsonData, _ := json.Marshal(data)
			w.Write([]byte("data: " + string(jsonData) + "\n\n"))
			flusher.Flush()
		}
		// 发送结束信号
		w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
	}))
	defer server.Close()

	model := NewOpenAIModel("test-model", WithBaseURL(server.URL))
	ch, err := model.GenerateStream(context.Background(), []ChatMessage{
		{Role: RoleUser, Content: "Hi"},
	}, nil)
	if err != nil {
		t.Fatalf("GenerateStream() error: %v", err)
	}

	var content strings.Builder
	for delta := range ch {
		if delta.Error != nil {
			t.Fatalf("stream error: %v", delta.Error)
		}
		content.WriteString(delta.Content)
	}

	if content.String() != "Hello from stream!" {
		t.Errorf("streamed content = %q, want %q", content.String(), "Hello from stream!")
	}
}

// compileTimeCheck
var _ Model = (*OpenAIModel)(nil)
