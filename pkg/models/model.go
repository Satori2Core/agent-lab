// Knowledge: AGENT-MODEL-INTERFACE — 模型接口抽象
// Agent 不绑定特定 LLM 供应商，通过 interface 实现解耦。
// Reference: smolagents models.py → Model 类
package models

import (
	"context"
)

// Model 是所有 LLM 后端的统一接口。
//
// Agent 通过此接口调用 LLM，不关心底层是 OpenAI、Ollama 还是其他供应商。
// 新增供应商只需实现此接口的两个方法。
//
// 对标 smolagents models.py → Model 类
//   Python: class Model(ABC): __call__(messages) | generate(messages)
//   Go:     interface Model { Generate() | GenerateStream() }
//
// context.Context 贯穿所有方法调用：
//   - 控制超时（context.WithTimeout）
//   - 支持取消（context.WithCancel）
//   - 传递请求级别的元数据
type Model interface {
	// Generate 发送消息列表并返回完整响应。
	//
	// 这是一个同步调用——等待模型生成完整结果后返回。
	// 适用于批量处理、后台任务等不需要实时展示的场景。
	//
	// 参数：
	//   - ctx: 请求上下文（超时控制、取消信号）
	//   - messages: 对话历史消息列表
	//
	// 返回：
	//   - *Response: 模型的完整响应（文本 + 可能的工具调用）
	//   - error: 网络错误、API 错误、超时等
	Generate(ctx context.Context, messages []ChatMessage) (*Response, error)

	// GenerateStream 发送消息列表并通过 channel 返回流式增量。
	//
	// 与 Generate 不同，此方法不等待模型完全生成——
	// 每收到一个 token 就通过 channel 发送，调用者可以实时展示。
	//
	// 参数：
	//   - ctx: 请求上下文（取消 ctx 会中止流式处理）
	//   - messages: 对话历史消息列表
	//
	// 返回：
	//   - <-chan Delta: 流式增量 channel（Done=true 时表示结束）
	//   - error: 仅表示请求初始化失败；流式过程中的错误通过 Delta.Error 传递
	GenerateStream(ctx context.Context, messages []ChatMessage) (<-chan Delta, error)
}
