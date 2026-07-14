// Knowledge: AGENT-TYPE-AUDIO — 音频类型
// AgentAudio 统一音频处理，支持多源输入、延迟加载和最低限度的 WAV 解析。
// Reference: smolagents agent_types.py → AgentAudio 类
//
// 对标关系：
//   Python: class AgentAudio(AgentType, str)
//     底层持有 torch.Tensor，String() 返回路径
//   Go:
//     底层持有 AudioData{Samples []float32, SampleRate int}
//     String() 返回文件路径或标识
//
// 设计说明：
//   Python 版用 torch.Tensor 作为音频载体，Go 没有 torch，
//   因此用原生 []float32 表示采样数据。这是一个简化，但对 Agent
//   场景足够 —— Agent 不需要做 FFT/滤波，只需要传递音频数据。
package types

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"sync"
)

// AudioData 表示音频的原始采样数据。
//
// 对标 smolagents 中 torch.Tensor 的角色 —— 是 AgentAudio.ToRaw() 的返回类型。
type AudioData struct {
	Samples    []float32 // 归一化到 [-1.0, 1.0] 的浮点采样
	SampleRate int       // 采样率，如 16000
}

// AgentAudio 表示 Agent 输出的音频类型。
//
// AgentAudio 统一了三种输入源：
//   1. []float32 + sampleRate — 内存中的采样数据
//   2. WAV 文件路径 — 延迟加载，ToRaw() 时才解码
//   3. []byte — WAV 编码的字节流，ToRaw() 时才解码
//
// 设计原则：
//   - 延迟加载：如果只有路径/字节，不立即解码
//   - 线程安全：sync.Once 保护延迟加载
//   - 零依赖：WAV 解析完全用标准库实现
//
// 对标 smolagents：
//   Python: AgentAudio(value, samplerate=16000)
//     支持 str path / torch.Tensor / tuple(samplerate, array)
//   Go:     NewAgentAudio(src, sampleRate)
//     支持 []float32 / string path / []byte
type AgentAudio struct {
	data       AudioData // 解码后的音频数据（延迟填充）
	path       string    // 文件路径（延迟加载时使用）
	rawBytes   []byte    // WAV 字节流（延迟解码时使用）
	sampleRate int       // 用户指定的采样率（WAV 文件自动覆盖）

	once sync.Once // 保证延迟加载只执行一次
}

// NewAgentAudio 从多种输入源构造 AgentAudio。
//
// 支持的输入类型：
//   - []float32: 原始采样数据，需同时指定 sampleRate
//   - string: WAV 文件路径，sampleRate 会被 WAV 头覆盖（传 0 即可）
//   - []byte: WAV 编码的字节流，sampleRate 会被 WAV 头覆盖（传 0 即可）
//
// 参数：
//   - src: 输入源
//   - sampleRate: 采样率（当 src 是 []float32 时必需；WAV 文件自动检测）
//
// 返回：
//   - 构造好的 *AgentAudio
//   - 不支持的类型返回 nil + error
func NewAgentAudio(src any, sampleRate int) (*AgentAudio, error) {
	switch v := src.(type) {
	case []float32:
		if sampleRate <= 0 {
			return nil, fmt.Errorf("AgentAudio: []float32 输入必须指定正数的 sampleRate，收到了 %d", sampleRate)
		}
		samples := make([]float32, len(v))
		copy(samples, v)
		return &AgentAudio{
			data:       AudioData{Samples: samples, SampleRate: sampleRate},
			sampleRate: sampleRate,
		}, nil
	case string:
		return &AgentAudio{path: v, sampleRate: sampleRate}, nil
	case []byte:
		return &AgentAudio{rawBytes: v, sampleRate: sampleRate}, nil
	default:
		return nil, fmt.Errorf("AgentAudio: 不支持的类型 %T，支持 []float32 / string(path) / []byte(WAV)", src)
	}
}

