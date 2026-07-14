// Knowledge: AGENT-MODEL-OPENAI — OpenAI 模型适配
// OpenAIModel 实现 Model 接口，支持 OpenAI 兼容 API（包括 Ollama、vLLM 等）。
// Reference: smolagents models.py → OpenAIModel(ApiModel) 类
package models

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OpenAIModel 是 OpenAI Chat Completions API 的客户端实现。
//
// 支持任何 OpenAI 兼容的 API 端点（通过 BaseURL 配置）：
//   - OpenAI:     https://api.openai.com/v1
//   - Ollama:     http://localhost:11434/v1
//   - vLLM:       http://localhost:8000/v1
//   - LocalAI:    http://localhost:8080/v1
//
// 对标 smolagents models.py → OpenAIModel 类
//   Python: OpenAIModel(model_id="gpt-4o", api_key=..., base_url=...)
//   Go:     NewOpenAIModel("gpt-4o", WithBaseURL(...), WithAPIKey(...))
type OpenAIModel struct {
	modelID    string
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// OpenAIOption 是 OpenAIModel 的函数选项（Functional Options 模式）。
type OpenAIOption func(*OpenAIModel)

// WithBaseURL 设置 API 端点的基础 URL。
//
// 默认: https://api.openai.com/v1
func WithBaseURL(url string) OpenAIOption {
	return func(m *OpenAIModel) {
		m.baseURL = strings.TrimRight(url, "/")
	}
}

// WithAPIKey 设置 API 密钥。
//
// 默认: 从环境变量 OPENAI_API_KEY 读取（通过 os.Getenv）
func WithAPIKey(key string) OpenAIOption {
	return func(m *OpenAIModel) {
		m.apiKey = key
	}
}

// WithHTTPClient 设置自定义 HTTP 客户端。
//
// 用于注入 mock 客户端（测试用）或配置代理/超时。
func WithHTTPClient(client *http.Client) OpenAIOption {
	return func(m *OpenAIModel) {
		m.httpClient = client
	}
}

// NewOpenAIModel 创建一个新的 OpenAIModel 实例。
//
// 参数：
//   - modelID: 模型标识符（如 "gpt-4o", "llama3.1" 等）
//   - opts: 可选的配置函数
//
// 返回：
//   - 配置好的 *OpenAIModel（实现 Model 接口）
func NewOpenAIModel(modelID string, opts ...OpenAIOption) *OpenAIModel {
	m := &OpenAIModel{
		modelID:    modelID,
		baseURL:    "https://api.openai.com/v1",
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// Generate 发送消息列表并返回完整响应（非流式）。
//
// 发送 POST /v1/chat/completions，stream=false，等待完整响应。
// 对标 smolagents OpenAIModel.__call__/generate()。
func (m *OpenAIModel) Generate(ctx context.Context, messages []ChatMessage) (*Response, error) {
	body := m.buildRequestBody(messages, false)
	respData, err := m.doRequest(ctx, body)
	if err != nil {
		return nil, err
	}
	return m.parseResponse(respData)
}

// GenerateStream 发送消息列表并通过 channel 返回流式增量。
//
// 发送 POST /v1/chat/completions，stream=true，解析 SSE 事件流。
// 对标 smolagents 中的流式处理逻辑。
func (m *OpenAIModel) GenerateStream(ctx context.Context, messages []ChatMessage) (<-chan Delta, error) {
	body := m.buildRequestBody(messages, true)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.buildURL(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("OpenAI: 创建请求失败: %w", err)
	}
	m.setHeaders(req)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI: 请求失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI: API 错误 %d: %s", resp.StatusCode, string(body))
	}

	ch := make(chan Delta, 10)
	go m.streamSSE(ctx, resp.Body, ch)
	return ch, nil
}

// ─── 内部方法 ───

// buildURL 构造完整的 API 端点 URL。
func (m *OpenAIModel) buildURL() string {
	return m.baseURL + "/chat/completions"
}

// setHeaders 设置请求头（Content-Type、Authorization）。
func (m *OpenAIModel) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if m.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+m.apiKey)
	}
}

// buildRequestBody 构造 HTTP 请求体。
func (m *OpenAIModel) buildRequestBody(messages []ChatMessage, stream bool) []byte {
	req := map[string]any{
		"model":    m.modelID,
		"messages": messages,
		"stream":   stream,
	}
	data, _ := json.Marshal(req)
	return data
}

// doRequest 发送非流式请求并返回响应体。
func (m *OpenAIModel) doRequest(ctx context.Context, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.buildURL(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("OpenAI: 创建请求失败: %w", err)
	}
	m.setHeaders(req)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI: 请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("OpenAI: 读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI: API 错误 %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// parseResponse 解析非流式响应的 JSON。
func (m *OpenAIModel) parseResponse(data []byte) (*Response, error) {
	var raw struct {
		Choices []struct {
			Message struct {
				Content   string     `json:"content"`
				ToolCalls []ToolCall `json:"tool_calls"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("OpenAI: 解析响应失败: %w", err)
	}

	if len(raw.Choices) == 0 {
		return nil, fmt.Errorf("OpenAI: 响应中没有 choices")
	}

	choice := raw.Choices[0]
	return &Response{
		Content:      choice.Message.Content,
		ToolCalls:    choice.Message.ToolCalls,
		FinishReason: choice.FinishReason,
	}, nil
}

// streamSSE 在 goroutine 中解析 SSE 流并发送 Delta。
func (m *OpenAIModel) streamSSE(ctx context.Context, body io.ReadCloser, ch chan<- Delta) {
	defer close(ch)
	defer body.Close()

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			ch <- Delta{Done: true, Error: ctx.Err()}
			return
		default:
		}

		line := scanner.Text()
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			ch <- Delta{Done: true}
			return
		}

		delta := m.parseSSEData(data)
		ch <- delta
	}

	if err := scanner.Err(); err != nil {
		ch <- Delta{Done: true, Error: fmt.Errorf("OpenAI: SSE 读取错误: %w", err)}
	} else {
		ch <- Delta{Done: true}
	}
}

// parseSSEData 解析单条 SSE data 行。
func (m *OpenAIModel) parseSSEData(data string) Delta {
	var raw struct {
		Choices []struct {
			Delta struct {
				Content   string     `json:"content"`
				ToolCalls []ToolCall `json:"tool_calls"`
			} `json:"delta"`
		} `json:"choices"`
	}

	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		return Delta{Error: fmt.Errorf("OpenAI: SSE 解析失败: %w", err)}
	}

	if len(raw.Choices) == 0 {
		return Delta{}
	}

	d := raw.Choices[0].Delta
	return Delta{
		Content:   d.Content,
		ToolCalls: d.ToolCalls,
	}
}
