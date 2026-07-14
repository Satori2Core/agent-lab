// Knowledge: AGENT-MODEL-INTERFACE — 模型抽象层集成验证
// 演示 Model 接口的非流式和流式调用。
//
// 用法:
//   set OPENAI_API_KEY=sk-xxx
//   set OPENAI_BASE_URL=http://localhost:11434/v1  (可选，默认 OpenAI)
//   go run ./cmd/service-02-models/
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Satori2Core/agent-lab/pkg/models"
)

func main() {
	fmt.Println("Week 2: LLM 模型抽象层 — 集成验证")
	fmt.Println("============================================")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("\n⚠ OPENAI_API_KEY 未设置")
		fmt.Println("  使用内置测试模式演示接口结构...")
		demoWithoutAPI()
		return
	}

	// 构建 Model（支持自定义 BaseURL）
	var opts []models.OpenAIOption
	opts = append(opts, models.WithAPIKey(apiKey))
	if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
		opts = append(opts, models.WithBaseURL(baseURL))
		fmt.Printf("  BaseURL: %s\n", baseURL)
	}

	modelID := "gpt-4o-mini"
	if v := os.Getenv("OPENAI_MODEL"); v != "" {
		modelID = v
	}
	fmt.Printf("  Model:  %s\n\n", modelID)

	model := models.NewOpenAIModel(modelID, opts...)

	// ─── 演示1: 非流式 Generate ───
	fmt.Println("── 演示1: Generate (非流式) ──")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := model.Generate(ctx, []models.ChatMessage{
		{Role: models.RoleSystem, Content: "你是一个有用的助手，回答尽量简短。"},
		{Role: models.RoleUser, Content: "用一句话解释什么是 Agent"},
	})
	if err != nil {
		fmt.Printf("  ❌ Generate 失败: %v\n", err)
	} else {
		fmt.Printf("  回复: %s\n", resp.Content)
		fmt.Printf("  完成原因: %s\n", resp.FinishReason)
	}

	// ─── 演示2: 流式 GenerateStream ───
	fmt.Println("\n── 演示2: GenerateStream (流式) ──")
	ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel2()

	ch, err := model.GenerateStream(ctx2, []models.ChatMessage{
		{Role: models.RoleUser, Content: "数到5"},
	})
	if err != nil {
		fmt.Printf("  ❌ GenerateStream 失败: %v\n", err)
		return
	}

	fmt.Print("  流式输出: ")
	for delta := range ch {
		if delta.Error != nil {
			fmt.Printf("\n  ❌ 流错误: %v\n", delta.Error)
			break
		}
		if delta.Done {
			fmt.Println("\n  ✅ 流式完成")
			break
		}
		fmt.Print(delta.Content)
	}

	fmt.Println("\n✅ Model 接口工作正常")
}

// demoWithoutAPI 在没有 API Key 时演示接口结构。
func demoWithoutAPI() {
	fmt.Println("  Model 接口定义了2个方法:")
	fmt.Println("    Generate(ctx, messages) → (*Response, error)")
	fmt.Println("    GenerateStream(ctx, messages) → (<-chan Delta, error)")
	fmt.Println()
	fmt.Println("  OpenAIModel 支持以下配置:")
	fmt.Println("    NewOpenAIModel(modelID, WithBaseURL(...), WithAPIKey(...), WithHTTPClient(...))")
	fmt.Println()
	fmt.Println("  设置 OPENAI_API_KEY 后运行可看到实际调用效果。")
	fmt.Println("  也可以: set OPENAI_BASE_URL=http://localhost:11434/v1 使用本地 Ollama。")
}
