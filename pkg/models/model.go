// Knowledge: AGENT-MODEL-INTERFACE — 模型接口抽象
// Agent 不绑定特定 LLM 供应商，通过 interface 实现解耦。
// Reference: smolagents models.py → Model 类
package models

import "context"

// Model 是所有 LLM 后端的统一接口。
//
// 对标 smolagents models.py → Model 类
type Model interface {
	// Generate 发送消息列表和工具定义，返回完整响应。
	//
	// tools 参数对应 OpenAI function calling 的 tools 字段——
	// 告诉模型有哪些工具可用。传 nil 表示不使用工具。
	Generate(ctx context.Context, messages []ChatMessage, tools []map[string]any) (*Response, error)

	// GenerateStream 流式版本，参数同 Generate。
	GenerateStream(ctx context.Context, messages []ChatMessage, tools []map[string]any) (<-chan Delta, error)
}
