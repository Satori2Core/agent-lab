# 实验：编码 Agent 行为来源验证

> **核心问题**：编码 Agent 的行为——
> 是系统组件（comment-check）强制产生的？
> 还是 AI Chat "恰好听话"？

---

## 实验 A：系统独立性 ✅

### 问题
comment-check 工具是否独立于 AI Chat 工作？

### 操作
```bash
bash specs/week-00-coding-agent/verify.sh
```

### 结果（已执行）
```
✅ 23/23 检查通过
```

### 解读
comment-check 是编译好的 Go 程序。它用 `go/parser` 解析 AST → 遍历公开符号 → 检查注释格式。
**整个过程不调用 LLM、不读 CLAUDE.md、不依赖 AI。**
你可以关掉 AI Chat，在终端直接跑，结果一样。

**结论**：验证能力来自 comment-check 程序本身，不是 AI Chat。

---

## 实验 B：AI 行为合规性 ✅

### 问题
在 CLAUDE.md 规则约束下，AI 写出的代码是否自动符合 godoc 规范？

### 操作
1. 在 **agent-lab 项目**的 Claude Code 中发送：

```
在 specs/week-00-coding-agent/experiment/with_agent/ 目录下，
写一个包 greeter，新建 greeter.go，包含以下公开符号：
- Greeter 结构体，字段 Name string  
- Greet() string 方法，返回 "Hello, {Name}!"
- NewGreeter(name string) *Greeter 构造函数

遵循 CLAUDE.md 规范（godoc 注释 + Knowledge 映射）。
```

2. AI 完成后，**在终端**（不用 AI）运行：

```bash
go run ./cmd/comment-check/ -- specs/week-00-coding-agent/experiment/with_agent/
```

### 结果

| 项目 | 值 |
|------|-----|
| 时间 | 2026-07-13 |
| 退出码 | 0 |
| 违规数 | 0 |
| 结论 | ✅ AI 在 CLAUDE.md 约束下产出了合规代码 |

### 解读

**CLAUDE.md 规则在此次会话中生效了。** AI 写出的 greeter.go 中所有公开符号都有合规的 godoc 注释。

**但这是一个"软约束"的成功案例，不是系统保证。** AI 可能因为以下原因遵守规则：
- 会话刚开始，上下文清晰，规则在注意力范围内
- 任务简单（一个文件、四个公开符号）
- 模型恰好"愿意配合"

**如果换成以下场景，结果可能不同：**
- 长对话后上下文被稀释
- 多文件批量修改
- 不同的 AI 模型

**关键认知**：CLAUDE.md 能让 AI "知道该做什么"，但不能保证 AI "一定会做"。

---

## 实验 C：反馈闭环 ✅

### 问题
当 AI 产出违规代码时，comment-check 能否驱动 AI 修正？
AI 是"知道怎么写对的"还是"靠工具引导才能写对"？

### 材料
`experiment/feedback_loop/bad.go` — 4 处故意违规：

| 符号 | 违规类型 | 原始注释 |
|------|---------|---------|
| `Calculator` | 格式错误 | `// 违规1: 公开类型，没有 godoc 注释` |
| `NewCalculator` | 格式错误 | `// 违规2: 公开函数，没有 godoc 注释` |
| `Add` | 格式错误 | `// 违规3: 有注释但格式错误` |
| `Result` | 格式错误 | `// 违规4: 有注释但格式错误` |

### 操作

**第 1 轮 — 触发违规**：AI 运行 comment-check，看到 4 个违规。

**第 2 轮 — AI 修正**：要求 AI "只根据违规报告修改，不要读 CLAUDE.md，不要用你自己的代码审查能力判断"。

**第 3 轮 — 验证**：再次运行 comment-check → 0 违规。

### 结果

| 项目 | 值 |
|------|-----|
| 修正轮数 | 1 轮 |
| 最终是否通过 | ✅ 0 违规 |
| AI 只用违规报告？ | ❌ 否 |
| AI 实际行为 | **选择了"直接重写整个文件"，而不是逐行修正违规** |

### 🔴 关键发现：工具可靠性决定了 Agent 行为边界

AI 的完整行为日志揭示了一个更精确的真相：

```
阶段 1 — Agent 模式（精准修正，按工具反馈逐行操作）
  Edit 修正 Calculator   → ✅ 成功
  Edit 修正 NewCalculator → ✅ 成功
  Edit 修正 Add          → ❌ 失败（tab 字符匹配问题）
  Edit 修正 Add（重试）   → ❌ 失败
  Edit 修正 Result       → ❌ 失败

阶段 2 — 工具降级
  AI: "Edit 工具有问题，直接重写整个文件"
  Write 全量覆盖 → ✅ 成功

阶段 3 — 验证
  comment-check → exit 0 → ✅ 全部通过
```

