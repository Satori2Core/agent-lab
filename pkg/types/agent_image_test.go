// Knowledge: AGENT-TYPE-IMAGE — 图像类型
// 测试 AgentImage：统一图片类型，延迟加载，线程安全。
// Reference: smolagents agent_types.py → AgentImage 类
package types

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// newTestImage 创建一个简单的测试用 image.Image（2x2 红色方块）。
func newTestImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	img.Set(0, 1, color.RGBA{255, 0, 0, 255})
	img.Set(1, 0, color.RGBA{255, 0, 0, 255})
	img.Set(1, 1, color.RGBA{255, 0, 0, 255})
	return img
}

// TestAgentImageFromImage 验证从 image.Image 构造并立即获取原始数据。
func TestAgentImageFromImage(t *testing.T) {
	img := newTestImage()
	ai, err := NewAgentImage(img)
	if err != nil {
		t.Fatalf("NewAgentImage(image.Image) unexpected error: %v", err)
	}

	// String() 应返回 data URI（因为没有文件路径）
	str := ai.String()
	if str == "" {
		t.Error("String() should not be empty for in-memory image")
	}

	// ToRaw() 应立即返回原始数据（无需延迟加载）
	raw, err := ai.ToRaw()
	if err != nil {
		t.Fatalf("ToRaw() unexpected error: %v", err)
	}
	if raw == nil {
		t.Fatal("ToRaw() should return non-nil image")
	}
	if _, ok := raw.(image.Image); !ok {
		t.Fatalf("ToRaw() returned %T, want image.Image", raw)
	}
}

// TestAgentImageFromPath 验证从文件路径构造，延迟加载生效。
func TestAgentImageFromPath(t *testing.T) {
	// 创建临时 PNG 文件
	tmpDir := t.TempDir()
	pngPath := filepath.Join(tmpDir, "test.png")

	img := newTestImage()
	f, err := os.Create(pngPath)
	if err != nil {
		t.Fatalf("创建测试 PNG 失败: %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatalf("编码 PNG 失败: %v", err)
	}
	f.Close()

	ai, err := NewAgentImage(pngPath)
	if err != nil {
		t.Fatalf("NewAgentImage(path) unexpected error: %v", err)
	}

	// String() 应返回文件路径
	if ai.String() != pngPath {
		t.Errorf("String() = %q, want %q", ai.String(), pngPath)
	}

	// ToRaw() 应触发延迟加载并返回解码后的 image
	raw, err := ai.ToRaw()
	if err != nil {
		t.Fatalf("ToRaw() from path unexpected error: %v", err)
	}
	decoded, ok := raw.(image.Image)
	if !ok {
		t.Fatalf("ToRaw() returned %T, want image.Image", raw)
	}
	if decoded.Bounds().Dx() != 2 || decoded.Bounds().Dy() != 2 {
		t.Errorf("decoded image size = %dx%d, want 2x2",
			decoded.Bounds().Dx(), decoded.Bounds().Dy())
	}
}

// TestAgentImageFromBytes 验证从编码后的字节流构造。
func TestAgentImageFromBytes(t *testing.T) {
	img := newTestImage()

	// 用临时文件生成 PNG 字节
	tmpFile := filepath.Join(t.TempDir(), "tmp.png")
	f, _ := os.Create(tmpFile)
	png.Encode(f, img)
	f.Close()
	pngBytes, _ := os.ReadFile(tmpFile)

	ai, err := NewAgentImage(pngBytes)
	if err != nil {
		t.Fatalf("NewAgentImage([]byte) unexpected error: %v", err)
	}

	raw, err := ai.ToRaw()
	if err != nil {
		t.Fatalf("ToRaw() from bytes unexpected error: %v", err)
	}
	if raw == nil {
		t.Fatal("ToRaw() should return non-nil image from bytes")
	}
}

// TestAgentImageFromUnsupported 验证不支持的输入类型返回错误。
func TestAgentImageFromUnsupported(t *testing.T) {
	_, err := NewAgentImage(42)
	if err == nil {
		t.Error("NewAgentImage(int) should return error")
	}
}

// TestAgentImageFromAgentImage 验证传递已有 AgentImage 时不重复包装。
func TestAgentImageFromAgentImage(t *testing.T) {
	img := newTestImage()
	ai1, _ := NewAgentImage(img)
	ai2, err := NewAgentImage(ai1)
	if err != nil {
		t.Fatalf("NewAgentImage(*AgentImage) unexpected error: %v", err)
	}
	// 应该返回同一个实例（避免重复包装）
	if ai1 != ai2 {
		t.Error("NewAgentImage(*AgentImage) should return the same instance")
	}
}

// TestAgentImageSave 验证 Save 方法。
func TestAgentImageSave(t *testing.T) {
	img := newTestImage()
	ai, _ := NewAgentImage(img)

	tmpFile := filepath.Join(t.TempDir(), "saved.png")
	if err := ai.Save(tmpFile); err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	// 验证保存的文件可以重新读取
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Save() should create the file")
	}
}

// compileTimeCheck 确保 AgentImage 满足 AgentType 接口。
var _ AgentType = (*AgentImage)(nil)
