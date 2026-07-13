# 工程验收手册：编码 Agent 系统

> 验证 Week 0 构建的编码 Agent 系统组件是否正常运作。
>
> **核心原则**：此验收不依赖 AI Chat — 你可以在终端独立运行所有检查。

---

## 验收总览

| 编号 | 检查项 | 自动化 | 验证方式 |
|------|--------|--------|---------|
| V1 | 项目结构完整性 | ✅ 脚本 | `verify.sh` 检查 1 |
| V2 | comment-check 编译 | ✅ 脚本 | `verify.sh` 检查 2 |
| V3 | comment-check 自检 | ✅ 脚本 | `verify.sh` 检查 3 |
| V4 | 违规检测能力 | ✅ 脚本 | `verify.sh` 检查 4 |
| V5 | CLAUDE.md 规则完整性 | ✅ 脚本 | `verify.sh` 检查 5 |
| V6 | KNOWLEDGE_MAP 覆盖度 | ✅ 脚本 | `verify.sh` 检查 6 |
| V7 | AI 行为合规性 | ❌ 手动 | 给 AI 编码任务，运行 comment-check 验证 |
| V8 | 反馈闭环 | ❌ 手动 | 给 AI 违规代码，看能否仅凭工具反馈修正 |

V1-V6：一键验证（`verify.sh`），完全自动化。
V7-V8：需要 AI Chat 参与的定性实验。

---

## V1-V6：一键系统验证

### 操作

```bash
# 在 agent-lab 项目根目录下
cd agent-lab
bash specs/week-00-coding-agent/verify.sh
```

### 预期结果

```
━━━ 检查 1: 项目结构完整性 ━━━
  ✅ 存在: CLAUDE.md
  ✅ 存在: KNOWLEDGE_MAP.md
  ✅ 存在: PRACTICE.md
  ✅ 存在: go.mod
  ✅ 存在: .claude/settings.json
  ✅ 存在: cmd/comment-check/main.go
  ✅ 存在: specs/week-00-coding-agent/SPEC.md
  ✅ 存在: specs/week-00-coding-agent/TASKS.md
  ✅ 存在: specs/week-00-coding-agent/PRACTICE_GUIDE.md

━━━ 检查 2: comment-check 编译 ━━━
  ✅ 编译成功

━━━ 检查 3: comment-check 自检 ━━━
  ✅ 自检通过: 0 个违规

━━━ 检查 4: 违规检测能力测试 ━━━
  ✅ 违规检测: 检测到 4 个违规

━━━ 检查 5: CLAUDE.md 规则完整性 ━━━
  ✅ 规则存在: godoc
  ✅ 规则存在: 知识映射
  ✅ 规则存在: 测试先行
  ✅ 规则存在: 禁止行为
  ✅ 规则存在: 工作流程

━━━ 检查 6: KNOWLEDGE_MAP 覆盖度 ━━━
  ✅ 知识映射存在: Week 0-5

============================================================
  验证结果: 全部通过，0 失败
============================================================
```

### 每项检查测什么

| 检查 | 验证目标 | 失败意味着什么 |
|------|---------|---------------|
| V1 | 9 个关键文件都存在 | 项目初始化不完整，可能丢文件 |
| V2 | `comment-check` 能编译 | Go 代码有语法错误 |
| V3 | `comment-check` 自己的代码有 godoc 注释 | 工具没有遵守自己定的规则 |
| V4 | 能检测到"无注释"和"格式错误"两种违规类型 | 工具检测逻辑有 bug |
| V5 | `CLAUDE.md` 包含 5 项核心规则 | 编码 Agent 的 System Prompt 不完整 |
| V6 | `KNOWLEDGE_MAP.md` 覆盖了 Week 0-5 | 知识映射有盲区 |

### 如果某项失败

```bash
# 查看具体哪个文件缺失
ls CLAUDE.md KNOWLEDGE_MAP.md PRACTICE.md go.mod .claude/settings.json

# 查看 comment-check 编译错误
cd cmd/comment-check && go build

# 单独测试 comment-check 对一个文件的检查
go run ./cmd/comment-check/ -- path/to/file.go
```

