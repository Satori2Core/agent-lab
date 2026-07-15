// Knowledge: AGENT-LOOP-REACT — ReAct 核心循环
// MultiStepAgent 组装 Model + Tools + Memory，实现自主推理与行动。
// Reference: smolagents agents.py → MultiStepAgent 类
//
// 这是整个 agent-lab 项目的核心——之前四周构建的所有模块
// （类型系统、模型抽象、工具注册、记忆管理）在这里汇聚成一个
// 能真正自主运行的 Agent。
//
// ReAct 循环:
//   用户任务 → Agent 思考 → 调用工具 → 观测结果 → 再思考 → ... → 回答
//               ↑_______________________________________________↓
//                       循环直到任务完成或超步数
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Satori2Core/agent-lab/pkg/memory"
	"github.com/Satori2Core/agent-lab/pkg/models"
	"github.com/Satori2Core/agent-lab/pkg/tools"
)

// RunResult 包含一次 Agent 运行的完整结果。
//
// 对标 smolagents agents.py → RunResult 类。
type RunResult struct {
	// Answer Agent 的最终回答
	Answer string
	// Steps 执行的步数
	Steps int
	// Duration 总耗时
	Duration time.Duration
	// Memory 完整的执行轨迹（可回放）
	Memory *memory.AgentMemory
}

// MultiStepAgent 是一个基于 ReAct 循环的自主 Agent。
//
// 它组合了三个核心模块：
//   - Model（Week 2）：调用 LLM 进行推理
//   - ToolRegistry（Week 3）：提供可调用的工具
//   - AgentMemory（Week 4）：记录执行历史和上下文
//
// 对标 smolagents agents.py → MultiStepAgent 类
//   Python: class MultiStepAgent(ABC): run(), step(), _run()
//   Go:     type MultiStepAgent struct { Run(), RunStream() }
type MultiStepAgent struct {
	model    models.Model
	tools    *tools.ToolRegistry
	maxSteps int
	name     string
	prompt   *SystemPromptBuilder
}

// AgentOption 是 MultiStepAgent 的配置函数。
type AgentOption func(*MultiStepAgent)

// WithMaxSteps 设置最大执行步数（默认 10）。
func WithMaxSteps(n int) AgentOption {
	return func(a *MultiStepAgent) { a.maxSteps = n }
}

// WithSystemPrompt 设置自定义系统提示（覆盖默认的 prompt builder）。
func WithSystemPrompt(builder *SystemPromptBuilder) AgentOption {
	return func(a *MultiStepAgent) { a.prompt = builder }
}

// NewMultiStepAgent 创建一个新的 MultiStepAgent。
//
// 参数：
//   - name: Agent 的名称（用于系统提示）
//   - model: LLM 模型（实现 models.Model 接口）
//   - reg: 工具注册中心
//   - opts: 可选的配置函数
func NewMultiStepAgent(name string, model models.Model, reg *tools.ToolRegistry, opts ...AgentOption) *MultiStepAgent {
	a := &MultiStepAgent{
		model:    model,
		tools:    reg,
		maxSteps: 10,
		name:     name,
		prompt:   NewSystemPromptBuilder(name, reg),
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Run 执行 Agent 的主循环。
//
// 这是 Agent 的核心方法——将用户任务作为输入，
// 通过 ReAct 循环调用 LLM 和工具，最终返回答案。
//
// 参数：
//   - ctx: 上下文（超时控制、取消信号）
//   - task: 用户的任务描述
//
// 返回：
//   - *RunResult: 包含答案、步数、耗时、完整记忆
//   - error: 仅在达到最大步数仍无答案或模型调用失败时返回
func (a *MultiStepAgent) Run(ctx context.Context, task string) (*RunResult, error) {
	startTime := time.Now()
	mem := memory.NewAgentMemory()

	// Step 0: 初始化记忆——写入系统提示和用户任务
	sysPrompt := a.prompt.Build()
	mem.Record(memory.NewSystemPromptStep(sysPrompt))
	mem.Record(memory.NewActionStep(0, "收到任务", "无", task, 0, nil))

	// 用户消息单独添加（不作为 MemoryStep，直接拼到 Messages 末尾）
	userMsg := models.ChatMessage{Role: models.RoleUser, Content: task}
	messages := mem.Messages()
	messages = append(messages, userMsg)

	// 主循环
	for step := 1; step <= a.maxSteps; step++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("Agent %q: 上下文已取消: %w", a.name, ctx.Err())
		default:
		}

		// 1. 调用模型（传入工具定义，支持 function calling）
		response, err := a.model.Generate(ctx, messages, a.tools.ToOpenAI())
		if err != nil {
			return nil, fmt.Errorf("Agent %q: 模型调用失败 (step %d): %w", a.name, step, err)
		}

		// 2. 检查是否需要调用工具
		if len(response.ToolCalls) > 0 {
			// 执行工具调用
			tc := response.ToolCalls[0] // 目前只处理第一个 tool call

			tool, ok := a.tools.Get(tc.Function.Name)
			if !ok {
				obs := fmt.Sprintf("错误: 工具 %q 不存在", tc.Function.Name)
				mem.Record(memory.NewActionStep(step, response.Content, tc.Function.Name, obs, 0, fmt.Errorf("%s", obs)))
			} else {
				// 校验参数
				params := json.RawMessage(tc.Function.Arguments)
				if err := tool.Validate(params); err != nil {
					obs := fmt.Sprintf("参数校验失败: %v", err)
					mem.Record(memory.NewActionStep(step, response.Content, tc.Function.Name, obs, 0, err))
				} else {
					// 执行工具
					result, err := tool.Fn(ctx, params)
					duration := time.Since(startTime).Seconds()
					if err != nil {
						mem.Record(memory.NewActionStep(step, response.Content, tc.Function.Name, "", duration, err))
					} else {
						obs := result.String()
						mem.Record(memory.NewActionStep(step, response.Content, tc.Function.Name, obs, duration, nil))
					}
				}
			}

			// 更新 messages——添加 tool 调用结果
			messages = mem.Messages()
			messages = append(messages, userMsg)
		} else {
			// 3. 模型返回纯文本 = 最终答案
			mem.Record(memory.NewFinalAnswerStep(response.Content))
			return &RunResult{
				Answer:   response.Content,
				Steps:    step,
				Duration: time.Since(startTime),
				Memory:   mem,
			}, nil
		}
	}

	// 达到最大步数
	return nil, fmt.Errorf("Agent %q: 达到最大步数 %d，未能完成任务", a.name, a.maxSteps)
}

// RunStream 执行 Agent 主循环（流式版本）。
//
// 与 Run() 的区别：每完成一步，通过 channel 发送 StreamEvent，
// 调用者可以实时看到 Agent 的思考过程。
//
// 参数：
//   - ctx: 上下文
//   - task: 用户任务
//
// 返回：
//   - <-chan StreamEvent: 事件流（Done=true 时结束）
//   - error: 仅初始化失败
func (a *MultiStepAgent) RunStream(ctx context.Context, task string) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent, 10)
	go func() {
		defer close(ch)
		result, err := a.Run(ctx, task)
		if err != nil {
			ch <- StreamEvent{Type: EventError, Error: err, Done: true}
			return
		}
		ch <- StreamEvent{Type: EventComplete, Result: result, Done: true}
	}()
	return ch, nil
}

// StreamEvent 是流式执行中的一个事件。
type StreamEvent struct {
	Type   EventType
	Result *RunResult
	Error  error
	Done   bool
}

// EventType 表示流式事件的类型。
type EventType string

const (
	EventComplete EventType = "complete"
	EventError    EventType = "error"
)
