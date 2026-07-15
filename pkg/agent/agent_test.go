// Knowledge: AGENT-LOOP-REACT — ReAct 循环测试
// 用 mock Model 测试 MultiStepAgent 的核心循环。
// Reference: smolagents agents.py → MultiStepAgent.run()
package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/Satori2Core/agent-lab/pkg/models"
	"github.com/Satori2Core/agent-lab/pkg/tools"
	"github.com/Satori2Core/agent-lab/pkg/types"
)

// mockReActModel 模拟 ReAct 循环中的 LLM 行为。
// 第一次调用返回 tool call，第二次返回最终答案。
type mockReActModel struct {
	callCount int
}

func (m *mockReActModel) Generate(ctx context.Context, messages []models.ChatMessage, tools []map[string]any) (*models.Response, error) {
	m.callCount++
	if m.callCount == 1 {
		// 第一次：请求调用工具
		return &models.Response{
			Content: "",
			ToolCalls: []models.ToolCall{{
				ID: "call-1",
				Function: models.FunctionCall{
					Name:      "get_weather",
					Arguments: `{"city":"Beijing"}`,
				},
			}},
		}, nil
	}
	// 第二次及以后：返回最终答案
	return &models.Response{
		Content:      "北京今天晴天，适合外出。",
		FinishReason: "stop",
	}, nil
}

func (m *mockReActModel) GenerateStream(ctx context.Context, messages []models.ChatMessage, tools []map[string]any) (<-chan models.Delta, error) {
	return nil, nil
}

// compile-time check
var _ models.Model = (*mockReActModel)(nil)

// TestAgentRunWithTool 验证完整的 ReAct 循环（调用工具 → 获取答案）。
func TestAgentRunWithTool(t *testing.T) {
	// 注册工具
	reg := tools.NewToolRegistry()
	type WeatherInput struct {
		City string `json:"city" desc:"城市名"`
	}
	weatherTool, _ := tools.NewTool("get_weather", "查询天气",
		func(ctx context.Context, input WeatherInput) (*types.AgentText, error) {
			return types.NewAgentText(input.City + ": 晴天 22°C")
		})
	reg.Register(weatherTool)

	// 创建 Agent（mock model + 真实工具）
	model := &mockReActModel{}
	agent := NewMultiStepAgent("天气助手", model, reg, WithMaxSteps(5))

	// 执行
	result, err := agent.Run(context.Background(), "北京今天天气怎么样？")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if !strings.Contains(result.Answer, "晴天") {
		t.Errorf("Answer = %q, 应包含 '晴天'", result.Answer)
	}
	if result.Steps != 2 {
		t.Errorf("Steps = %d, want 2（1次工具调用 + 1次最终答案）", result.Steps)
	}
	if result.Memory == nil {
		t.Fatal("Memory 不应为 nil")
	}

	// 验证记忆中有 ActionStep 和 FinalAnswerStep
	_, hasAnswer := result.Memory.FinalAnswer()
	if !hasAnswer {
		t.Error("记忆应包含 FinalAnswer")
	}
	if result.Memory.StepCount() < 3 {
		t.Errorf("StepCount = %d, want >= 3", result.Memory.StepCount())
	}
}

// TestAgentRunNoTool 验证不需要工具时的直答行为。
func TestAgentRunNoTool(t *testing.T) {
	// mock model 直接返回答案（不调用工具）
	directModel := &struct{ mockReActModel }{}
	// 重置 callCount 的副作用——第一次就返回答案
	directModel.callCount = 0
	// 覆盖 Generate 让它直接返回答案
	// 实际上 mockReActModel 第二次才返回答案，这里测试无工具场景直接用另一个 mock

	directAnswerModel := &directAnswerModel{}
	agent := NewMultiStepAgent("助手", directAnswerModel, tools.NewToolRegistry())
	result, err := agent.Run(context.Background(), "你好")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if result.Answer == "" {
		t.Error("Answer 不应为空")
	}
}

// directAnswerModel 直接返回答案的 mock。
type directAnswerModel struct{}

func (m *directAnswerModel) Generate(ctx context.Context, messages []models.ChatMessage, tools []map[string]any) (*models.Response, error) {
	return &models.Response{Content: "你好！有什么可以帮助你的？", FinishReason: "stop"}, nil
}
func (m *directAnswerModel) GenerateStream(ctx context.Context, messages []models.ChatMessage, tools []map[string]any) (<-chan models.Delta, error) {
	return nil, nil
}
var _ models.Model = (*directAnswerModel)(nil)

// TestAgentSystemPrompt 验证系统提示包含工具信息。
func TestAgentSystemPrompt(t *testing.T) {
	reg := tools.NewToolRegistry()
	type SearchInput struct {
		Query string `json:"query" desc:"搜索关键词"`
	}
	searchTool, _ := tools.NewTool("search", "搜索互联网", func(ctx context.Context, input SearchInput) (*types.AgentText, error) {
		return types.NewAgentText("搜索结果")
	})
	reg.Register(searchTool)

	builder := NewSystemPromptBuilder("测试助手", reg)
	prompt := builder.Build()

	if !strings.Contains(prompt, "search") {
		t.Error("prompt 应包含工具名 'search'")
	}
	if !strings.Contains(prompt, "搜索互联网") {
		t.Error("prompt 应包含工具描述")
	}
	if !strings.Contains(prompt, "query") {
		t.Error("prompt 应包含参数名 'query'")
	}
}

// TestAgentMaxSteps 验证最大步数限制。
func TestAgentMaxSteps(t *testing.T) {
	// mock 始终返回 tool call → Agent 永远无法完成
	loopModel := &loopModel{}
	agent := NewMultiStepAgent("loop", loopModel, tools.NewToolRegistry(), WithMaxSteps(3))

	_, err := agent.Run(context.Background(), "测试")
	if err == nil {
		t.Error("达到 maxSteps 应返回错误")
	}
}

type loopModel struct{}

func (m *loopModel) Generate(ctx context.Context, messages []models.ChatMessage, tools []map[string]any) (*models.Response, error) {
	return &models.Response{
		ToolCalls: []models.ToolCall{{ID: "x", Function: models.FunctionCall{Name: "nonexistent"}}},
	}, nil
}
func (m *loopModel) GenerateStream(ctx context.Context, messages []models.ChatMessage, tools []map[string]any) (<-chan models.Delta, error) {
	return nil, nil
}
var _ models.Model = (*loopModel)(nil)
