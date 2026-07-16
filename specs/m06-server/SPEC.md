# Module 6: Agent HTTP 服务化

## 要解决的问题

M05 的 Agent 只能在终端里跑。要让它能被其他程序调用，需要包装成 HTTP API。

在 smolagents 中，这个能力由 `gradio_ui.py` 提供（Web UI）。
我们直接做更通用的 HTTP API + SSE 流式。

## API 设计

### POST /chat

非流式：发送消息，等待完整响应。

```json
// Request
{"message": "北京今天天气怎么样？"}

// Response
{"answer": "北京今天晴天，22-28°C", "steps": 2, "duration_ms": 6900}
```

### GET /chat/stream (SSE)

流式：通过 Server-Sent Events 推送每一步的执行状态。

```
event: step
data: {"step":1,"action":"get_weather","observation":"Beijing: 晴天","duration_ms":3500}

event: step
data: {"step":1,"action":"get_weather","observation":"Shanghai: 多云","duration_ms":10}

event: done
data: {"answer":"两个城市都适合","steps":2,"duration_ms":7000}
```

### GET /health

```
{"status":"ok","model":"deepseek-v4-pro"}
```

## 架构

```
Client → HTTP → Handler → MultiStepAgent → Model (DeepSeek)
                         ↓
                    SSE Stream ← StepObserver ← Agent Loop
```

## 设计约束

1. **零外部框架** — 只用 `net/http`，不引入 gin/echo/chi
2. **Agent 复用** — 一个 Agent 实例处理所有请求
3. **并发安全** — 每个请求创建独立的 Agent.Run() 调用
4. **配置项** — PORT / API_KEY / BASE_URL / MODEL 均通过环境变量