---

## V7：AI 行为合规性（手动）

### 操作步骤

1. 确认在 **agent-lab 项目目录**下（CLAUDE.md 生效）
2. 对 AI 发送以下任务（复制整段）：

```
在 specs/week-00-coding-agent/experiment/with_agent/ 目录下，
新建文件 greeter.go，包含：

- 包名 greeter
- 公开类型 Greeter，字段 Name string
- 公开方法 Greet() string，返回 "Hello, {Name}!"
- 公开函数 NewGreeter(name string) *Greeter

// Knowledge: AGENT-SYS-PROMPT — System Prompt 实践
// 本文件由编码 Agent 生成，验证 CLAUDE.md 规则是否被遵守
```

3. AI 完成代码后，在终端运行（不要用 AI 运行）：

```bash
go run ./cmd/comment-check/ -- specs/week-00-coding-agent/experiment/with_agent/
```

### 预期结果

| 情况 | 退出码 | 解读 |
|------|--------|------|
| 0 违规 | 0 | ✅ AI 在 CLAUDE.md 约束下产出了合规代码 |
| 1-2 违规 | 1-2 | ⚠️ AI 部分遵守规则（格式小问题） |
| 3+ 违规 | 3+ | ❌ AI 基本忽略了 CLAUDE.md 的注释规则 |

### 结果记录

```
实验时间: _______
退出码: _______
违规数量: _______
违规详情（如有）: _______
结论: _______
```

---

## V8：反馈闭环（手动）

### 操作步骤

1. 告诉 AI：
```
运行这个命令并观察输出:
go run ./cmd/comment-check/ -- specs/week-00-coding-agent/experiment/feedback_loop/bad.go

输出显示有 4 处违规。
你的任务：仅根据上面的违规报告来修改 bad.go，使 comment-check 通过。
不要读 CLAUDE.md，不要自己审查代码 — 只根据工具输出来修正。
修改后自己再跑一次 comment-check 确认通过。
```

2. 观察 AI 的行为：
   - 它是否**只根据违规报告修改**？（而不是重写整个文件）
   - 修改后第二次运行是否通过？

### 预期结果

| 观察点 | 预期 | 实际 |
|--------|------|------|
| AI 使用了 comment-check 输出（而非自行审查） | ✅ | |
| 第一次修改后即通过 | ⚠️ 可能（取决于 AI） | |
| 需要 2-3 轮修正才通过 | ✅ 正常（迭代闭环） | |
| AI 拒绝只根据工具输出修改 | ❌ | |

### 理想行为（Agent 闭环）：

```
第 1 轮: AI 看到 4 个违规 → 修改注释 → 重新检查 → 还剩 2 个
第 2 轮: AI 看到 2 个违规 → 修改注释 → 重新检查 → 0 个违规 → ✅
```

### 结果记录

```
修正轮数: _______
最终是否通过: _______
AI 是否只用工具输出: _______
观察到的闭环行为: _______
```

---

## 验收结论

完成 V1-V8 后，填写：

| 验收项 | 结果 | 备注 |
|--------|------|------|
| V1-V6 系统组件 | ☐ 全部通过 / ☐ 有失败 | |
| V7 AI 合规性 | ☐ 通过 / ☐ 部分 / ☐ 失败 | |
| V8 反馈闭环 | ☐ 有效 / ☐ 部分 / ☐ 无效 | |
| **总体评估** | | |

---

## 快捷命令参考

```bash
# 一键系统验证（V1-V6）
bash specs/week-00-coding-agent/verify.sh

# 单独检查某文件
go run ./cmd/comment-check/ -- <文件路径>

# 检查整个目录
go run ./cmd/comment-check/ -- <目录路径>

# JSON 格式输出（可集成到其他工具）
go run ./cmd/comment-check/ -json -- <路径>

# 编译安装到 GOPATH/bin
go install ./cmd/comment-check/
comment-check ./
```
