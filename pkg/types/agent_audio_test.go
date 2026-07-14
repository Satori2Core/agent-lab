// Knowledge: AGENT-TYPE-AUDIO — 音频类型
// 测试 AgentAudio：统一音频类型，延迟加载，支持 WAV/原始采样。
// Reference: smolagents agent_types.py → AgentAudio 类
package types

import (
	"os"
	"path/filepath"
	"testing"
)

// testSamples 生成一个简单的 440Hz 正弦波测试音频。
func testSamples() []float32 {
	samples := make([]float32, 1000)
	for i := range samples {
		samples[i] = float32(i) / 1000.0
	}
	return samples
}

// TestAgentAudioFromSamples 验证从 []float32 + sampleRate 构造。
func TestAgentAudioFromSamples(t *testing.T) {
	samples := testSamples()
	aa, err := NewAgentAudio(samples, 16000)
	if err != nil {
		t.Fatalf("NewAgentAudio(samples) unexpected error: %v", err)
	}

	// String() 应返回非空标识
	if aa.String() == "" {
		t.Error("String() should not be empty")
	}

	raw, err := aa.ToRaw()
	if err != nil {
		t.Fatalf("ToRaw() unexpected error: %v", err)
	}
	result, ok := raw.(AudioData)
	if !ok {
		t.Fatalf("ToRaw() returned %T, want AudioData", raw)
	}
	if result.SampleRate != 16000 {
		t.Errorf("SampleRate = %d, want 16000", result.SampleRate)
	}
	if len(result.Samples) != 1000 {
		t.Errorf("Samples length = %d, want 1000", len(result.Samples))
	}
}

// TestAgentAudioFromPath 验证从 WAV 文件路径构造，延迟加载生效。
func TestAgentAudioFromPath(t *testing.T) {
	// 创建测试 WAV 文件
	tmpDir := t.TempDir()
	wavPath := filepath.Join(tmpDir, "test.wav")
	samples := testSamples()

	if err := writeWAV(wavPath, samples, 16000); err != nil {
		t.Fatalf("创建测试 WAV 失败: %v", err)
	}

	aa, err := NewAgentAudio(wavPath, 0) // sampleRate 从 WAV 读取
	if err != nil {
		t.Fatalf("NewAgentAudio(path) unexpected error: %v", err)
	}

	// String() 应返回文件路径
	if aa.String() != wavPath {
		t.Errorf("String() = %q, want %q", aa.String(), wavPath)
	}

	// ToRaw() 应触发延迟加载
	raw, err := aa.ToRaw()
	if err != nil {
		t.Fatalf("ToRaw() from path unexpected error: %v", err)
	}
	result := raw.(AudioData)
	if result.SampleRate != 16000 {
		t.Errorf("SampleRate = %d, want 16000", result.SampleRate)
	}
}

// TestAgentAudioFromBytes 验证从 WAV 字节流构造。
func TestAgentAudioFromBytes(t *testing.T) {
	// 创建测试 WAV 字节
	tmpDir := t.TempDir()
	wavPath := filepath.Join(tmpDir, "tmp.wav")
	samples := testSamples()
	writeWAV(wavPath, samples, 16000)
	wavBytes, _ := os.ReadFile(wavPath)

	aa, err := NewAgentAudio(wavBytes, 0)
	if err != nil {
		t.Fatalf("NewAgentAudio([]byte) unexpected error: %v", err)
	}

	raw, _ := aa.ToRaw()
	result := raw.(AudioData)
	if result.SampleRate != 16000 {
		t.Errorf("SampleRate = %d, want 16000", result.SampleRate)
	}
}

// TestAgentAudioFromUnsupported 验证不支持的输入类型返回错误。
func TestAgentAudioFromUnsupported(t *testing.T) {
	_, err := NewAgentAudio(42, 0)
	if err == nil {
		t.Error("NewAgentAudio(int) should return error")
	}
}

// TestAgentAudioSave 验证 Save 方法（保存为 WAV）。
func TestAgentAudioSave(t *testing.T) {
	samples := testSamples()
	aa, _ := NewAgentAudio(samples, 16000)

	tmpFile := filepath.Join(t.TempDir(), "saved.wav")
	if err := aa.Save(tmpFile); err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Save() should create the file")
	}
}

// compileTimeCheck 确保 AgentAudio 满足 AgentType 接口。
var _ AgentType = (*AgentAudio)(nil)
