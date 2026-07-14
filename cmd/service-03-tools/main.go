// Knowledge: AGENT-TOOL-INTERFACE — 工具系统集成验证
// 演示 Tool 定义、注册、Schema 生成、校验、执行的完整流程。
//
// 用法:
//   go run ./cmd/service-03-tools/
package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Satori2Core/agent-lab/pkg/tools"
	"github.com/Satori2Core/agent-lab/pkg/types"
)

func main() {
	fmt.Println("Week 3: Tool 系统 — 集成验证")
	fmt.Println("============================================")

	// ─── 1. 创建 ToolRegistry ───
	registry := tools.NewToolRegistry()

	// ─── 2. 定义并注册工具 ───

	// 工具1: 天气查询
	type WeatherInput struct {
		City string `json:"city" desc:"城市名称，如 Beijing"`
		Days int    `json:"days" desc:"查询未来几天，1-7"`
	}
	weatherTool, _ := tools.NewTool("get_weather", "查询指定城市未来几天的天气",
		func(ctx context.Context, input WeatherInput) (*types.AgentText, error) {
			result := fmt.Sprintf("%s 未来%d天: 晴天，22-28°C", input.City, input.Days)
			return types.NewAgentText(result)
		})
	registry.Register(weatherTool)

	// 工具2: 计算器
	type CalcInput struct {
		Expression string `json:"expression" desc:"数学表达式，如 '2+3*4'"`
	}
	calcTool, _ := tools.NewTool("calculate", "计算数学表达式",
		func(ctx context.Context, input CalcInput) (*types.AgentText, error) {
			return types.NewAgentText(input.Expression + " = 14 (演示)")
		})
	registry.Register(calcTool)

	// ─── 3. 展示 JSON Schema ───
	fmt.Println("── 工具的 JSON Schema ──")
	for _, t := range registry.List() {
		fmt.Printf("\n  %s: %s\n", t.Name, t.Description)
		var schema map[string]any
		json.Unmarshal(t.Parameters, &schema)
		schemaJSON, _ := json.MarshalIndent(schema, "    ", "  ")
		fmt.Printf("  Schema: %s\n", string(schemaJSON))
	}

	// ─── 4. 展示 OpenAI function calling 格式 ───
	fmt.Println("\n── OpenAI function calling 格式 ──")
	openaiTools := registry.ToOpenAI()
	openaiJSON, _ := json.MarshalIndent(openaiTools, "", "  ")
	fmt.Println(string(openaiJSON))

	// ─── 5. 模拟 Tool 调用 ───
	fmt.Println("\n── 模拟 Tool 调用 ──")

	// 模拟 LLM 调用 get_weather
	weather, _ := registry.Get("get_weather")
	params := json.RawMessage(`{"city":"Beijing","days":3}`)

	fmt.Printf("  调用: %s(%s)\n", weather.Name, string(params))

	// 先校验
	if err := weather.Validate(params); err != nil {
		fmt.Printf("  ❌ 校验失败: %v\n", err)
		return
	}
	fmt.Println("  ✅ 参数校验通过")

	// 再执行
	result, err := weather.Fn(context.Background(), params)
	if err != nil {
		fmt.Printf("  ❌ 执行失败: %v\n", err)
		return
	}
	fmt.Printf("  结果: %s\n", result.String())

	// ─── 6. 演示校验失败 ───
	fmt.Println("\n── 演示参数校验失败 ──")
	badParams := json.RawMessage(`{"city": 123}`)
	if err := weather.Validate(badParams); err != nil {
		fmt.Printf("  ❌ 预期内错误: %v\n", err)
	}

	fmt.Println("\n✅ Tool 系统工作正常")
}
