// Knowledge: AGENT-LOOP-REACT — 系统提示生成
// SystemPromptBuilder 从 Agent 配置和工具列表生成 LLM 系统提示。
// Reference: smolagents agents.py → PromptTemplates / system_prompt
package agent

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Satori2Core/agent-lab/pkg/tools"
)

// SystemPromptBuilder 根据 Agent 配置生成系统提示。
//
// 对标 smolagents 中的 PromptTemplates.system_prompt。
// 生成的提示包含三个部分：
//   1. Agent 的角色描述
//   2. 可用工具列表（名称+描述+参数 Schema）
//   3. 行为指令（ReAct 循环规则）
type SystemPromptBuilder struct {
	name  string
	tools *tools.ToolRegistry
}

// NewSystemPromptBuilder 创建一个新的提示构建器。
func NewSystemPromptBuilder(name string, reg *tools.ToolRegistry) *SystemPromptBuilder {
	return &SystemPromptBuilder{name: name, tools: reg}
}

// Build 生成完整的系统提示文本。
func (b *SystemPromptBuilder) Build() string {
	var sb strings.Builder

	// 1. 角色定义
	sb.WriteString(fmt.Sprintf("你是 %s，一个 AI Agent。\n", b.name))
	sb.WriteString("你的任务是通过逐步推理和调用工具来完成用户的需求。\n\n")

	// 2. 可用工具
	toolList := b.tools.List()
	if len(toolList) > 0 {
		sb.WriteString("## 可用工具\n\n")
		sb.WriteString("你可以调用以下工具。调用格式为：\n")
		sb.WriteString("```json\n{\"name\": \"工具名\", \"arguments\": {...}}\n```\n\n")

		for _, t := range toolList {
			sb.WriteString(fmt.Sprintf("### %s\n", t.Name))
			sb.WriteString(fmt.Sprintf("%s\n\n", t.Description))

			// 参数 Schema（人类可读）
			var schema map[string]any
			if err := json.Unmarshal(t.Parameters, &schema); err == nil {
				if props, ok := schema["properties"].(map[string]any); ok {
					sb.WriteString("参数:\n")
					for propName, propVal := range props {
						propMap, _ := propVal.(map[string]any)
						propType, _ := propMap["type"].(string)
						propDesc, _ := propMap["description"].(string)
						sb.WriteString(fmt.Sprintf("  - %s (%s): %s\n", propName, propType, propDesc))
					}
					if required, ok := schema["required"].([]any); ok && len(required) > 0 {
						reqNames := make([]string, len(required))
						for i, r := range required {
							reqNames[i] = r.(string)
						}
						sb.WriteString(fmt.Sprintf("  必填: %s\n", strings.Join(reqNames, ", ")))
					}
					sb.WriteString("\n")
				}
			}
		}
	}

	// 3. 行为指令
	sb.WriteString("## 行为规则\n\n")
	sb.WriteString("1. 分析用户需求，决定是否需要调用工具\n")
	sb.WriteString("2. 如果需要工具：返回 function call JSON\n")
	sb.WriteString("3. 工具执行后，你会看到结果，然后决定下一步\n")
	sb.WriteString("4. 当信息足够时，直接给出最终答案（不要继续调用工具）\n")
	sb.WriteString("5. 如果没有可用的工具能帮助回答，也直接给出答案\n")

	return sb.String()
}
