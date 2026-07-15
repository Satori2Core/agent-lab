// Knowledge: AGENT-MEMORY-STEP — 步骤追踪
// 测试 MemoryStep 接口和各步骤类型。
// Reference: smolagents memory.py → MemoryStep / ActionStep / PlanningStep 等
package memory

import (
	"errors"
	"testing"
)

// TestMemoryStepTypes 验证每种步骤类型的 Type() 和字段。
func TestMemoryStepTypes(t *testing.T) {
	// ActionStep
	action := NewActionStep(1, "我需要查询天气", "get_weather", "晴天 22°C", 0.5, nil)
	if action.Type() != StepAction {
		t.Errorf("ActionStep.Type() = %v", action.Type())
	}
	if action.StepNumber != 1 {
		t.Errorf("StepNumber = %d", action.StepNumber)
	}

	// ActionStep with error
	errAction := NewActionStep(2, "重试", "get_weather", "", 1.0, errors.New("timeout"))
	if errAction.Error == nil {
		t.Error("error should be set")
	}

	// PlanningStep
	plan := NewPlanningStep("1. 查天气 2. 根据天气建议穿搭")
	if plan.Type() != StepPlanning {
		t.Errorf("PlanningStep.Type() = %v", plan.Type())
	}

	// FinalAnswerStep
	answer := NewFinalAnswerStep("今天北京晴天，适合外出")
	if answer.Type() != StepFinalAnswer {
		t.Errorf("FinalAnswerStep.Type() = %v", answer.Type())
	}

	// SystemPromptStep
	sys := NewSystemPromptStep("你是一个有用的助手")
	if sys.Type() != StepSystemPrompt {
		t.Errorf("SystemPromptStep.Type() = %v", sys.Type())
	}
}

// TestMemoryStepInterface 验证所有类型都实现了 MemoryStep 接口。
func TestMemoryStepInterface(t *testing.T) {
	steps := []MemoryStep{
		NewActionStep(1, "think", "act", "obs", 0.1, nil),
		NewPlanningStep("plan"),
		NewFinalAnswerStep("answer"),
		NewSystemPromptStep("system"),
	}
	for i, s := range steps {
		if s.Type() == "" {
			t.Errorf("step[%d] has empty type", i)
		}
	}
}

// compileTimeCheck
var (
	_ MemoryStep = (*ActionStep)(nil)
	_ MemoryStep = (*PlanningStep)(nil)
	_ MemoryStep = (*FinalAnswerStep)(nil)
	_ MemoryStep = (*SystemPromptStep)(nil)
)
