// Knowledge: AGENT-MEMORY-CTX — 记忆系统集成验证
// 模拟完整的 Agent 多步执行 → 回放轨迹 → 生成 LLM 消息。
//
// 用法:
//   go run ./cmd/service-04-memory/
package main

import (
	"encoding/json"
	"fmt"

	"github.com/Satori2Core/agent-lab/pkg/memory"
)

func main() {
	fmt.Println("Module 4: 记忆系统 — 集成验证")
	fmt.Println("============================================")

	// ─── 模拟 Agent 执行轨迹 ───
	mem := memory.NewAgentMemory()

	// 1. 记录系统提示
	mem.Record(memory.NewSystemPromptStep(
		"你是一个天气助手。当用户问天气时，先用 get_weather 查询，再根据结果给出穿搭建议。",
	))

	// 2. 记录计划
	mem.Record(memory.NewPlanningStep(
		"1. 调用 get_weather 查询北京天气\n2. 根据天气给出穿搭建议\n3. 输出最终答案",
	))

	// 3. Step 1: 查询天气
	mem.Record(memory.NewActionStep(
		1,
		"用户想知道北京今天天气如何，我需要调用 get_weather 工具。",
		"get_weather(city=\"Beijing\")",
		"北京今天晴天，22-28°C，微风",
		0.5,
		nil,
	))

	// 4. Step 2: 给穿搭建议
	mem.Record(memory.NewActionStep(
		2,
		"天气是晴天22-28°C，适合穿轻薄的衣服。我可以给出具体建议。",
		"无（根据已有信息直接回答）",
		"",
		0.1,
		nil,
	))

	// 5. 最终答案
	mem.Record(memory.NewFinalAnswerStep(
		"北京今天晴天，22-28°C，微风。建议穿短袖或薄衬衫，带一件薄外套以防早晚温差。适合户外活动。",
	))

	// ─── 展示回放 ───
	fmt.Println(mem.Replay())

	// ─── 展示 LLM 消息 ───
	fmt.Println("── LLM 消息列表（完整）──")
	msgs := mem.Messages()
	for i, msg := range msgs {
		content := msg.Content
		if len(content) > 80 {
			content = content[:80] + "..."
		}
		fmt.Printf("  [%d] %s: %s\n", i, msg.Role, content)
	}
	fmt.Printf("  总计 %d 条消息\n\n", len(msgs))

	// ─── 展示上下文截断 ───
	fmt.Println("── 上下文截断演示（限制 3 条）──")
	truncated := mem.MessagesWithLimit(3)
	for i, msg := range truncated {
		content := msg.Content
		if len(content) > 60 {
			content = content[:60] + "..."
		}
		fmt.Printf("  [%d] %s: %s\n", i, msg.Role, content)
	}
	fmt.Printf("  截断后: %d 条消息（保留 system + 最近 2 条）\n\n", len(truncated))

	// ─── 展示步骤统计 ───
	fmt.Println("── 步骤统计 ──")
	fmt.Printf("  总步骤数: %d\n", mem.StepCount())
	if last := mem.LastAction(); last != nil {
		fmt.Printf("  最后行动: Step %d — %s\n", last.StepNumber, last.Thought[:min(30, len(last.Thought))]+"...")
	}
	if answer, ok := mem.FinalAnswer(); ok {
		answerJSON, _ := json.Marshal(answer)
		fmt.Printf("  最终答案: %s\n", string(answerJSON))
	}

	fmt.Println("\n✅ 记忆系统工作正常")
}
