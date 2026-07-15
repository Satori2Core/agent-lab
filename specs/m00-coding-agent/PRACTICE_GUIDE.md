# Module 0 实践指南：构建编码 Agent

## 一、我们做了什么？

### 四个交付物

```
┌──────────────────────────────────────────────────────────┐
│                     编码 Agent 系统                       │
│                                                          │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────┐  │
│  │ CLAUDE.md   │  │ KNOWLEDGE_   │  │ comment-check  │  │
│  │             │  │ MAP.md       │  │                │  │
│  │ 定义规则:    │  │ 知识点映射:   │  │ 系统级验证:     │  │
│  │ "必须写注释" │  │ "这段代码对应  │  │ 解析 Go AST    │  │
│  │ "测试先行"  │  │  AGENT-TOOL"  │  │ 检查公开符号    │  │
│  │             │  │              │  │ 报告违规        │  │
│  └──────┬──────┘  └──────┬───────┘  └───────┬────────┘  │
│         │                │                   │           │
│         │                │                   │           │
│    ┌────▼────────────────▼───────────────────▼──────┐    │
│    │              Hook (settings.json)               │    │
│    │  保存文件 → PostToolUse → 运行 comment-check    │    │
│    │  违规 → 输出到聊天 → AI 看到 → 修改 → 再保存   │    │
│    └────────────────────────────────────────────────┘    │
│                                                          │
│        反馈闭环: AI 写代码 → 工具检查 → 违规 → 修正      │
└──────────────────────────────────────────────────────────┘
```

---

## 二、每个组件的工作原理

### 组件 A：CLAUDE.md（软约束 — 靠 AI 自觉）

**本质是什么？**

Claude Code 启动时，读取项目根目录的 `CLAUDE.md`，将其内容追加到系统提示中。
AI 模型看到这些规则后，在生成回答时"参考"它们。

**工作方式：**

```
你发送消息 "帮我写 AgentText"
    ↓
Claude Code 读取 CLAUDE.md
    ↓
系统提示 = (默认提示) + (CLAUDE.md 内容)
    ↓
AI 看到 "每个公开符号必须有 godoc 注释"
    ↓
AI 生成代码时参考这条规则 ← 但 AI 可能忽略它
```

**局限**：**没有任何机制强制 AI 遵守。** AI 可以因为上下文太长、注意力分散、或模型切换而"忘记"规则。这是 **prompt engineering**，不是 **agent engineering**。

---

### 组件 B：comment-check（硬约束 — 系统级验证）

**本质是什么？**

一个编译好的 Go 程序。它用 `go/parser` 解析 Go 源文件的 AST，遍历所有公开符号，检查是否有 godoc 注释。**它不依赖 AI、不依赖 LLM、不依赖网络。** 它是确定性的：同样的输入永远产生同样的输出。

**工作方式：**

```
comment-check ./pkg/types/agent_text.go
    ↓
解析 AST → 找到 type AgentText → 检查有没有注释
    ↓                           ↓
    有 → 注释以 "AgentText" 开头？ → 有 → 通过
                                → 无 → 报违规: "godoc 注释应以 AgentText 开头"
    ↓
    没有 → 报违规: "缺少 godoc 注释"

退出码 = 违规数量（0 = 通过）
```

**为什么这是"系统级"？** 因为它可以被任何东西调用——你可以手动运行、CI 可以运行、Git Hook 可以运行、另一个程序可以调用它。它是真正独立于 AI Chat 的。

---

### 组件 C：KNOWLEDGE_MAP.md（软约束 — 知识注入）

与 CLAUDE.md 相同的机制：通过 Claude Code 注入到聊天上下文。
给 AI 提供领域知识（Agent 理论概念），让 AI 能理解代码与理论之间的映射。

**同样有局限：AI 可能忽略它、记错它、或在长对话中遗忘。**

---

### 组件 D：settings.json Hook（混合 — 系统触发 + AI 感知）

**工作方式：**

```
AI 调用 Write 工具保存文件
    ↓
PostToolUse Hook 触发 ← 系统级，不由 AI 控制
    ↓
运行: go run ./cmd/comment-check/ -- 刚才保存的文件路径
    ↓
两种情况:
  ┌─ 通过 → 无输出 → AI 不知道（Hook 不产生聊天消息）
  └─ 违规 → 错误信息输出到聊天 → AI 看到 → 修正
```

**Hook 本身的触发是系统级的**（Claude Code 框架执行），但**修正行为仍然依赖 AI 的回应**。

---

## 三、诚实评估：什么可控？什么不可控？

| 行为 | 可控性 | 说明 |
|------|--------|------|
| comment-check 检查结果 | ✅ 完全可控 | 确定性程序，可单独运行 |
| Hook 的触发时机 | ✅ 可控 | Claude Code 框架保证 PostToolUse 必触发 |
| AI 是否写 godoc | ⚠️ 间接可控 | 不可直接控制 AI 行为，但可通过反馈闭环纠正 |
| AI 是否映射知识点 | ⚠️ 间接可控 | 依赖 AI 遵守 CLAUDE.md + KNOWLEDGE_MAP.md |
| AI 是否先写测试 | ⚠️ 间接可控 | 目前没有系统级工具强制"测试要先于实现" |
| comment-check 的检测范围 | ✅ 可控 | 可以扩展（检查 Knowledge 注释、测试文件等） |

**核心结论**：真正 100% 可控的是 **comment-check 工具**和 **Hook 触发机制**。
AI 行为的合规性，是通过 **"违规 → 反馈 → 修正"** 的闭环来实现的，而不是通过"AI 听话"。

---

## 四、与 smolagents 的对标

我们构建的编码 Agent 本质上是在复现 `MultiStepAgent` 的架构：

```
smolagents                        Agent Lab (编码 Agent)
─────────                        ──────────────────────
SystemPromptStep                  CLAUDE.md
  (Agent 的初始状态和规则)           (编码规则和角色定义)

PlanningStep                      SPEC.md + TASKS.md
  (拆解任务)                        (接口契约 + 任务分解)

ActionStep                        AI 写代码 (Write/Edit 工具)
  ("执行工具调用")                   (生成 .go 文件)

Observation                       comment-check 输出
  ("观测工具返回结果")               (违规报告)

AgentMemory                       PRACTICE.md + KNOWLEDGE_MAP.md
  (记忆步骤)                        (进度追踪 + 知识映射)

MultiStepAgent.run()              完整的编码工作流
  (循环: 思考→行动→观测→重复)       (写→检查→修正→通过)
```

**但有一个关键差异**：smolagents 的 Agent 是通过代码逻辑（`while` 循环）实现的，而我们的编码 Agent 部分依赖 AI 的"自觉"。Module 5 我们实现 `MultiStepAgent` 时，这个差异会消除——我们会写出真正由代码驱动循环的 Agent。

---

## 五、一键验证

见 `verify.ps1`（同目录下）。

运行方式：

```powershell
# 在 agent-lab 项目根目录
.\specs\m00-coding-agent\verify.ps1
```

它会：
1. 检查 comment-check 工具是否能编译
2. 用 comment-check 自检（工具检查自己的代码）
3. 检查项目结构是否完整
4. 输出通过/失败

**这个验证脚本不依赖任何 AI —— 你可以关掉 AI，在终端直接运行。**
