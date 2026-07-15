# Module 3: Tool 系统

## 要解决的问题

LLM 只能生成文本，无法与外部世界交互。**Tool（工具）是 Agent 的"手"**——LLM 决定调用哪个工具、传什么参数，Tool 负责执行并返回结果。

Tool 系统的核心挑战：
1. **怎么描述工具给 LLM？** → JSON Schema（function calling 标准）
2. **怎么校验 LLM 传的参数？** → 执行前验证
3. **怎么注册和管理多个工具？** → ToolRegistry
4. **怎么让工具返回的结果能被 Agent 统一处理？** → 返回 `AgentType`（Module 1）

在 smolagents 中，`@tool` 装饰器 + `ToolCollection` 承担了这个角色。

## 我们要实现的 Go 版本

### 核心结构

```go
// Tool 表示一个 Agent 可调用的工具。
// 对标 smolagents tools.py → Tool 类
type Tool struct {
    Name        string          // 工具名（LLM function calling 使用）
    Description string          // 工具描述（告诉 LLM 这个工具做什么）
    Parameters  json.RawMessage // JSON Schema（参数类型定义）
    Fn          ToolFunc        // 执行函数
}

// ToolFunc 工具执行函数的通用签名。
// 输入：json.RawMessage（LLM 传的 JSON 参数）
// 输出：types.AgentType（Module 1 的类型系统）
type ToolFunc func(ctx context.Context, input json.RawMessage) (types.AgentType, error)
```

### JSON Schema 自动生成

```go
// NewTool 创建一个类型安全的 Tool。
// TInput 必须是一个可序列化的 struct（字段名 → JSON Schema 属性）
// TOutput 必须实现 types.AgentType
//
// 示例：
//   type WeatherInput struct {
//       City string `json:"city" desc:"城市名称"`
//   }
//   tool := NewTool("get_weather", "查询天气", getWeather)
func NewTool[TInput, TOutput any](
    name string,
    description string,
    fn func(context.Context, TInput) (TOutput, error),
) (*Tool, error)
```

**关键设计**：用 Go 泛型捕获 `TInput` 的类型，用 `reflect` 自动生成 JSON Schema。
这是 Python `@tool` 装饰器在 Go 中的等价实现。

### ToolRegistry

```go
// ToolRegistry 管理一组 Tool。
// 对标 smolagents tools.py → ToolCollection
type ToolRegistry struct { ... }

func (r *ToolRegistry) Register(t *Tool) error        // 注册工具
func (r *ToolRegistry) Get(name string) (*Tool, bool)  // 按名查找
func (r *ToolRegistry) List() []*Tool                  // 列出所有工具
func (r *ToolRegistry) ToOpenAI() []map[string]any     // 转为 OpenAI function calling 格式
```

### 参数校验

```go
// Validate 校验 LLM 传入的 JSON 参数是否符合 JSON Schema。
// 对标 smolagents tools.py → validate_tool_arguments()
func (t *Tool) Validate(input json.RawMessage) error
```

### 设计约束

1. **零外部依赖** — JSON Schema 生成只用 `reflect` + `encoding/json`
2. **类型安全** — `NewTool` 用泛型在编译期检查输入输出类型
3. **与 Module 1 集成** — Tool 返回 `types.AgentType`，不是裸 `any`
4. **OpenAI 兼容** — `ToOpenAI()` 输出可直接塞进 Chat Completions API 的 `tools` 字段

### 参考

- smolagents tools.py: `Tool`, `BaseTool`, `ToolCollection`, `validate_tool_arguments()`
