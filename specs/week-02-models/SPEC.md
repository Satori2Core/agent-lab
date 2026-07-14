# Week 2: LLM 模型抽象层

## 要解决的问题

Agent 需要调用 LLM，但不应绑定特定供应商。今天用 OpenAI，明天可能换成 Anthropic 或本地模型。**模型抽象层**让 Agent 框架与具体 LLM 实现解耦。

在 smolagents 中，`Model` 抽象类和 `ApiModel` 承担了这个角色——定义统一的生成接口，由子类实现不同供应商的 HTTP 调用。

## 我们要实现的 Go 版本

### 核心接口

```go
// Model 是所有 LLM 后端的统一接口。
// 对标 smolagents models.py → Model 类
type Model interface {
    // Generate 发送消息列表，返回完整响应。
    Generate(ctx context.Context, messages []ChatMessage, tools []Tool) (*Response, error)

    // GenerateStream 发送消息列表，通过 channel 返回流式增量。
    GenerateStream(ctx context.Context, messages []ChatMessage, tools []Tool) (<-chan Delta, error)
}
```

### 数据结构

```go
// MessageRole 消息角色（对标 smolagents MessageRole 枚举）
type MessageRole string
const (
    RoleSystem    MessageRole = "system"
    RoleUser      MessageRole = "user"
    RoleAssistant MessageRole = "assistant"
    RoleTool      MessageRole = "tool"
)

// ChatMessage 一条对话消息（对标 smolagents ChatMessage）
type ChatMessage struct {
    Role    MessageRole
    Content string
    Name    string     // 工具名（仅 RoleTool 时使用）
}

// Response 模型完整响应（对标 smolagents 返回的 ChatMessage）
type Response struct {
    Content   string     // 模型生成的文本
    ToolCalls []ToolCall // 模型请求的工具调用
}

// ToolCall 模型请求的工具调用
type ToolCall struct {
    ID        string
    Name      string
    Arguments string // JSON 格式的参数
}

// Delta 流式增量
type Delta struct {
    Content   string // 本次增量文本
    Done      bool   // 是否完成
    ToolCalls []ToolCall
}
```

### 具体实现

| 类型 | 说明 |
|------|------|
| `OpenAIModel` | OpenAI 兼容 API（支持任意 base URL，复用于 vLLM/Ollama 等） |

smolagents 有 10+ 种 Model 子类。我们只实现 `OpenAIModel`——它是 80% 场景下实际使用的，且 OpenAI 兼容 API 是事实标准（Ollama、vLLM、LocalAI 等都兼容）。

### 设计约束

1. **零依赖** — HTTP 调用只用 `net/http`，JSON 序列化用 `encoding/json`
2. **context 贯穿** — 所有方法接收 `context.Context`，支持超时和取消
3. **流式用 channel** — `GenerateStream` 返回 `<-chan Delta`，Go 原生并发模型
4. **可测试** — OpenAIModel 接受 `*http.Client` 注入，方便 mock

### 参考

- smolagents: src/smolagents/models.py (2102 行)
- 我们聚焦: Model 接口 + ChatMessage + OpenAIModel
- 不实现: VLLM/MLX/Transformers/LiteLLM 等（它们本质上都是 HTTP 调用，OpenAI 兼容）
