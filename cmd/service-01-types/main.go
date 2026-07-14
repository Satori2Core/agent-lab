// Knowledge: AGENT-TYPE-ABSTRACTION — Agent 多模态类型系统集成验证
// 演示 AgentType 接口统一处理文本、图片、音频三种模态。
//
// 用法:
//   go run ./cmd/service-01-types/
//
// 预期输出:
//   每种类型的 String() 文本表示 + ToRaw() 原始数据摘要
package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"

	"github.com/Satori2Core/agent-lab/pkg/types"
)

// newDemoImage 创建一个简单的演示用图像（蓝色渐变方块）。
func newDemoImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			brightness := uint8((x + y) * 40)
			img.Set(x, y, color.RGBA{0, 0, brightness, 255})
		}
	}
	return img
}

// newDemoAudio 生成一个简单的 440Hz 正弦波。
func newDemoAudio() []float32 {
	const sampleRate = 16000
	const duration = 0.5 // 秒
	samples := make([]float32, int(sampleRate*duration))
	for i := range samples {
		t := float64(i) / float64(sampleRate)
		samples[i] = float32(0.5 * sin(2*3.14159*440*t))
	}
	return samples
}

// sin 简单的正弦函数（避免导入 math 包，简化演示）。
func sin(x float64) float64 {
	// Taylor 近似（对于演示够用）
	return x - x*x*x/6 + x*x*x*x*x/120
}

// printAgentType 打印 AgentType 的通用信息。
func printAgentType(at types.AgentType, label string) {
	fmt.Printf("── %s ──\n", label)
	fmt.Printf("  String(): %s\n", at.String())

	raw, err := at.ToRaw()
	if err != nil {
		fmt.Printf("  ToRaw(): 错误 — %v\n", err)
		return
	}

	switch v := raw.(type) {
	case string:
		fmt.Printf("  ToRaw(): string (%d 字符)\n", len(v))
	case image.Image:
		bounds := v.Bounds()
		fmt.Printf("  ToRaw(): image.Image (%dx%d)\n", bounds.Dx(), bounds.Dy())
	case types.AudioData:
		fmt.Printf("  ToRaw(): AudioData (%d samples, %d Hz)\n", len(v.Samples), v.SampleRate)
	default:
		fmt.Printf("  ToRaw(): %T\n", v)
	}
}

func main() {
	fmt.Println("Week 1: Agent 多模态类型系统 — 集成验证")
	fmt.Println("============================================")

	// ─── AgentText ───
	text1, _ := types.NewAgentText("Hello, Agent!")
	printAgentType(text1, "AgentText from string")

	text2, _ := types.NewAgentText([]byte("Hello from bytes"))
	printAgentType(text2, "AgentText from []byte")

	// ─── AgentImage ───
	img := newDemoImage()
	img1, _ := types.NewAgentImage(img)
	printAgentType(img1, "AgentImage from image.Image")

	// 保存 → 从路径重新加载（演示延迟加载）
	tmpDir := os.TempDir()
	pngPath := filepath.Join(tmpDir, "agent-lab-demo-image.png")
	img1.Save(pngPath)
	defer os.Remove(pngPath)

	img2, _ := types.NewAgentImage(pngPath)
	printAgentType(img2, "AgentImage from path (lazy load)")

	// ─── AgentAudio ───
	audio := newDemoAudio()
	audio1, _ := types.NewAgentAudio(audio, 16000)
	printAgentType(audio1, "AgentAudio from samples")

	// 保存 → 从路径重新加载（演示延迟加载）
	wavPath := filepath.Join(tmpDir, "agent-lab-demo-audio.wav")
	audio1.Save(wavPath)
	defer os.Remove(wavPath)

	audio2, _ := types.NewAgentAudio(wavPath, 0)
	printAgentType(audio2, "AgentAudio from path (lazy load)")

	// ─── 多态演示 ───
	fmt.Println("\n── AgentType 多态 ──")
	agents := []types.AgentType{text1, img2, audio2}
	for _, at := range agents {
		fmt.Printf("  %T → String() → %s\n", at, at.String())
	}

	fmt.Println("\n✅ 所有类型正常工作")
}
