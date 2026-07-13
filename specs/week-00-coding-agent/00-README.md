# Week 0: 构建"编码 Agent"（元工具层）

> **一句话**：设计一套 Agent 工程约束，让 AI 编码过程变得自主可控。
>
> **核心产出**：`CLAUDE.md`（System Prompt） + `comment-check`（验证 Tool） + Hook（反馈闭环）

---

## 📖 阅读指引

按编号顺序阅读，每份文档有明确的读者和目的：

| 编号 | 文档 | 是什么 | 适合谁 | 读完能回答 |
|------|------|--------|--------|-----------|
| 00 | `00-README.md` | **本文档，阅读指引** | 所有人 | 这些文档是什么？按什么顺序读？ |
| 01 | `SPEC.md` | 设计文档 | 想理解"为什么要做这个" | 编码 Agent 要解决什么问题？对标哪些 Agent 知识点？ |
| 02 | `TASKS.md` | 任务分解 | 想了解执行过程 | 分了几个 Task？每步做什么？完成标准是什么？ |
| 03 | `PRACTICE_GUIDE.md` | 实践指南 | 想深入理解原理 | 每个组件怎么工作？什么可控什么不可控？和 smolagents 怎么对标？ |
| 04 | `04-VERIFICATION.md` | 工程验收手册 | 要验证这套系统 | 怎么一键验证？每项检查测什么？预期结果是什么？怎么解读？ |
| 05 | `05-EXPERIMENT.md` | 行为实验 | 想验证 Agent 行为来源 | AI 写代码是系统约束还是恰好听话？反馈闭环是否有效？ |

---

## 🗺️ 推荐阅读路径

```
快速了解（10 分钟）
  00-README.md  →  01-SPEC.md
                           ↓
深入理解（30 分钟）
  02-TASKS.md  →  03-PRACTICE_GUIDE.md
                           ↓
动手验证（20 分钟）
  04-VERIFICATION.md  →  05-EXPERIMENT.md
```

---

## 📂 项目级产出

这些文件在 agent-lab 项目根目录，不在 specs/ 下：

| 文件 | 作用 | 对标 Agent 概念 |
|------|------|----------------|
| `CLAUDE.md` | 编码 Agent 的 System Prompt | System Prompt 设计 |
| `KNOWLEDGE_MAP.md` | 知识点 ↔ 代码映射表 | RAG 知识注入 |
| `cmd/comment-check/main.go` | 注释检查工具源码 | Tool 设计 |
| `.claude/settings.json` | 自动验证 Hook | Observation 反馈 |
| `PRACTICE.md` | 7 Week 总路线图 | Memory 记忆 |

---

## 🔗 关联

- 项目路线图：[PRACTICE.md](../../PRACTICE.md)
- 知识映射：[KNOWLEDGE_MAP.md](../../KNOWLEDGE_MAP.md)
- 编码 Agent 规则：[CLAUDE.md](../../CLAUDE.md)
