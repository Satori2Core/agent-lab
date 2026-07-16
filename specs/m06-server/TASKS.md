# Module 6 — 任务分解

## Task 1: HTTP Server + 非流式端点

- 文件：`cmd/agent-server/main.go`
- 内容：`POST /chat` + `GET /health` + Agent 初始化
- 验证：`curl -X POST localhost:8080/chat -d '{"message":"你好"}'`

## Task 2: SSE 流式端点

- 文件：`cmd/agent-server/main.go`
- 内容：`GET /chat/stream` — 通过 SSE 推送 StepObserver 事件
- 验证：`curl -N localhost:8080/chat/stream -d '{"message":"查天气"}'`

## Task 3: 集成验证

- 同时测试非流式和流式两个端点
- 验证并发请求不互相干扰
