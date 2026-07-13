# 知识映射表：Agent 知识点 ↔ 代码实现

> 编码 Agent 每写一个文件，必须在文件头部用 `// Knowledge:` 引用本表中的知识点。
> 这确保每一行代码都有对应的 Agent 理论支撑，不是盲写。

---

## Week 0: 编码 Agent 自身（元工具层）

| 知识点 ID | Agent 概念 | 代码体现 | 说明 |
|-----------|-----------|---------|------|
| `AGENT-SYS-PROMPT` | System Prompt 设计 | `CLAUDE.md` | 用结构化指令约束 Agent 行为，定义角色/规则/禁止行为 |
| `AGENT-TOOL` | Tool 设计 | `cmd/comment-check/main.go` | 给 Agent 一个可调用的验证工具：输入=Go文件路径，输出=违规列表 |
| `AGENT-RAG` | 知识注入 (RAG) | `KNOWLEDGE_MAP.md` | 将领域知识（Agent 理论）注入到编码上下文中 |
| `AGENT-OBSERVE` | Observation 反馈 | `.claude/settings.json` Hook | Agent 保存文件后自动触发验证，获得即时反馈 |
| `AGENT-PLAN` | Planning 规划 | `specs/*/SPEC.md` + `TASKS.md` | 大任务拆成可验证的小步，每步有明确的完成标准 |
| `AGENT-MEMORY` | Memory 记忆 | `PRACTICE.md` 进度表 | 追踪已完成/未完成的 Week 和 Task，持久化状态 |

---

## Week 1: Agent 多模态类型系统

| 知识点 ID | Agent 概念 | 代码体现 | 说明 |
|-----------|-----------|---------|------|
| `AGENT-TYPE-ABSTRACTION` | Agent 输出类型抽象 | `AgentType` interface | Agent 的输出不只有文本，需要统一的多模态类型接口 |
| `AGENT-TYPE-TEXT` | 文本类型 | `AgentText` struct | 对 string 的封装，实现 AgentType 接口 |
| `AGENT-TYPE-IMAGE` | 图像类型 | `AgentImage` struct | 支持多种输入源（路径/字节/Image对象），延迟加载，线程安全 |
| `AGENT-TYPE-AUDIO` | 音频类型 | `AgentAudio` struct | 支持 WAV/原始采样，延迟加载 |
| `AGENT-TYPE-LAZY` | 延迟加载策略 | `sync.Once` 在 `ToRaw()` 中 | Agent 不总是需要原始数据，按需加载优化性能 |
| `AGENT-TYPE-CONSTRUCTOR` | 多源构造模式 | `NewAgentText/Image/Audio()` | 统一多种输入格式，Agent 需要处理来自不同来源的数据 |

---

## Week 2: LLM 模型抽象层

| 知识点 ID | Agent 概念 | 代码体现 | 说明 |
|-----------|-----------|---------|------|
| `AGENT-MODEL-INTERFACE` | 模型接口抽象 | `Model` interface | Agent 不绑定特定 LLM，通过接口实现供应商无关 |
| `AGENT-MODEL-STREAM` | 流式输出 | `<-chan Delta` | Agent 需要边生成边展示，不是等全部完成 |
| `AGENT-MODEL-MESSAGE` | 消息结构 | `ChatMessage` struct | system/user/assistant/tool 四种角色的消息建模 |
| `AGENT-MODEL-OPENAI` | OpenAI 适配 | `OpenAIModel` struct | 具体模型的 HTTP 调用实现 |

---

## Week 3: 工具系统

| 知识点 ID | Agent 概念 | 代码体现 | 说明 |
|-----------|-----------|---------|------|
| `AGENT-TOOL-INTERFACE` | Tool 接口 | `Tool` struct | 工具的名称、描述、参数 schema、执行函数 |
| `AGENT-TOOL-SCHEMA` | JSON Schema 生成 | 从 Go struct 反射生成 | LLM 的 function calling 需要 JSON Schema 描述参数 |
| `AGENT-TOOL-REGISTRY` | 工具注册 | `ToolRegistry` | Agent 需要知道有哪些工具可用 |
| `AGENT-TOOL-VALIDATE` | 参数校验 | `validate()` 方法 | 调用前校验参数，防止 LLM 传错格式 |

---

## Week 4: 记忆系统

| 知识点 ID | Agent 概念 | 代码体现 | 说明 |
|-----------|-----------|---------|------|
| `AGENT-MEMORY-STEP` | 步骤追踪 | `ActionStep/PlanningStep` | 每一步思考/行动/观察都被记录下来 |
| `AGENT-MEMORY-CTX` | 上下文窗口管理 | `Messages()` 方法 | 从记忆中提取 LLM 输入，处理截断 |
| `AGENT-MEMORY-REPLAY` | 记忆回放 | 多步对话模拟 | 支持从记忆重建完整的 Agent 执行轨迹 |

---

## Week 5: ReAct 核心循环

| 知识点 ID | Agent 概念 | 代码体现 | 说明 |
|-----------|-----------|---------|------|
| `AGENT-LOOP-REACT` | ReAct 循环 | `MultiStepAgent.Run()` | Think → Act → Observe → Repeat |
| `AGENT-LOOP-EXIT` | 退出条件 | `maxSteps` + `FinalAnswer` | Agent 必须知道什么时候停 |
| `AGENT-LOOP-ERROR` | 错误处理 | 工具调用失败的降级策略 | Agent 遇到错误是继续还是退出 |

---

## Week 6: Agent 服务化

| 知识点 ID | Agent 概念 | 代码体现 | 说明 |
|-----------|-----------|---------|------|
| `AGENT-SERVICE-HTTP` | HTTP API | `agent-server/main.go` | Agent 暴露为网络服务 |
| `AGENT-SERVICE-SSE` | 流式响应 | SSE handler | Agent 边思考边输出给客户端 |

---

## Week 7: 代码执行 Agent

| 知识点 ID | Agent 概念 | 代码体现 | 说明 |
|-----------|-----------|---------|------|
| `AGENT-CODE-SANDBOX` | 安全沙箱 | Docker/进程隔离执行 | LLM 生成的代码不可信，必须隔离执行 |
| `AGENT-CODE-TIMEOUT` | 执行超时 | context.WithTimeout | 防止死循环耗尽资源 |
