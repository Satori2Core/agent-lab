# Week 4: 记忆系统

## 要解决的问题

Agent 不是一次性调用 LLM 就结束了——它需要多步推理。每一步产生的结果需要被"记住"，以便下一步基于之前的信息做决策。

记忆系统做三件事：
1. **记录**每一步的思考、行动、观测结果
2. **回放**——将步骤历史转换为 LLM 可理解的消息列表
3. **截断**——当历史太长超过上下文窗口时，保留最相关的部分

在 smolagents 中，`AgentMemory` + 各种 `MemoryStep` 子类承担了这个角色。

## 我们要实现的 Go 版本

### 步骤类型

```go
// MemoryStep 是记忆步骤的统一接口。
type MemoryStep interface {
    Type() StepType
}

// ActionStep — 思考+行动+观测（一次 ReAct 循环）
type ActionStep struct {
    StepNumber  int
    Thought     string   // LLM 的推理过程
    Action      string   // 工具调用描述
    Observation string   // 工具返回结果
    Duration    float64  // 执行耗时(秒)
    Error       error    // 执行错误(nil=成功)
}

// PlanningStep — 规划步骤
type PlanningStep struct {
    Plan string // Agent 制定的计划
}

// SystemPromptStep — 系统提示（记忆的起点）
type SystemPromptStep struct {
    Prompt string
}

// FinalAnswerStep — 最终答案
type FinalAnswerStep struct {
    Answer string
}
```

### AgentMemory

```go
type AgentMemory struct {
    Steps       []MemoryStep
    SystemPrompt *SystemPromptStep
}

// 核心方法
func (m *AgentMemory) Record(step MemoryStep)       // 记录一步
func (m *AgentMemory) Messages() []ChatMessage      // 转为 LLM 消息列表
func (m *AgentMemory) LastAction() *ActionStep      // 最后一步行动
func (m *AgentMemory) Replay() string               // 人类可读的回放
```

### 上下文窗口管理

```go
// WithMaxMessages 限制发送给 LLM 的最大消息数。
// 当历史超过限制时，保留 system prompt + 最近的 N 条消息。
func (m *AgentMemory) MessagesWithLimit(maxMessages int) []ChatMessage
```

### 设计约束

1. **接口统一** — 所有步骤类型实现 `MemoryStep`，通过 `Type()` 区分
2. **可回放** — `Replay()` 输出完整的执行轨迹，便于调试
3. **上下文截断** — `MessagesWithLimit` 防止超出 LLM 上下文窗口
4. **与 Week 2 集成** — `Messages()` 返回 `models.ChatMessage` 列表，可直接传给 `Model.Generate()`

### 参考

- smolagents memory.py: `MemoryStep`, `ActionStep`, `AgentMemory`, `write_messages()`
