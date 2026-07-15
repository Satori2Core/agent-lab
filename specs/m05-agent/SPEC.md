# Module 5: ReAct 核心循环

## 要解决的问题

前面四周构建了 Agent 的零件（类型、模型、工具、记忆），现在要把它们**组装成能自主运行的 Agent**。

核心问题是 **ReAct 循环**（Reasoning + Acting）：

```
用户任务 → Agent 思考 → 调用工具 → 观测结果 → 再思考 → ... → 回答
            ↑_______________________________________________↓
                        循环直到任务完成或超步数
```

在 smolagents 中，`MultiStepAgent.run()` + `step()` 承担了这个角色。

## 我们要实现的 Go 版本

### MultiStepAgent

```go
type MultiStepAgent struct {
    model       models.Model
    tools       *tools.ToolRegistry
    memory      *memory.AgentMemory
    maxSteps    int
    systemPrompt string
}

func (a *MultiStepAgent) Run(ctx context.Context, task string) (*RunResult, error)
func (a *MultiStepAgent) RunStream(ctx context.Context, task string) (<-chan StreamEvent, error)
```

### ReAct 循环流程

```
1. 初始化记忆：写入 SystemPrompt + 用户 Task
2. 进入循环（最多 maxSteps 步）：
   a. 从 memory.Messages() 获取 LLM 上下文
   b. 调用 model.Generate(ctx, messages)
   c. 如果模型返回 ToolCall → 执行工具 → 记录 Observation → 继续循环
   d. 如果模型返回纯文本 → 记录 FinalAnswer → 退出循环
3. 达到 maxSteps 仍未完成 → 返回 ErrMaxSteps
```

### RunResult

```go
type RunResult struct {
    Answer      string        // 最终答案
    Steps       int           // 执行步数
    Duration    time.Duration // 总耗时
    Memory      *memory.AgentMemory // 完整执行轨迹
}
```

### 系统提示生成

系统提示告诉 LLM：
- 它的角色是什么
- 有哪些工具可用
- 如何使用工具（ReAct 格式的 instructions）

### 设计约束

1. **context.Context 贯穿** — 支持超时和取消
2. **streaming 支持** — RunStream 返回 channel，每一步实时推送
3. **错误不中断** — 工具执行失败时记录到记忆，继续下一步（Agent 自我修复）
4. **依赖注入** — Model + Tools + Memory 由外部传入，不内部创建

### 参考

- smolagents agents.py: `MultiStepAgent.run()` + `step()`
