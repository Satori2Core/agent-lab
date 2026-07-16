# Module 7: 代码执行 Agent

## 要解决的问题

M05 的 Agent 通过 function calling 调用预定义工具。但很多任务需要**动态生成代码**——写个脚本处理数据、算个复杂数学题、自动化操作。

在 smolagents 中，`CodeAgent` 让 LLM 生成 Python 代码，在沙箱中执行，观察结果后迭代修正。

**核心循环**：生成代码 → 沙箱执行 → 观察输出 → 修正 → 再执行 → 直到正确。

## 我们要实现的 Go 版本

### CodeExecutor

```go
type CodeExecutor interface {
    Execute(ctx context.Context, language string, code string) (string, error)
}
```

- `LocalExecutor` — 通过 `os/exec` 调用 Python/Shell
- 超时保护（context.WithTimeout，默认 30s）
- 工作目录隔离
- stdout/stderr 捕获

### CodeAgent

不是继承 MultiStepAgent，而是注册一个特殊的 `execute_code` 工具。
这样 M05 的 Agent 自动获得代码执行能力——**用 Tool 系统扩展 Agent。**

LLM 可以：
1. 写代码 → 调 `execute_code(python, code)` → 看输出
2. 输出有错误 → 修正代码 → 再调 `execute_code`
3. 输出正确 → 给出最终答案

### 设计约束

1. **Tool 扩展模式** — 不用继承，用组合。CodeAgent = MultiStepAgent + execute_code tool
2. **安全沙箱** — 超时、隔离工作目录、无网络、禁用危险操作
3. **错误反馈** — stdout+stderr 全部返回给 LLM，让它自我修正
