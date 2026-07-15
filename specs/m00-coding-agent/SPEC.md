# Module 0: 构建"编码 Agent"（元工具层）

## 要解决的问题

直接将 agent-lab 的代码编写任务交给 AI，会出现：
- 方法没有注释，不知道在干什么
- 写了代码但不理解对应的 Agent 知识点
- AI 自由发挥，产出不可控

**Module 0 的目标**：设计一套 Agent 工程约束，让 AI 编码过程变得**自主可控**。

## 对标 Agent 知识点

| 我们要构建的 | 对应的 Agent 概念 | 在 smolagents 中的对应 |
|-------------|-----------------|---------------------|
| `CLAUDE.md` 规则体系 | **System Prompt** — 定义 Agent 的角色和行为边界 | `PromptTemplates` — `system_prompt` 字段 |
| `KNOWLEDGE_MAP.md` | **知识注入 / RAG** — 给 Agent 注入领域知识 | `memory.py` — `SystemPromptStep` |
| `comment-check` 工具 | **Tool** — Agent 可调用的验证能力 | `tools.py` — `Tool` 类 |
| `.claude/settings.json` | **Hook / Observation** — Agent 行动后的自动反馈 | `MultiStepAgent.step()` 中的 observation 环节 |
| `SPEC.md` + `TASKS.md` | **Planning** — 任务分解与执行追踪 | `PlanningStep` — 规划步骤 |

## 交付物

1. **编码 Agent 的系统提示** → `CLAUDE.md`
2. **知识映射模板** → `KNOWLEDGE_MAP.md`
3. **注释检查工具** → `cmd/comment-check/main.go`
4. **自动验证 Hook** → `.claude/settings.json`
5. **Module 0 的 SPEC + TASKS** → `specs/m00-coding-agent/`

## 工作流设计

```
用户说"开始 Module N Task M"
        │
        ▼
┌─────────────────────────────────────────────┐
│  编码 Agent                                  │
│                                             │
│  1. 读取 SPEC.md      ← 理解接口契约       │
│  2. 读取 TASKS.md     ← 拆解执行步骤       │
│  3. 查阅 KNOWLEDGE_MAP ← 确认知识点映射     │
│  4. 写测试            ← TDD（先写验证）     │
│  5. 写实现            ← CLAUDE.md 约束注释  │
│  6. 保存文件           ← 触发 Hook          │
│  7. comment-check 验证 ← 自动反馈           │
│  8. 修正 → 通过        ← Observation 闭环  │
└─────────────────────────────────────────────┘
```

## 设计原则

- **约束而非建议** — CLAUDE.md 中的规则是强制执行的，不是"建议"
- **可验证** — 每条规则都有对应的自动检查手段
- **可追溯** — 每段代码都能回溯到对应的 Agent 知识点
