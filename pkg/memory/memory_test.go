// Knowledge: AGENT-MEMORY-CTX — 上下文管理
// 测试 AgentMemory：记录、回放、消息生成、截断。
// Reference: smolagents memory.py → AgentMemory 类
package memory

import (
	"strings"
	"testing"
)

// TestMemoryRecord 验证步骤记录和回放。
func TestMemoryRecord(t *testing.T) {
	m := NewAgentMemory()

	m.Record(NewSystemPromptStep("你是一个天气助手"))
	m.Record(NewActionStep(1, "查天气", "get_weather", "晴天", 0.5, nil))
	m.Record(NewFinalAnswerStep("今天晴天"))

	if m.StepCount() != 3 {
		t.Errorf("StepCount = %d, want 3", m.StepCount())
	}

	replay := m.Replay()
	if !strings.Contains(replay, "天气助手") {
		t.Error("Replay 应包含 SystemPrompt")
	}
	if !strings.Contains(replay, "get_weather") {
		t.Error("Replay 应包含 Action")
	}
	if !strings.Contains(replay, "今天晴天") {
		t.Error("Replay 应包含 FinalAnswer")
	}
}

// TestMemoryLastAction 验证获取最后一步 ActionStep。
func TestMemoryLastAction(t *testing.T) {
	m := NewAgentMemory()
	m.Record(NewActionStep(1, "t1", "a1", "o1", 0.1, nil))
	m.Record(NewActionStep(2, "t2", "a2", "o2", 0.2, nil))

	last := m.LastAction()
	if last.StepNumber != 2 {
		t.Errorf("last StepNumber = %d, want 2", last.StepNumber)
	}
}

// TestMemoryLastActionNone 验证没有 ActionStep 时返回 nil。
func TestMemoryLastActionNone(t *testing.T) {
	m := NewAgentMemory()
	if m.LastAction() != nil {
		t.Error("空记忆中 LastAction() 应返回 nil")
	}
}

// TestMemoryMessages 验证步骤转为 ChatMessage。
func TestMemoryMessages(t *testing.T) {
	m := NewAgentMemory()
	m.Record(NewSystemPromptStep("你是一个助手"))
	m.Record(NewPlanningStep("1.查天气 2.给建议"))
	m.Record(NewActionStep(1, "查天气", "get_weather(Beijing)", "晴天", 0.5, nil))
	m.Record(NewFinalAnswerStep("适合外出"))

	msgs := m.Messages()
	if len(msgs) == 0 {
		t.Fatal("Messages() 不应为空")
	}

	// 第一条应是 system 消息
	if string(msgs[0].Role) != "system" {
		t.Errorf("第一条消息 role = %s, want system", msgs[0].Role)
	}
}

// TestMemoryMessagesWithLimit 验证上下文截断。
func TestMemoryMessagesWithLimit(t *testing.T) {
	m := NewAgentMemory()
	m.Record(NewSystemPromptStep("sys"))
	// 模拟很多步骤
	for i := 1; i <= 10; i++ {
		m.Record(NewActionStep(i, "think", "act", "obs", 0.1, nil))
	}
	m.Record(NewFinalAnswerStep("done"))

	// 限制 5 条消息
	msgs := m.MessagesWithLimit(5)
	if len(msgs) > 5 {
		t.Errorf("MessagesWithLimit(5) = %d messages, want <= 5", len(msgs))
	}

	// system prompt 应该始终保留
	if string(msgs[0].Role) != "system" {
		t.Error("截断后第一条消息应是 system prompt")
	}
}

// TestMemoryReset 验证重置。
func TestMemoryReset(t *testing.T) {
	m := NewAgentMemory()
	m.Record(NewActionStep(1, "t", "a", "o", 0.1, nil))
	m.Reset()

	if m.StepCount() != 0 {
		t.Errorf("Reset 后 StepCount = %d, want 0", m.StepCount())
	}
}

// TestMemoryFinalAnswer 验证获取最终答案。
func TestMemoryFinalAnswer(t *testing.T) {
	m := NewAgentMemory()
	m.Record(NewFinalAnswerStep("答案是42"))

	answer, ok := m.FinalAnswer()
	if !ok {
		t.Fatal("应该能找到 FinalAnswer")
	}
	if answer != "答案是42" {
		t.Errorf("answer = %q", answer)
	}
}
