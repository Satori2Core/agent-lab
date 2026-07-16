// Knowledge: AGENT-CODE-SANDBOX — 代码执行工具
// 将 CodeExecutor 包装为 Agent 可调用的 Tool，实现"生成代码→执行→观测→修正"循环。
// Reference: smolagents agents.py → CodeAgent 类
package codeagent

import (
	"context"
	"fmt"

	"github.com/Satori2Core/agent-lab/pkg/tools"
	"github.com/Satori2Core/agent-lab/pkg/types"
)

// ExecuteCodeInput 是 execute_code 工具的参数。
type ExecuteCodeInput struct {
	// Language 代码语言（python / shell）
	Language string `json:"language" desc:"代码语言: python 或 shell"`
	// Code 要执行的代码内容
	Code string `json:"code" desc:"要执行的代码"`
}

// NewExecuteCodeTool 创建一个 execute_code 工具。
//
// 这个工具让 Agent 能够：
//   1. 生成代码
//   2. 在沙箱中执行
//   3. 观察输出（包括错误信息）
//   4. 根据反馈修正代码
//
// 对标 smolagents CodeAgent 的核心能力——
// 区别是 Go 版本通过 Tool 扩展实现，而非继承新的 Agent 类。
func NewExecuteCodeTool(executor CodeExecutor) (*tools.Tool, error) {
	return tools.NewTool(
		"execute_code",
		"在安全的沙箱环境中执行代码。支持 python 和 shell。代码在隔离的临时目录中运行，有超时保护。将 stdout 和 stderr 返回。如果代码有错误，返回错误信息帮助你修正。",
		func(ctx context.Context, input ExecuteCodeInput) (*types.AgentText, error) {
			output, err := executor.Execute(ctx, input.Language, input.Code)
			if err != nil {
				// 返回错误信息（让 LLM 看到并修正代码）
				return types.NewAgentText(fmt.Sprintf("代码执行出错:\n%s", err.Error()))
			}
			if output == "" {
				return types.NewAgentText("代码执行成功，无输出。")
			}
			return types.NewAgentText(output)
		},
	)
}