// String 返回音频的文本表示。
//
// 对标 smolagents：to_string() 返回文件路径（用于 LLM 上下文）。
func (aa *AgentAudio) String() string {
	if aa.path != "" {
		return aa.path
	}
	return fmt.Sprintf("<AgentAudio:%dHz,%d samples>", aa.sampleRate, len(aa.data.Samples))
}

// ToRaw 返回底层原始音频数据。
//
// 如果采样数据已持有，直接返回。
// 如果只有路径或字节流，触发延迟加载并解码 WAV。
// 多次调用只在第一次时执行解码（sync.Once 保护）。
//
// 返回：
//   - AudioData 类型（包含 Samples []float32 和 SampleRate int）
//   - 解码失败时返回 nil + error
func (aa *AgentAudio) ToRaw() (any, error) {
	var loadErr error
	aa.once.Do(func() {
		// 已持有数据，无需加载
		if len(aa.data.Samples) > 0 {
			return
		}

		// 从路径加载 WAV
		if aa.path != "" {
			aa.data, loadErr = loadWAV(aa.path)
			return
		}

		// 从字节解码 WAV
		if len(aa.rawBytes) > 0 {
			aa.data, loadErr = parseWAV(aa.rawBytes)
			return
		}
	})

	if loadErr != nil {
		return nil, fmt.Errorf("AgentAudio: 加载失败: %w", loadErr)
	}
	if len(aa.data.Samples) == 0 {
		return nil, fmt.Errorf("AgentAudio: 无可用音频数据")
	}
	return aa.data, nil
}

// Save 将音频保存为 WAV 文件（16-bit PCM）。
//
// 参数：
//   - dst: 目标文件路径
//
// 可能的错误：
//   - 音频数据不可用
//   - 写入文件失败
func (aa *AgentAudio) Save(dst string) error {
	raw, err := aa.ToRaw()
	if err != nil {
		return fmt.Errorf("AgentAudio.Save: %w", err)
	}
	data := raw.(AudioData)
	return writeWAV(dst, data.Samples, data.SampleRate)
}

// ─── 最小 WAV 解析器（仅支持 PCM 16-bit 和 32-bit float）───

// loadWAV 从文件路径读取并解析 WAV。
func loadWAV(path string) (AudioData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AudioData{}, fmt.Errorf("读取 WAV 文件失败: %w", err)
	}
	return parseWAV(data)
}

