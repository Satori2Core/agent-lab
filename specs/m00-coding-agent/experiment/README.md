# 编码 Agent 行为实验

## 核心问题

**编码 Agent（CLAUDE.md + comment-check + Hook）的行为来源是什么？**

- 是系统组件（comment-check）强制产生的？
- 还是 AI Chat 根据 CLAUDE.md "恰好听话"？

## 实验设计

三个实验，顺序执行：

### 实验 A：系统组件独立性（不需要 AI）

**问题**：comment-check 工具是否独立工作？

**步骤**：
```bash
# 运行验证脚本 — 全程不需要 AI
bash specs/m00-coding-agent/verify.sh
```

**预期结果**：23 项检查全部通过（我们在上一步已验证）

**解释**：comment-check 工具是编译好的 Go 程序，它解析 AST、检查注释、报告违规。
这个过程不调用 LLM、不依赖网络、不读 CLAUDE.md。它是**确定性的系统组件**。

**结论**：验证能力来自于 comment-check 程序本身，不是 AI Chat。

---

### 实验 B：AI 行为合规性（需要 AI）

**问题**：在 CLAUDE.md 规则约束下，AI 写出的代码是否自动符合 godoc 规范？

**步骤**：
1. 在 agent-lab 项目目录下，对 AI 说：
   > "在 specs/m00-coding-agent/experiment/with_agent/ 目录下，
   > 写一个包 `greeter`，包含：
   > - 一个公开类型 `Greeter`，有一个字段 `Name string`
   > - 一个公开方法 `Greet() string`，返回 "Hello, {Name}!"
   > - 一个公开函数 `NewGreeter(name string) *Greeter`
   > 遵循 CLAUDE.md 规范。"

2. AI 生成代码后，**手动运行**（不用 AI）：
   ```bash
   go run ./cmd/comment-check/ -- specs/m00-coding-agent/experiment/with_agent/
   ```

3. 记录结果：
   - 如果退出码 = 0：AI 在规则约束下产出了合规代码
   - 如果退出码 > 0：AI 忽略了规则（某些公开符号缺少注释）

**解释**：这个实验验证的是**"软约束是否生效"**。即使 CLAUDE.md 写了规则，
AI 也可能因为各种原因忽略。只有通过 comment-check 验证后，才能确认 AI 确实遵守了规则。

---

### 实验 C：反馈闭环（需要 AI + 系统组件协作）

**问题**：当 AI 的代码有违规时，comment-check 能否驱动 AI 修正？

**步骤**：
1. 先给 AI 看 `experiment/feedback_loop/bad.go`（已故意写错）
2. 要求 AI **不用读 CLADE.md 或自行审查**，直接运行：
   ```bash
   go run ./cmd/comment-check/ -- specs/m00-coding-agent/experiment/feedback_loop/bad.go
   ```
3. AI 看到违规输出后，要求它**只根据违规报告**修正代码
4. 修正后再次运行 comment-check
5. 记录：AI 能否仅凭工具反馈就修正所有违规？

**解释**：这个实验验证**"闭环是否工作"**。一个真正的 Agent 系统，
Agent（AI）不需要预先知道怎么写出完美代码。它只需要：
1. 写一版
2. 跑工具看反馈
3. 根据反馈修正
4. 重复直到通过

这恰恰是 `MultiStepAgent` 的核心循环——Agent 不靠"天生完美"，
而是靠"写→观测→修正"的迭代闭环。

---

## 结果记录

| 实验 | 验证方式 | AI 参与 | 结果 |
|------|---------|--------|------|
| A: 系统独立性 | verify.sh | ❌ 不需要 | ✅ 23/23 通过 |
| B: AI 合规性 | comment-check | ✅ 需要 | ⬜ 待执行 |
| C: 反馈闭环 | comment-check | ✅ 需要 | ⬜ 待执行 |

---

## 核心洞察

| 问题 | 答案 |
|------|------|
| AI 写代码时一定会写 godoc 吗？ | 不一定。CLAUDE.md 是软约束，AI 可能忽略。 |
| 能发现 AI 没写 godoc 吗？ | 能。comment-check 是硬约束，独立运行。 |
| 发现后 AI 能修好吗？ | 取决于反馈闭环是否有效（实验 C 验证）。 |
| 这就是 Agent 吗？ | 不完全。真正的 Agent 是代码驱动的循环（Module 5 实现）。目前的编码 Agent 是"AI + 工具 + 反馈"的混合系统。 |
