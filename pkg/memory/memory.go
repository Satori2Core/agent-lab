// Knowledge: AGENT-MEMORY-CTX — 上下文管理
// AgentMemory 管理 Agent 的执行历史：记录步骤、回放轨迹、生成 LLM 消息。
// Reference: smolagents memory.py → AgentMemory 类
//
// 对标关系：
//   Python: AgentMemory.steps + write_messages() → LLM 消息列表
//   Go:     AgentMemory.Record() + Messages() → []ChatMessage
//
// 记忆系统的核心职责：
//   1. 记录每一步（思考→行动→观测）
//   2. 将步骤历史转为 LLM 可读的消息列表
//   3. 当历史过长时截断，保留最相关的部分
package memory

import (
	"fmt"
	"strings"

	"github.com/Satori2Core/agent-lab/pkg/models"
)

// AgentMemory 管理 Agent 的执行历史和上下文。
//
// Agent 每执行一步，就通过 Record() 将步骤记录到记忆中。
// 下一步执行时，Messages() 将记忆转为 LLM 输入消息，
// 让模型知道"之前发生了什么"。
//
// 对标 smolagents memory.py → AgentMemory 类
//   Python: AgentMemory(system_prompt=..., steps=[])
//     方法: replay(), write_messages(), step_number
//   Go:
//     方法: Record(), Replay(), Messages(), MessagesWithLimit()
type AgentMemory struct {
	steps        []MemoryStep
	systemPrompt *SystemPromptStep
}

// NewAgentMemory 创建一个空的记忆实例。
func NewAgentMemory() *AgentMemory {
	return &AgentMemory{
		steps: make([]MemoryStep, 0),
	}
}

// Record 记录一个步骤到记忆中。
//
// 如果步骤是 SystemPromptStep，会同时保存为系统提示的引用。
func (m *AgentMemory) Record(step MemoryStep) {
	if sys, ok := step.(*SystemPromptStep); ok {
		m.systemPrompt = sys
	}
	m.steps = append(m.steps, step)
}

// StepCount 返回记忆中的步骤总数。
func (m *AgentMemory) StepCount() int {
	return len(m.steps)
}

// LastAction 返回最新的 ActionStep（如果存在），否则返回 nil。
//
// 用于 Agent 循环中判断"上一步发生了什么"。
func (m *AgentMemory) LastAction() *ActionStep {
	for i := len(m.steps) - 1; i >= 0; i-- {
		if action, ok := m.steps[i].(*ActionStep); ok {
			return action
		}
	}
	return nil
}

// FinalAnswer 返回最终答案（如果存在）。
//
// 返回：
//   - 答案文本和 true（找到了 FinalAnswerStep）
//   - "" 和 false（还没生成最终答案）
func (m *AgentMemory) FinalAnswer() (string, bool) {
	for _, step := range m.steps {
		if fa, ok := step.(*FinalAnswerStep); ok {
			return fa.Answer, true
		}
	}
	return "", false
}

// Reset 清空所有记忆。
func (m *AgentMemory) Reset() {
	m.steps = make([]MemoryStep, 0)
	m.systemPrompt = nil
}

// Replay 生成人类可读的完整执行轨迹。
//
// 对标 smolagents AgentMemory.replay() —— 用于调试和日志。
// 输出格式为 Markdown 文本，可直接打印或写日志。
func (m *AgentMemory) Replay() string {
	var b strings.Builder
	b.WriteString("## Agent 执行轨迹\n\n")

	for _, step := range m.steps {
		switch s := step.(type) {
		case *SystemPromptStep:
			b.WriteString("### 系统提示\n\n")
			b.WriteString(s.Prompt)
			b.WriteString("\n\n")

		case *PlanningStep:
			b.WriteString("### 计划\n\n")
			b.WriteString(s.Plan)
			b.WriteString("\n\n")

		case *ActionStep:
			b.WriteString(fmt.Sprintf("### Step %d\n\n", s.StepNumber))
			b.WriteString(fmt.Sprintf("**思考**: %s\n\n", s.Thought))
			b.WriteString(fmt.Sprintf("**行动**: %s\n\n", s.Action))
			if s.Error != nil {
				b.WriteString(fmt.Sprintf("**错误**: %v\n\n", s.Error))
			} else {
				b.WriteString(fmt.Sprintf("**观测**: %s\n\n", s.Observation))
			}

		case *FinalAnswerStep:
			b.WriteString("### 最终答案\n\n")
			b.WriteString(s.Answer)
			b.WriteString("\n\n")
		}
	}

	return b.String()
}

// Messages 将记忆转为 LLM 消息列表（无截断）。
//
// 对标 smolagents memory.py → write_messages()。
// 转换规则：
//   SystemPromptStep → system 角色消息
//   PlanningStep      → assistant 角色消息（计划内容）
//   ActionStep        → assistant 消息（思考+行动）+ tool 消息（观测）
//   FinalAnswerStep   → assistant 角色消息（答案）
func (m *AgentMemory) Messages() []models.ChatMessage {
	return m.MessagesWithLimit(0) // 0 表示不限制
}

// MessagesWithLimit 将记忆转为 LLM 消息列表，限制最大消息数。
//
// 当 maxMessages > 0 且消息总数超过限制时，采用"保留头部+尾部"策略：
//   - 始终保留 system prompt（第一条）
//   - 保留最近的 maxMessages-1 条消息
//   - 中间的旧消息被丢弃
//
// 这种策略保证：LLM 知道自己的角色（system prompt）和最近的上下文，
// 同时不超出上下文窗口限制。
//
// 参数：
//   - maxMessages: 最大消息数，0 表示不限制
func (m *AgentMemory) MessagesWithLimit(maxMessages int) []models.ChatMessage {
	messages := m.buildMessages()

	if maxMessages <= 0 || len(messages) <= maxMessages {
		return messages
	}

	// 保留头部 system prompt（1条）+ 尾部最近的 N-1 条
	result := make([]models.ChatMessage, 0, maxMessages)
	result = append(result, messages[0])                     // system prompt
	tail := messages[len(messages)-(maxMessages-1):]          // 最近的 N-1 条
	result = append(result, tail...)

	return result
}

// buildMessages 执行步骤到 ChatMessage 的转换逻辑。
func (m *AgentMemory) buildMessages() []models.ChatMessage {
	messages := make([]models.ChatMessage, 0)

	for _, step := range m.steps {
		switch s := step.(type) {
		case *SystemPromptStep:
			messages = append(messages, models.ChatMessage{
				Role:    models.RoleSystem,
				Content: s.Prompt,
			})

		case *PlanningStep:
			messages = append(messages, models.ChatMessage{
				Role:    models.RoleAssistant,
				Content: fmt.Sprintf("计划:\n%s", s.Plan),
			})

		case *ActionStep:
			// 思考+行动 → assistant 消息
			msgContent := fmt.Sprintf("思考: %s\n行动: %s", s.Thought, s.Action)
			messages = append(messages, models.ChatMessage{
				Role:    models.RoleAssistant,
				Content: msgContent,
			})

			// 观测 → tool 消息
			obsContent := s.Observation
			if s.Error != nil {
				obsContent = fmt.Sprintf("错误: %v", s.Error)
			}
			messages = append(messages, models.ChatMessage{
				Role:    models.RoleTool,
				Content: obsContent,
			})

		case *FinalAnswerStep:
			messages = append(messages, models.ChatMessage{
				Role:    models.RoleAssistant,
				Content: s.Answer,
			})
		}
	}

	return messages
}
