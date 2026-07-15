// Knowledge: AGENT-LOOP-REACT — Agent 核心循环端到端验证
// 用真实的 LLM + 工具演示 ReAct 循环的完整工作流程。
//
// 用法:
//   export OPENAI_API_KEY=sk-xxx
//   export OPENAI_BASE_URL=https://api.deepseek.com/v1  (可选，默认 OpenAI)
//   export OPENAI_MODEL=deepseek-v4-pro  (可选，默认 gpt-4o-mini)
//   go run ./cmd/service-05-agent/
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
		fmt.Println("\n⚠ OPENAI_API_KEY 未设置，使用内置模拟演示 Agent 架构...")
		demoWithoutAPI()
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
	fmt.Printf("  Model:  %s\n  API:    %s\n\n", modelID, baseURL)

	// ─── 1. 创建 Model ───
	model := models.NewOpenAIModel(modelID,
		models.WithBaseURL(baseURL),
		models.WithAPIKey(apiKey),
	)

	// ─── 2. 注册工具 ───
	reg := tools.NewToolRegistry()

	// 工具1: 天气查询
	type WeatherInput struct {
		City string `json:"city" desc:"城市名称，如 Beijing"`
	}
	weatherTool, _ := tools.NewTool("get_weather", "查询指定城市的天气",
		func(ctx context.Context, input WeatherInput) (*types.AgentText, error) {
			// 模拟天气数据
			weatherMap := map[string]string{
				"beijing":  "北京: 晴天, 22-28°C, 微风",
				"shanghai": "上海: 多云, 25-32°C, 东南风3级",
				"shenzhen": "深圳: 阵雨, 26-30°C, 南风2级",
				"tokyo":    "东京: 阴天, 18-24°C, 北风4级",
			}
			city := input.City
			if weather, ok := weatherMap[city]; ok {
				return types.NewAgentText(weather)
			}
			return types.NewAgentText(input.City + ": 晴转多云, 20-28°C")
		})
	reg.Register(weatherTool)

	// 工具2: 计算器
	type CalcInput struct {
		Expression string `json:"expression" desc:"数学表达式，如 '2+3*4' 或 'sqrt(16)'"`
	}
	calcTool, _ := tools.NewTool("calculate", "执行数学计算",
		func(ctx context.Context, input CalcInput) (*types.AgentText, error) {
			// 简单模拟
			return types.NewAgentText("计算结果: " + input.Expression + " ≈ 42 (演示)")
		})
	reg.Register(calcTool)

	// ─── 3. 创建 Agent ───
	ag := agent.NewMultiStepAgent("实用助手", model, reg, agent.WithMaxSteps(5))

	// ─── 4. 执行任务 ───
	task := "北京和上海今天天气怎么样？哪个城市更适合户外活动？"

	fmt.Println("── 任务 ──")
	fmt.Printf("  %s\n\n", task)
	fmt.Println("── Agent 执行中 ──")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := ag.Run(ctx, task)
	if err != nil {
		fmt.Printf("\n  ❌ Agent 失败: %v\n", err)
		return
	}

	fmt.Printf("\n── 结果（耗时 %.1f秒，%d 步）──\n", result.Duration.Seconds(), result.Steps)
	fmt.Printf("  %s\n\n", result.Answer)

	// ─── 5. 回放执行轨迹 ───
	fmt.Println(result.Memory.Replay())
}

func demoWithoutAPI() {
	fmt.Println("  Agent 架构说明:")
	fmt.Println("    MultiStepAgent = Model + ToolRegistry + AgentMemory")
	fmt.Println()
	fmt.Println("  ReAct 循环:")
	fmt.Println("    for step in 1..maxSteps:")
	fmt.Println("      response = model.Generate(messages, tools)")
	fmt.Println("      if response.ToolCalls:")
	fmt.Println("        execute tool → record observation → continue loop")
	fmt.Println("      else:")
	fmt.Println("        record final answer → exit loop")
	fmt.Println()
	fmt.Println("  设置 OPENAI_API_KEY 后运行可看到 Agent 自主调用工具完成任务。")
}
