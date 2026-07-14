# Week 2 — 任务分解

## Task 1: 定义 ChatMessage 和辅助类型

- 文件：`pkg/models/message.go`
- 内容：`MessageRole`、`ChatMessage`、`Response`、`ToolCall`、`Delta`
- 验证：编译通过 + 测试结构体字段

## Task 2: 定义 Model 接口

- 文件：`pkg/models/model.go`
- 内容：`Model` interface（`Generate` + `GenerateStream`）
- 验证：编译通过 + mock 实现满足接口

## Task 3: 实现 OpenAIModel

- 文件：`pkg/models/openai.go`
- 内容：`OpenAIModel` struct — 构造 OpenAI 兼容的 HTTP 请求，解析响应
- 支持自定义 BaseURL（用于 Ollama/vLLM 等）
- 注入 `*http.Client`（支持测试 mock）
- 验证：单元测试 + 环境变量 `OPENAI_API_KEY`

## Task 4: 实现 GenerateStream 流式输出

- 文件：`pkg/models/openai.go`
- 内容：SSE 解析，channel 返回增量
- 验证：流式测试（mock HTTP server）

## Task 5: 集成验证

- 文件：`cmd/service-02-models/main.go`
- 内容：调用 OpenAI API（需要 API Key），打印完整和流式响应
- 验证：`go run` 后看到模型输出
