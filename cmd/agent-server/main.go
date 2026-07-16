// Knowledge: AGENT-SERVICE-HTTP + AGENT-SERVICE-SSE — Agent HTTP 服务化
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Satori2Core/agent-lab/pkg/agent"
	"github.com/Satori2Core/agent-lab/pkg/agent/codeagent"
	"github.com/Satori2Core/agent-lab/pkg/models"
	"github.com/Satori2Core/agent-lab/pkg/tools"
	"github.com/Satori2Core/agent-lab/pkg/types"
)

type ChatRequest struct {
	Message string `json:"message"`
}

type ChatResponse struct {
	Answer     string  `json:"answer"`
	Steps      int     `json:"steps"`
	DurationMs float64 `json:"duration_ms"`
}

type SSEEvent struct {
	Type        string  `json:"type"`
	Step        int     `json:"step,omitempty"`
	Action      string  `json:"action,omitempty"`
	Observation string  `json:"observation,omitempty"`
	DurationMs  float64 `json:"duration_ms,omitempty"`
	Answer      string  `json:"answer,omitempty"`
	Error       string  `json:"error,omitempty"`
}

func main() {
	model := initModel()
	ag := initAgent(model)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler())
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
		WriteTimeout: 120 * time.Second,
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("正在关闭服务...")
		server.Shutdown(context.Background())
	}()

	log.Printf("Agent Server: http://localhost:%s", port)
	log.Printf("  POST /chat /chat/stream | GET /health")
	log.Fatal(server.ListenAndServe())
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

	// Weather Tool — 真实 wttr.in API
	type WeatherInput struct {
		City string `json:"city" desc:"城市名（支持中英文）"`
	}
	weatherTool, _ := tools.NewTool("get_weather", "查询指定城市的实时天气（数据来源: wttr.in）",
		func(ctx context.Context, input WeatherInput) (*types.AgentText, error) {
			return queryRealWeather(ctx, input.City)
		})
	reg.Register(weatherTool)

	// Code Execution Tool
	executor := codeagent.NewLocalExecutor()
	execTool, _ := codeagent.NewExecuteCodeTool(executor)
	reg.Register(execTool)

	return agent.NewMultiStepAgent("助手", model, reg, agent.WithMaxSteps(10))
}

// ─── 真实天气 API（wttr.in，免费，无需 Key）───

func queryRealWeather(ctx context.Context, city string) (*types.AgentText, error) {
	url := fmt.Sprintf("https://wttr.in/%s?format=j1", city)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return types.NewAgentText(fmt.Sprintf("天气查询失败: %v", err))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return types.NewAgentText(fmt.Sprintf("天气 API 请求失败: %v", err))
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var data struct {
		CurrentCondition []struct {
			TempC       string `json:"temp_C"`
			FeelsLikeC  string `json:"FeelsLikeC"`
			Humidity    string `json:"humidity"`
			WeatherDesc []struct {
				Value string `json:"value"`
			} `json:"weatherDesc"`
			WindspeedKmph string `json:"windspeedKmph"`
		} `json:"current_condition"`
	}

	if err := json.Unmarshal(body, &data); err != nil || len(data.CurrentCondition) == 0 {
		return types.NewAgentText(fmt.Sprintf("%s: 暂无天气数据", city))
	}

	c := data.CurrentCondition[0]
	desc := "未知"
	if len(c.WeatherDesc) > 0 {
		desc = c.WeatherDesc[0].Value
	}

	result := fmt.Sprintf("%s: %s, %s°C (体感 %s°C), 湿度 %s%%, 风速 %s km/h",
		city, desc, c.TempC, c.FeelsLikeC, c.Humidity, c.WindspeedKmph)
	return types.NewAgentText(result)
}

// ─── Handlers ───

func healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func chatHandler(ag *agent.MultiStepAgent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Message == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
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
	}
}

func streamHandler(ag *agent.MultiStepAgent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Message == "" {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			return
		}

		type streamResult struct {
			result *agent.RunResult
			err    error
		}
		stepCh := make(chan agent.StepInfo, 10)
		resultCh := make(chan streamResult, 1)

		go func() {
			defer close(stepCh)
			a := agent.NewMultiStepAgent(ag.Name(), ag.Model(), ag.Tools(),
				agent.WithMaxSteps(5),
				agent.WithStepObserver(func(info agent.StepInfo) {
					stepCh <- info
				}),
			)
			result, err := a.Run(r.Context(), req.Message)
			resultCh <- streamResult{result, err}
		}()

		totalSteps := 0
		for info := range stepCh {
			totalSteps = info.Step
			event := SSEEvent{
				Type:        "step",
				Step:        info.Step,
				Action:      info.Action,
				Observation: info.Observation,
				DurationMs:  info.Duration * 1000,
			}
			if info.Error != nil {
				event.Error = info.Error.Error()
			}
			writeSSE(w, flusher, event)
		}

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
