# Module 1 — 任务分解

> 完成条件：每个任务完成后运行 `go test ./pkg/types/...` 通过

## Task 1: 定义核心接口

- 文件：`pkg/types/agent_type.go`
- 内容：`AgentType` interface
- 验证：编译通过

## Task 2: 实现 AgentText

- 文件：`pkg/types/agent_text.go`
- 实现 `AgentType` 接口
- 构造函数 `NewAgentText(src any) (*AgentText, error)` 支持 `string` / `[]byte` / `io.Reader`
- 验证：测试 3 种输入源

## Task 3: 实现 AgentImage

- 文件：`pkg/types/agent_image.go`
- 实现 `AgentType` 接口
- 构造函数支持 `image.Image` / 文件路径 / `[]byte`（编码后图片）
- `ToRaw()` 延迟加载：如果只有路径，调用时才解码
- 验证：测试每种输入源，验证 `ToRaw()` 并发安全

## Task 4: 实现 AgentAudio

- 文件：`pkg/types/agent_audio.go`
- 实现 `AgentType` 接口
- 构造函数支持 WAV 文件路径 / `[]byte`（WAV 编码）/ `[]float32` + 采样率
- `ToRaw()` 延迟加载
- 验证：测试每种输入源

## Task 5: 集成验证

- 文件：`cmd/service-01-types/main.go`
- 演示：创建不同类型的 AgentType，打印 `String()` 和 `ToRaw()` 结果
- 运行：`go run cmd/service-01-types/main.go`
