// Knowledge: AGENT-MEMORY-STEP — 步骤追踪
// 定义 Agent 执行过程中的各类步骤类型。
// Reference: smolagents memory.py → MemoryStep / ActionStep / PlanningStep 等
//
// 对标关系：
//   Python: class MemoryStep / ActionStep(MemoryStep) / PlanningStep(MemoryStep) ...
//   Go:     interface MemoryStep + struct ActionStep/PlanningStep/... + StepType enum
//
// 在 Python 中通过继承区分步骤类型，Go 中用 interface + Type() 方法。
package memory

// StepType 表示记忆步骤的类型。
type StepType string

const (
	// StepAction — ReAct 循环中的一个执行步骤（思考→行动→观测）
	StepAction StepType = "action"
	// StepPlanning — Agent 制定的计划
	StepPlanning StepType = "planning"
	// StepFinalAnswer — Agent 的最终答案
	StepFinalAnswer StepType = "final_answer"
	// StepSystemPrompt — 系统的初始提示（记忆的起点）
	StepSystemPrompt StepType = "system_prompt"
)

// MemoryStep 是所有记忆步骤必须实现的接口。
//
// 对标 smolagents memory.py → MemoryStep 基类。
// 在 Python 中通过 isinstance() 区分步骤类型，Go 中用 Type() 方法。
type MemoryStep interface {
	// Type 返回步骤的类型标识。
	Type() StepType
}

// ActionStep 表示 ReAct 循环中的一个完整步骤。
//
// 包含 Agent 的思考过程（Thought）、执行的动作（Action）、
// 以及工具返回的观测结果（Observation）。
// 对标 smolagents memory.py → ActionStep 类。
type ActionStep struct {
	// StepNumber 步骤编号（从 1 开始）
	StepNumber int
	// Thought LLM 的推理过程（"我需要查询天气..."）
	Thought string
	// Action 工具调用的描述（"调用 get_weather(city=Beijing)"）
	Action string
	// Observation 工具返回的结果（"晴天 22°C"）
	Observation string
	// Duration 步骤执行耗时（秒）
	Duration float64
	// Error 执行过程中的错误（nil = 成功）
	Error error
}

// NewActionStep 创建一个新的 ActionStep。
func NewActionStep(num int, thought, action, observation string, duration float64, err error) *ActionStep {
	return &ActionStep{
		StepNumber:  num,
		Thought:     thought,
		Action:      action,
		Observation: observation,
		Duration:    duration,
		Error:       err,
	}
}

// Type 返回 StepAction（实现 MemoryStep 接口）。
func (s *ActionStep) Type() StepType { return StepAction }

// IsSuccess 返回此步骤是否成功执行。
func (s *ActionStep) IsSuccess() bool { return s.Error == nil }

// PlanningStep 表示 Agent 制定的计划。
//
// 对标 smolagents memory.py → PlanningStep 类。
// Agent 在执行前可能会先制定一个计划，记录在 PlanningStep 中。
type PlanningStep struct {
	// Plan 计划的文本内容
	Plan string
}

// NewPlanningStep 创建一个新的 PlanningStep。
func NewPlanningStep(plan string) *PlanningStep {
	return &PlanningStep{Plan: plan}
}

// Type 返回 StepPlanning（实现 MemoryStep 接口）。
func (s *PlanningStep) Type() StepType { return StepPlanning }

// FinalAnswerStep 表示 Agent 的最终答案。
//
// 这是 Agent 执行结束的标志——当 Agent 决定"任务已完成"时，
// 会记录一个 FinalAnswerStep 并退出循环。
// 对标 smolagents memory.py → FinalAnswerStep 类。
type FinalAnswerStep struct {
	// Answer 最终答案的文本内容
	Answer string
}

// NewFinalAnswerStep 创建一个新的 FinalAnswerStep。
func NewFinalAnswerStep(answer string) *FinalAnswerStep {
	return &FinalAnswerStep{Answer: answer}
}

// Type 返回 StepFinalAnswer（实现 MemoryStep 接口）。
func (s *FinalAnswerStep) Type() StepType { return StepFinalAnswer }

// SystemPromptStep 表示系统提示——记忆的起点。
//
// 对标 smolagents memory.py → SystemPromptStep 类。
// 这是记忆中的第一条记录，定义了 Agent 的角色和行为规则。
type SystemPromptStep struct {
	// Prompt 系统提示的文本内容
	Prompt string
}

// NewSystemPromptStep 创建一个新的 SystemPromptStep。
func NewSystemPromptStep(prompt string) *SystemPromptStep {
	return &SystemPromptStep{Prompt: prompt}
}

// Type 返回 StepSystemPrompt（实现 MemoryStep 接口）。
func (s *SystemPromptStep) Type() StepType { return StepSystemPrompt }
