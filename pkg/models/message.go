// Knowledge: AGENT-MODEL-MESSAGE — 消息结构
// Agent 与 LLM 交互的消息建模：system/user/assistant/tool 四种角色。
// Reference: smolagents models.py → ChatMessage / MessageRole / ToolCall 类
package models

// MessageRole 表示对话消息的角色类型。
//
// 对标 smolagents models.py → MessageRole(str, Enum)
//   Python: class MessageRole(str, Enum): SYSTEM="system", USER="user", ...
//   Go:     type MessageRole string + const
type MessageRole string

const (
	// RoleSystem 系统消息——定义 Agent 的行为规则和背景信息。
	RoleSystem MessageRole = "system"
	// RoleUser 用户消息——用户输入的指令或问题。
	RoleUser MessageRole = "user"
	// RoleAssistant 助手消息——LLM 返回的回复。
	RoleAssistant MessageRole = "assistant"
	// RoleTool 工具消息——Tool 执行后的返回结果。
	RoleTool MessageRole = "tool"
)

// IsValid 检查角色类型是否有效。
func (r MessageRole) IsValid() bool {
	switch r {
	case RoleSystem, RoleUser, RoleAssistant, RoleTool:
		return true
	default:
		return false
	}
}

// ChatMessage 表示一条对话消息。
//
// 对标 smolagents models.py → ChatMessage 类（TypedDict）。
// 这是 OpenAI Chat Completions API 消息格式的 Go 表示。
type ChatMessage struct {
	// Role 消息角色：system / user / assistant / tool
	Role MessageRole `json:"role"`
	// Content 消息内容（文本）
	Content string `json:"content"`
	// Name 发送者名称（仅 role=tool 时用于标识工具名）
	Name string `json:"name,omitempty"`
	// ToolCallID 关联的工具调用 ID（仅 role=tool 时使用）
	ToolCallID string `json:"tool_call_id,omitempty"`
}

// ToolCall 表示模型请求的工具调用（function calling）。
//
// 对标 smolagents models.py → ChatMessageToolCall 类。
// 当 LLM 决定调用工具时，响应中会包含一个或多个 ToolCall。
type ToolCall struct {
	// ID 工具调用的唯一标识（由模型生成）
	ID string `json:"id"`
	// Function 调用的函数信息
	Function FunctionCall `json:"function"`
}

// FunctionCall 表示一个具体的函数调用。
type FunctionCall struct {
	// Name 函数名
	Name string `json:"name"`
	// Arguments 函数参数（JSON 字符串）
	Arguments string `json:"arguments"`
}

// Response 表示一次模型调用的完整响应。
//
// 对标 smolagents 中 Model.generate() 的返回值。
// 包含模型生成的文本和/或工具调用请求。
type Response struct {
	// Content 模型生成的文本内容
	Content string `json:"content"`
	// ToolCalls 模型请求的工具调用列表（可能为空）
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	// FinishReason 完成原因：stop / tool_calls / length
	FinishReason string `json:"finish_reason,omitempty"`
}

// Delta 表示流式输出中的一个增量。
//
// 对标 smolagents models.py → ChatMessageStreamDelta。
// 通过 channel 传递给调用者，支持实时展示。
type Delta struct {
	// Content 本次增量的文本内容
	Content string `json:"content,omitempty"`
	// ToolCalls 增量中的工具调用信息
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	// Done 流是否结束
	Done bool `json:"done"`
	// Error 流处理中的错误（仅 Done=true 时可能非 nil）
	Error error `json:"-"`
}
