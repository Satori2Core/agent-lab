# Week 3 — 任务分解

## Task 1: Tool 结构体 + JSON Schema 生成

- 文件：`pkg/tools/tool.go`
- 内容：`Tool` struct + `NewTool[TInput, TOutput]()` 泛型构造函数
- 核心难点：从 Go struct 反射生成 JSON Schema（字段名、类型、描述、必填/可选）
- 验证：单元测试 — 定义一个 struct，检查生成的 JSON Schema 是否正确

## Task 2: ToolRegistry

- 文件：`pkg/tools/registry.go`
- 内容：`ToolRegistry` — Register/Get/List/ToOpenAI
- 验证：注册 3 个不同的 Tool，测试查重、查找、OpenAI 格式转换

## Task 3: 参数校验

- 文件：`pkg/tools/validate.go`
- 内容：`Tool.Validate()` — 将 LLM 传入的 JSON 参数与 JSON Schema 比对
- 验证：合法/非法 JSON 参数各测一组

## Task 4: 集成验证

- 文件：`cmd/service-03-tools/main.go`
- 内容：注册几个 Tool → 打印 JSON Schema → 模拟调用 → 打印结果
- 验证：`go run` 后看到工具注册、调用、返回的完整流程
