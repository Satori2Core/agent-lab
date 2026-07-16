# 自查清单：不看代码，能答出来吗？

> 用法：每道题先自己回答（口述或写下来），再对照代码验证。
> 答不出的标记 ❓，攒够了一起问我。

---

## M01 — Agent 多模态类型系统

1. AgentText 和普通 string 有什么区别？为什么要多包一层？
2. AgentImage 为什么用 `sync.Once`？去掉会怎样？
3. `AgentType` 接口只有 `String()` 和 `ToRaw()` 两个方法——为什么够用？

**写出代码（不看文件）**：
```go
// 写出 AgentType 接口的定义
```

---

## M02 — LLM 模型抽象层

1. 为什么要有 `Model` 接口，而不是直接写 `OpenAIModel`？
2. `Generate()` 和 `GenerateStream()` 的区别是什么？各自适合什么场景？
3. 如果让你支持 Anthropic（非 OpenAI 兼容），需要改哪些地方？

**写出代码（不看文件）**：
```go
// 写出 Model 接口
// 写出 ChatMessage 结构体的字段
```

---

## M03 — Tool 系统

1. `NewTool` 用到了 Go 泛型——`TInput` 和 `TOutput` 分别是干什么的？
2. JSON Schema 是怎么从 Go struct 生成出来的？用了什么 Go 特性？
3. `Tool.Validate()` 在 Agent 循环中起什么作用？去掉会怎样？

**写出代码（不看文件）**：
```go
// 定义一个带参数的工具（你自己设计一个 struct + tag）
// 写出 NewTool 的泛型签名
```

---

## M04 — 记忆系统

1. 为什么 `Messages()` 和 `MessagesWithLimit(3)` 是两个方法而不是一个？
2. `ActionStep` 有 Thought、Action、Observation 三个字段——分别对应 ReAct 循环的哪个阶段？
3. 上下文截断策略是"保留 system prompt + 最近的 N 条"——如果改成"保留最旧的"会怎样？

**写出代码（不看文件）**：
```go
// ActionStep 有哪些字段？写出来
```

---

## M05 — ReAct 核心循环

1. **ReAct 循环的四个阶段是什么？**（提示：T→A→O→R）
2. 什么时候 Agent 会停止循环？有哪些退出条件？
3. 模型返回 `ToolCalls` 后，Agent 做了什么？用伪代码描述。

**写出伪代码（不看文件）**：
```
func Run(task):
    messages = [system_prompt, user_task]
    for step in 1..maxSteps:
        response = model.Generate(messages, tools)
        if 有 ToolCalls:
            ???
        else:
            ???
```

---

## M06 — Agent HTTP 服务化

1. SSE（Server-Sent Events）和普通 JSON 响应的区别是什么？
2. `StepObserver` 在 SSE 中起什么作用？
3. 如果 10 个用户同时发请求，Agent 会互相干扰吗？为什么？

---

## M07 — 代码执行 Agent

1. `execute_code` 是怎么把"写代码"变成 Agent 的一个 Tool 的？
2. 为什么代码要在临时目录执行，而不是当前目录？
3. 超时控制是怎么实现的？

---

## 综合题

1. 从用户发消息到 Agent 返回答案，消息在哪些模块之间流转？画出数据流。
2. 如果 Agent 调了一个不存在的工具，会发生什么？
3. 如果要加一个新工具（比如查快递），需要改哪些代码？在哪加？

**画出架构图（口头描述即可）**：
```
用户消息 → ??? → ??? → ??? → 答案
```

---

## 自检结果

| 模块 | 答出 | 答不出 | 笔记 |
|------|------|--------|------|
| M01 类型系统 | /3 | | |
| M02 模型抽象 | /3 | | |
| M03 Tool 系统 | /3 | | |
| M04 记忆系统 | /3 | | |
| M05 ReAct 循环 | /3 | | |
| M06 HTTP 服务化 | /3 | | |
| M07 代码执行 | /3 | | |
| 综合 | /3 | | |

---

## 怎么用

1. 先口头答一遍（不需要完美，说个大概就行）
2. 答不出的标 ❓
3. 拿不准的标 ⚠️
4. 答完一个模块，可以对照 `specs/mXX-*/SPEC.md` 验证
5. 攒够 ❓ 来找我，我们针对性地讲

**目标不是答满分，是知道自己哪里没懂。**
