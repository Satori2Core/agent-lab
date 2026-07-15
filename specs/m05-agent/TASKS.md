# Module 5 — 任务分解

## Task 1: MultiStepAgent 结构 + Run()

- 文件：`pkg/agent/agent.go`
- 内容：`MultiStepAgent` struct + `Run()` 方法 — 核心 ReAct 循环
- 验证：mock Model 测试完整循环

## Task 2: 系统提示生成

- 文件：`pkg/agent/prompt.go`
- 内容：从工具列表生成 System Prompt（角色+工具描述+使用说明）
- 验证：测试生成的 prompt 包含工具名和描述

## Task 3: 集成验证

- 文件：`cmd/service-05-agent/main.go`
- 内容：用真实的 DeepSeek API + 注册几个工具 → 让 Agent 完成一个任务
- 验证：Agent 自主调用工具、返回正确答案（第一个真正可工作的 Agent！）
