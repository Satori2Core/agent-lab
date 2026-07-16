# Module 7 — 任务分解

## Task 1: CodeExecutor

- 文件：`pkg/agent/codeagent/executor.go`
- 内容：`CodeExecutor` 接口 + `LocalExecutor`（Python + Shell）
- 验证：执行一段简单 Python 代码，检查输出

## Task 2: execute_code Tool

- 文件：`pkg/agent/codeagent/tool.go`
- 内容：注册 `execute_code` 工具到 ToolRegistry
- 验证：手动调用工具，检查代码执行和错误捕获

## Task 3: 集成验证

- 文件：`cmd/agent-server/` 增加 `/execute` 端点
- 或直接修改 demo：让 Agent 自主写代码解数学题
- 验证：Agent 写出代码 → 执行 → 如果错了自己修正
