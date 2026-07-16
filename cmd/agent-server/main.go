// Knowledge: AGENT-SERVICE-HTTP + AGENT-SERVICE-SSE — Agent HTTP 服务化
// 将 M05 的 Agent 包装为 HTTP API + SSE 流式响应。
//
// 用法:
//   export OPENAI_API_KEY=sk-xxx
//   export OPENAI_BASE_URL=https://api.deepseek.com/v1
//   go run ./cmd/agent-server/
//   curl -X POST localhost:8080/chat -H 'Content-Type: application/json' -d '{"message":"你好"}'
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Satori2Core/agent-lab/pkg/agent"
	"github.com/Satori2Core/agent-lab/pkg/models"
	"github.com/Satori2Core/agent-lab/pkg/tools"
	"github.com/Satori2Core/agent-lab/pkg/types"
)

// ─── 请求/响应类型 ───

// ChatRequest POST 请求体。
type ChatRequest struct {
	Message string `json:"message"`
}

// ChatResponse 非流式响应。
type ChatResponse struct {
	Answer     string  `json:"answer"`
	Steps      int     `json:"steps"`
	DurationMs float64 `json:"duration_ms"`
}

// SSEEvent SSE 事件。
type SSEEvent struct {
	Type         string  `json:"type"`          // "step" 或 "done" 或 "error"
	Step         int     `json:"step,omitempty"`
	Action       string  `json:"action,omitempty"`
	Observation  string  `json:"observation,omitempty"`
	DurationMs   float64 `json:"duration_ms,omitempty"`
	Answer       string  `json:"answer,omitempty"`
	Error        string  `json:"error,omitempty"`
}

func main() {
	model := initModel()
	ag := initAgent(model)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler(model))
	mux.HandleFunc("/chat", chatHandler(ag))
	mux.HandleFunc("/chat/stream", streamHandler(ag))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      corsMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 120 * time.Second, // Agent 可能需要较长时间
	}

	// 优雅关闭
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("正在关闭服务...")
		server.Shutdown(context.Background())
	}()

	log.Printf("Agent Server 启动: http://localhost:%s", port)
	log.Printf("  POST /chat        — 非流式聊天")
	log.Printf("  POST /chat/stream — SSE 流式聊天")
	log.Printf("  GET  /health      — 健康检查")

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// ─── 初始化 ───

func initModel() models.Model {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY 未设置")
	}
	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}
	modelID := os.Getenv("OPENAI_MODEL")
	if modelID == "" {
		modelID = "deepseek-v4-pro"
	}
	log.Printf("Model: %s (%s)", modelID, baseURL)
	return models.NewOpenAIModel(modelID, models.WithBaseURL(baseURL), models.WithAPIKey(apiKey))
}

func initAgent(model models.Model) *agent.MultiStepAgent {
	reg := tools.NewToolRegistry()

	type WeatherInput struct {
		City string `json:"city" desc:"城市名称，如 Beijing"`
	}
	weatherTool, _ := tools.NewTool("get_weather", "查询指定城市的天气",
		func(ctx context.Context, input WeatherInput) (*types.AgentText, error) {
			weatherMap := map[string]string{
				"beijing":  "北京: 晴天, 22-28°C, 微风",
				"shanghai": "上海: 多云, 25-32°C, 东南风3级",
			}
			if w, ok := weatherMap[input.City]; ok {
				return types.NewAgentText(w)
			}
			return types.NewAgentText(input.City + ": 晴转多云, 20-28°C")
		})
	reg.Register(weatherTool)

	return agent.NewMultiStepAgent("助手", model, reg, agent.WithMaxSteps(5))
}

// ─── Handlers ───

func healthHandler(model models.Model) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// chatHandler 非流式 POST /chat。
func chatHandler(ag *agent.MultiStepAgent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "仅支持 POST", http.StatusMethodNotAllowed)
			return
		}

		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Message == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请提供 message 字段"})
			return
		}

		start := time.Now()
		result, err := ag.Run(r.Context(), req.Message)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, ChatResponse{
			Answer:     result.Answer,
			Steps:      result.Steps,
			DurationMs: float64(result.Duration.Microseconds()) / 1000,
		})
		_ = start
	}
}

// streamHandler SSE 流式 POST /chat/stream。
func streamHandler(ag *agent.MultiStepAgent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "仅支持 POST", http.StatusMethodNotAllowed)
			return
		}

		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Message == "" {
			http.Error(w, "请提供 message 字段", http.StatusBadRequest)
			return
		}

		// SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			log.Println("SSE: 不支持 Flusher")
			return
		}

		// channel 接收步骤事件 + 结果
		type streamResult struct {
			result *agent.RunResult
			err    error
		}
		stepCh := make(chan agent.StepInfo, 10)
		resultCh := make(chan streamResult, 1)

		// goroutine 执行 Agent
		go func() {
			defer close(stepCh)
			agWithObserver := agent.NewMultiStepAgent(ag.Name(), ag.Model(), ag.Tools(),
				agent.WithMaxSteps(5),
				agent.WithStepObserver(func(info agent.StepInfo) {
					stepCh <- info
				}),
			)
			result, err := agWithObserver.Run(r.Context(), req.Message)
			resultCh <- streamResult{result, err}
		}()

		// 读取步骤事件并写 SSE
		totalSteps := 0
		for info := range stepCh {
			totalSteps = info.Step
			event := SSEEvent{
				Type:         "step",
				Step:         info.Step,
				Action:       info.Action,
				Observation:  info.Observation,
				DurationMs:   info.Duration * 1000,
			}
			if info.Error != nil {
				event.Error = info.Error.Error()
			}
			writeSSE(w, flusher, event)
		}

		// 发送最终结果
		doneEvent := SSEEvent{Type: "done"}
		if res := <-resultCh; res.err == nil && res.result != nil {
			doneEvent.Answer = res.result.Answer
			doneEvent.Step = totalSteps
			doneEvent.DurationMs = float64(res.result.Duration.Microseconds()) / 1000
		} else if res.err != nil {
			doneEvent.Error = res.err.Error()
		}
		writeSSE(w, flusher, doneEvent)
	}
}

// ─── 工具函数 ───

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, event SSEEvent) {
	data, _ := json.Marshal(event)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, string(data))
	flusher.Flush()
}

func corsMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	}
}
