# Week 4 — 任务分解

## Task 1: 步骤类型

- 文件：`pkg/memory/step.go`
- 内容：`MemoryStep` 接口 + `ActionStep`/`PlanningStep`/`FinalAnswerStep`/`SystemPromptStep`
- 验证：测试各类型的 `Type()` 方法和字段

## Task 2: AgentMemory

- 文件：`pkg/memory/memory.go`
- 内容：`AgentMemory` — Record/LastAction/Replay/Reset
- 验证：模拟 3 步 Agent 循环，测试回放内容

## Task 3: 上下文管理

- 文件：`pkg/memory/memory.go`
- 内容：`Messages()` → 将步骤转为 `models.ChatMessage` 列表
- `MessagesWithLimit(max)` → 上下文窗口截断
- 验证：测试消息生成和截断逻辑

## Task 4: 集成验证

- 文件：`cmd/service-04-memory/main.go`
- 内容：模拟完整的 Agent 执行轨迹 → 回放 → 生成 LLM 消息
- 验证：`go run` 后看到人类可读的执行历史和 LLM 可用的消息列表
