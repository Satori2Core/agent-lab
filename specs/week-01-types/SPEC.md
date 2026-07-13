# Week 1: Agent 多模态类型系统

## 要解决的问题

LLM 的输入输出本质上是文本 token。但 Agent 可能需要处理：

- 图片（截屏、图表、照片）
- 音频（语音指令、环境声音）
- 文件（PDF、代码等）

**Agent 类型系统要解决的核心问题**：如何让 Agent 的输出在不同模态间保持统一接口，同时每种类型又能被正确显示、序列化、传递？

在 smolagents 中，`AgentType` 抽象类承担了这个角色：
- 它"表现得像"底层原始类型（文本像 string，图片像 PIL.Image）
- 它"可被转成字符串"（用于塞进 LLM 的上下文窗口）
- 它"可以被 notebook 正常显示"（`_ipython_display_`）

## 我们要实现的 Go 版本

### 核心接口

```go
// AgentType 是所有 Agent 输出类型的统一接口
type AgentType interface {
    fmt.Stringer                    // String() — 用于文本上下文（LLM prompt）
    ToRaw() (any, error)           // 返回原始数据 — 用于下游程序消费
}
```

### 具体类型

| 类型 | 底层数据 | String() 返回 | ToRaw() 返回 |
|------|---------|--------------|-------------|
| `AgentText` | `string` | 文本内容 | `string` |
| `AgentImage` | `image.Image` / 路径 / 字节流 | 文件路径或 URI | `image.Image` |
| `AgentAudio` | 浮点采样 + 采样率 / 路径 | 文件路径或 URI | `[]float32` + `sampleRate` |

### 构造函数策略

每个类型支持多种输入源，构造时统一处理：

```
AgentText  ← string / []byte / io.Reader
AgentImage ← image.Image / filepath / []byte(编码后的图片)
AgentAudio ← []float32{...} / filepath / []byte(WAV)
```

所有构造函数内部做归一化：如果传入的是路径，延迟加载；如果传入的是原始数据，直接持有。

### 设计约束

1. **不引入重量级依赖** — 只用 Go 标准库（`image`、`image/png`、`image/jpeg`）
2. **零值安全** — 各类型的零值必须是可用的
3. **线程安全** — `ToRaw()` 可被多次并发调用（延迟加载只执行一次）
4. **与 Python 版对齐** — 概念上一一对应，不自行发挥

### 参考

- smolagents `agent_types.py`：`AgentType`, `AgentText`, `AgentImage`, `AgentAudio`
- 文件路径：<https://github.com/huggingface/smolagents/blob/main/src/smolagents/agent_types.py>