// parseWAV 从字节流解析 WAV 格式。
//
// 仅支持：
//   - PCM 16-bit 整数（最常见）
//   - IEEE 32-bit float
//   - 单声道
func parseWAV(data []byte) (AudioData, error) {
	if len(data) < 44 {
		return AudioData{}, fmt.Errorf("WAV 文件太短 (%d bytes)", len(data))
	}

	// RIFF header
	if string(data[0:4]) != "RIFF" {
		return AudioData{}, fmt.Errorf("不是有效的 WAV 文件（缺少 RIFF 头）")
	}
	if string(data[8:12]) != "WAVE" {
		return AudioData{}, fmt.Errorf("不是有效的 WAV 文件（缺少 WAVE 标识）")
	}

	// 查找 fmt chunk（可能不在偏移 12，需要扫描）
	fmtOffset := 12
	for fmtOffset < len(data)-8 {
		chunkID := string(data[fmtOffset : fmtOffset+4])
		chunkSize := int(binary.LittleEndian.Uint32(data[fmtOffset+4 : fmtOffset+8]))
		if chunkID == "fmt " {
			break
		}
		fmtOffset += 8 + chunkSize
	}

	if fmtOffset >= len(data)-16 {
		return AudioData{}, fmt.Errorf("WAV 文件中找不到 fmt chunk")
	}

	fmtData := data[fmtOffset+8:] // skip "fmt " + chunksize
	audioFormat := binary.LittleEndian.Uint16(fmtData[0:2])
	numChannels := binary.LittleEndian.Uint16(fmtData[2:4])
	sampleRate := int(binary.LittleEndian.Uint32(fmtData[4:8]))
	bitsPerSample := binary.LittleEndian.Uint16(fmtData[14:16])

	// 查找 data chunk
	dataOffset := fmtOffset + 8 + int(binary.LittleEndian.Uint32(data[fmtOffset+4:fmtOffset+8]))
	for dataOffset < len(data)-8 {
		chunkID := string(data[dataOffset : dataOffset+4])
		chunkSize := int(binary.LittleEndian.Uint32(data[dataOffset+4 : dataOffset+8]))
		if chunkID == "data" {
			break
		}
		dataOffset += 8 + chunkSize
	}

	if dataOffset >= len(data)-8 {
		return AudioData{}, fmt.Errorf("WAV 文件中找不到 data chunk")
	}

	sampleData := data[dataOffset+8:]
	dataChunkSize := int(binary.LittleEndian.Uint32(data[dataOffset+4 : dataOffset+8]))
	if len(sampleData) > dataChunkSize {
		sampleData = sampleData[:dataChunkSize]
	}

	// 解码采样
	var samples []float32
	switch {
	case audioFormat == 1 && bitsPerSample == 16:
		// PCM 16-bit 整数 → float32 [-1.0, 1.0]
		_ = numChannels // 只处理单声道，多声道取第一通道
		numSamples := len(sampleData) / 2
		samples = make([]float32, numSamples)
		for i := 0; i < numSamples; i++ {
			val := int16(binary.LittleEndian.Uint16(sampleData[i*2 : i*2+2]))
			samples[i] = float32(val) / 32768.0
		}
	case audioFormat == 3 && bitsPerSample == 32:
		// IEEE 32-bit float
		numSamples := len(sampleData) / 4
		samples = make([]float32, numSamples)
		for i := 0; i < numSamples; i++ {
			bits := binary.LittleEndian.Uint32(sampleData[i*4 : i*4+4])
			samples[i] = math.Float32frombits(bits)
		}
	default:
		return AudioData{}, fmt.Errorf(
			"不支持的 WAV 格式: format=%d, bits=%d (仅支持 PCM 16-bit 或 IEEE 32-bit float)",
			audioFormat, bitsPerSample,
		)
	}

	return AudioData{Samples: samples, SampleRate: sampleRate}, nil
}

// writeWAV 将 float32 采样写入 WAV 文件（16-bit PCM 格式）。
func writeWAV(path string, samples []float32, sampleRate int) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()

	numSamples := len(samples)
	dataSize := numSamples * 2 // 16-bit = 2 bytes per sample
	fileSize := 36 + dataSize  // header (44) - 8

	// RIFF header
	f.Write([]byte("RIFF"))
	binary.Write(f, binary.LittleEndian, uint32(fileSize))
	f.Write([]byte("WAVE"))

	// fmt chunk (PCM, mono, 16-bit)
	f.Write([]byte("fmt "))
	binary.Write(f, binary.LittleEndian, uint32(16))   // chunk size
	binary.Write(f, binary.LittleEndian, uint16(1))    // PCM format
	binary.Write(f, binary.LittleEndian, uint16(1))    // mono
	binary.Write(f, binary.LittleEndian, uint32(sampleRate))
	binary.Write(f, binary.LittleEndian, uint32(sampleRate*2)) // byte rate
	binary.Write(f, binary.LittleEndian, uint16(2))    // block align
	binary.Write(f, binary.LittleEndian, uint16(16))   // bits per sample

	// data chunk
	f.Write([]byte("data"))
	binary.Write(f, binary.LittleEndian, uint32(dataSize))
	for _, s := range samples {
		// clamp to [-1.0, 1.0] and convert to int16
		val := s
		if val > 1.0 {
			val = 1.0
		} else if val < -1.0 {
			val = -1.0
		}
		intVal := int16(val * 32767.0)
		binary.Write(f, binary.LittleEndian, intVal)
	}

	return nil
}
