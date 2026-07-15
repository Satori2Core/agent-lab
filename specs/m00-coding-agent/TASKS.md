# Module 0 — 任务分解

## Task 1: 设计编码 Agent 的 System Prompt ✅

- 文件：`CLAUDE.md`
- 内容：定义编码 Agent 的角色、强制规则、工作流程、禁止行为
- 对标知识点：`AGENT-SYS-PROMPT` — System Prompt 设计
- 验证：人工审查规则是否完整、可执行

## Task 2: 设计知识映射模板 ✅

- 文件：`KNOWLEDGE_MAP.md`
- 内容：定义每个 Module 的 Agent 知识点 ↔ 代码映射关系
- 对标知识点：`AGENT-RAG` — 知识注入
- 验证：每个 Module 至少 3 个知识点条目

## Task 3: 实现注释检查工具 ✅

- 文件：`cmd/comment-check/main.go`
- 内容：解析 Go AST，检查公开符号是否有 godoc 注释
- 对标知识点：`AGENT-TOOL` — Tool 设计
- 验证：工具自检通过（`go run ./cmd/comment-check/ -- cmd/comment-check/main.go`）

## Task 4: 配置自动验证 Hook ✅

- 文件：`.claude/settings.json`
- 内容：PostToolUse Hook — 写文件后自动运行 comment-check
- 对标知识点：`AGENT-OBSERVE` — Observation 反馈回路
- 验证：下次写 .go 文件时自动触发（重启 Claude Code 后生效）

## 完成标准

- [x] CLAUDE.md 定义了 Agent 的完整行为规则
- [x] KNOWLEDGE_MAP.md 覆盖了 Module 0-7 的知识点
- [x] comment-check 工具可运行且通过自检
- [x] Hook 配置完成，下次启动生效
- [ ] 用户在 agent-lab 目录新开 VSCode 窗口，Module 0 结束
