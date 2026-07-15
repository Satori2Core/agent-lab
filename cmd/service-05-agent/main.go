// Knowledge: AGENT-LOOP-REACT — Agent 核心循环端到端验证
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Satori2Core/agent-lab/pkg/agent"
	"github.com/Satori2Core/agent-lab/pkg/models"
	"github.com/Satori2Core/agent-lab/pkg/tools"
	"github.com/Satori2Core/agent-lab/pkg/types"
)

func main() {
	fmt.Println("Module 5: ReAct 核心循环 — 端到端验证")
	fmt.Println("============================================")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("\nOPENAI_API_KEY 未设置")
		return
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}
	modelID := os.Getenv("OPENAI_MODEL")
	if modelID == "" {
		modelID = "deepseek-v4-pro"
	}
	fmt.Printf("  Model: %s | API: %s\n\n", modelID, baseURL)

	// 1. Model
	model := models.NewOpenAIModel(modelID,
		models.WithBaseURL(baseURL),
		models.WithAPIKey(apiKey),
	)

	// 2. Tools
	reg := tools.NewToolRegistry()

	type WeatherInput struct {
		City string `json:"city" desc:"城市名称，如 Beijing"`
	}
	weatherTool, _ := tools.NewTool("get_weather", "查询指定城市的天气",
		func(ctx context.Context, input WeatherInput) (*types.AgentText, error) {
			weatherMap := map[string]string{
				"beijing":  "北京: 晴天, 22-28°C, 微风",
				"shanghai": "上海: 多云, 25-32°C, 东南风3级",
				"shenzhen": "深圳: 阵雨, 26-30°C",
				"tokyo":    "东京: 阴天, 18-24°C",
			}
			if w, ok := weatherMap[input.City]; ok {
				return types.NewAgentText(w)
			}
			return types.NewAgentText(input.City + ": 晴转多云, 20-28°C")
		})
	reg.Register(weatherTool)

	type CalcInput struct {
		Expression string `json:"expression" desc:"数学表达式"`
	}
	calcTool, _ := tools.NewTool("calculate", "执行数学计算",
		func(ctx context.Context, input CalcInput) (*types.AgentText, error) {
			return types.NewAgentText("计算结果: " + input.Expression + " ≈ 42 (演示)")
		})
	reg.Register(calcTool)

	// 3. Agent（带实时观察者）
	ag := agent.NewMultiStepAgent("实用助手", model, reg,
		agent.WithMaxSteps(5),
		agent.WithStepObserver(func(info agent.StepInfo) {
			status := "✅"
			if info.Error != nil {
				status = "❌"
			}
			obs := info.Observation
			if len(obs) > 60 {
				obs = obs[:60] + "..."
			}
			fmt.Printf("  [Step %d] %s %s → %s (%.1fs)\n",
				info.Step, status, info.Action, obs, info.Duration)
		}),
	)

	// 4. Run
	task := "北京和上海今天天气怎么样？哪个城市更适合户外活动？"
	fmt.Println("── 任务 ──")
	fmt.Printf("  %s\n\n", task)
	fmt.Println("── Agent 实时日志 ──")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := ag.Run(ctx, task)
	if err != nil {
		fmt.Printf("\n  ❌ Agent 失败: %v\n", err)
		return
	}

	fmt.Printf("\n── 最终答案（%d步，%.1f秒）──\n%s\n",
		result.Steps, result.Duration.Seconds(), result.Answer)
}