**AI 最初确实按照 Agent 模式行动**——它读了违规报告的每一行，用行号定位问题，用 Edit 逐行修改。前两个成功了，后两个因为 Edit 工具对 tab 字符的精确匹配问题而失败。

**重写整个文件不是 AI 的"哲学选择"，而是工具失败后的唯一出路。** 当精确的 Edit 工具不可靠时，AI 降到唯一能用的大粒度工具（Write）。

**这意味着什么：**

```
Agent 闭环失败的原因        之前认为的            实际情况
───────────────────      ──────────────       ──────────────
AI 主动绕过工具反馈        ❌ 不准确             AI 最初确实按工具反馈逐行操作
AI 自主决定重写            ❌ 不准确             AI 在 Edit 失败后才降级到 Write
工具可靠性不足             ✅ 真正原因           Edit 工具对 tab/空格敏感
```

**修正后的认知：**

| | 发现 | 工程意义 |
|---|---|---|
| AI 愿意做 Agent | ✅ AI 天然会按工具反馈逐行操作 | 不需要"逼"AI 当 Agent |
| 工具可靠性 | ⚠️ Edit 工具在 tab 匹配时脆弱 | **Agent 的工具必须可靠，否则回路断裂** |
| 降级行为 | Write 全量覆盖 = 大粒度工具 = 副作用风险 | Agent 需要"原子性"工具（单文件修改不重写） |
| 核心教训 | Agent 行为边界 = 工具的可靠性边界 | **Tool 设计即 Agent 设计** |

### 解读

**这不是"AI 不是 Agent"的问题，而是"Agent 的工具不够可靠"的问题。**

AI 在实验 C 中表现出了天然的 Agent 行为倾向：读反馈 → 定位问题 → 逐项修正。但当工具（Edit）因为技术原因（tab 匹配）失败时，回路就断了。

**这对 agent-lab 项目的核心启示变了：**

Week 3（Tool 系统）和 Week 5（Agent Loop）的关联远比我们想的更紧密：
- Agent 的循环逻辑（Week 5）依赖于工具（Week 3）的可靠性
- 如果 Tool 不可靠，Agent Loop 再精致也没用
- **一个好的 Tool 必须：** ① 输入输出明确 ② 原子操作 ③ 不依赖上下文猜测

```go
// 这个认知会直接影响 Week 3 的 Tool 接口设计：
// Tool 不能只是 "Execute(input) → output"
// 还必须有错误分类，让 Agent 知道"重试"还是"降级"：

type ToolResult struct {
    Output   any
    Error    error
    Retryable bool   // ← 关键：告诉 Agent 这个错误可以重试吗？
}
```

---

## 三实验总结

| 实验 | 验证内容 | 结果 | 关键发现 |
|------|---------|------|---------|
| A | 系统能不能独立验证？ | ✅ 23/23 通过 | comment-check 是独立工具，不依赖 AI |
| B | AI 在规则下能否产出合规代码？ | ✅ 0 违规 | **单次有效**，但 CLAUDE.md 是软约束，不保证每次生效 |
| C | 工具反馈能否驱动 AI 修正？ | ⚠️ 通过，但工具可靠性是关键瓶颈 | AI 最初按 Agent 模式逐行操作，但 Edit 工具的 tab 匹配问题导致降级到 Write。**Agent 行为边界 = 工具的可靠性边界。** |

### 核心认知

```
真正的 Agent 系统 = 硬约束（工具/程序） + 软约束（提示/规则）

硬约束：100% 可控、可复现、可独立验证
软约束：依赖 AI 理解能力，有概率失效

两者的边界在实验 C 中被重新揭示：
  AI 天然愿意按工具反馈逐行操作（Agent 行为）
  但工具不可靠时（Edit tab 匹配失败），回路断裂
  AI 被迫降级到唯一可用的大粒度工具（Write）

结论：Tool 设计即 Agent 设计。
     工具有多可靠，Agent 就能多可控。
```

### 实验复现

```bash
# 验证实验 C 的最终结果（bad.go 已被修正）
go run ./cmd/comment-check/ -- specs/week-00-coding-agent/experiment/feedback_loop/bad.go
# 预期：退出码 0，无输出

# 重置实验材料（恢复 4 处违规）后重新做实验 C：
# bad.go 的原始版本在 git 中可以恢复：
# git checkout -- specs/week-00-coding-agent/experiment/feedback_loop/bad.go
```
