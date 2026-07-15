# Agent Lab — 实践路线图

> 通过亲手构建 smolagents 的 Go 实现，渐进掌握 Agent 工程核心知识。
>
> 关联笔记项目：[LearnAgent](https://github.com/Satori2Core/LearnAgent)

## 总览

| Week | 模块 | 路径 | 核心概念 | 状态 |
|------|------|------|---------|------|
| 0 | 构建编码 Agent | `CLAUDE.md` + `cmd/comment-check/` | System Prompt、Tool、Hook、RAG | ✅ |
| 1 | Agent 多模态类型系统 | `pkg/types/` | interface 多态、AgentType 抽象 | ✅ |
| 2 | LLM 模型抽象层 | `pkg/models/` | 模型接口、流式输出、多供应商 | ✅ |
| 3 | 工具系统 | `pkg/tools/` | Tool 定义/注册/调用、JSON Schema | ✅ |
| 4 | 记忆系统 | `pkg/memory/` | 步骤追踪、上下文管理 | ✅ |
| 5 | ReAct 核心循环 | `pkg/agent/` | 思考-行动-观察循环 | ✅ |
| 6 | Agent HTTP 服务化 | `cmd/agent-server/` | SSE 流式、生产部署 | ⬜ |
| 7 | 代码执行 Agent | `pkg/agent/codeagent/` | 沙箱执行、安全隔离 | ⬜ |

## 元层：Agent 工程实践

每个 Service 使用 Agent 工程模式来约束 AI 辅助编码：

| Agent 概念 | 对应文件/实践 | 说明 |
|-----------|-------------|------|
| System Prompt | `CLAUDE.md` | 编码 Agent 的角色、规则、工作流程 |
| Tool | `cmd/comment-check/` | 给 Agent 一个可调用的 godoc 验证工具 |
| Planning | `specs/<week>/TASKS.md` | 拆成可验证的小步 |
| Observation | `*_test.go` + Hook | 测试验证 + 保存文件自动检查 |
| RAG | `KNOWLEDGE_MAP.md` | Agent 知识点注入到编码上下文 |
| Memory | `PRACTICE.md` 进度表 | 追踪已完成/未完成的 Week 和 Task |

## 开始

```bash
cd agent-lab
go test ./...                 # 运行所有测试
go run ./cmd/comment-check/   # 手动运行注释检查
codegraph status              # 查看代码图谱
```
