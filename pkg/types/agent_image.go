// Knowledge: AGENT-TYPE-IMAGE — 图像类型
// AgentImage 统一图片处理，支持多源输入和延迟加载。
// Reference: smolagents agent_types.py → AgentImage 类
//
// 对标关系：
//   Python: class AgentImage(AgentType, PIL.Image.Image)  # 多重继承
//   Go:     type AgentImage struct { ... }                # 组合 + 接口
//
// 关键差异：
//   Python 版通过多重继承让 AgentImage "就是"一个 PIL.Image，
//   可以直接 .save()、.resize() 等。
//   Go 版通过组合持有 image.Image 和 path 两种状态，
//   暴露 Save() 方法而非嵌入整个 image.Image 接口。
package types

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// AgentImage 表示 Agent 输出的图像类型。
//
// AgentImage 统一了三种输入源：
//   1. image.Image — 内存中的图像对象（直接持有）
//   2. 文件路径 — 延迟加载，ToRaw() 时才解码
//   3. []byte — 编码后的图片数据（PNG/JPEG），ToRaw() 时才解码
//
// 设计原则：
//   - 延迟加载：如果只有路径，不立即解码，节省内存
//   - 线程安全：sync.Once 保证延迟加载只执行一次
//   - 零拷贝：如果已持有原始数据，直接返回
//
// 对标 smolagents：
//   Python: AgentImage 构造函数支持 PIL.Image / path / bytes / torch.Tensor
//   Go:     NewAgentImage 支持 image.Image / string path / []byte / *AgentImage
type AgentImage struct {
	raw    image.Image // 解码后的原始图像（延迟填充）
	path   string      // 文件路径（延迟加载时使用）
	data   []byte      // 编码后的字节（延迟解码时使用）
	format string      // 图片格式：png / jpeg

	once sync.Once // 保证延迟加载只执行一次
}

// NewAgentImage 从多种输入源构造 AgentImage。
//
// 支持的输入类型：
//   - image.Image: 直接持有，String() 返回 data URI 标识
//   - string: 文件路径，String() 返回路径，ToRaw() 时延迟加载
//   - []byte: 编码后的图片字节（PNG/JPEG），ToRaw() 时延迟解码
//   - *AgentImage: 直接返回（避免重复包装）
//
// 参数：
//   - src: 输入源
//
// 返回：
//   - 构造好的 *AgentImage
//   - 不支持的类型返回 nil + error
func NewAgentImage(src any) (*AgentImage, error) {
	switch v := src.(type) {
	case image.Image:
		return &AgentImage{raw: v}, nil
	case string:
		return &AgentImage{path: v, format: formatFromExt(v)}, nil
	case []byte:
		return &AgentImage{data: v, format: detectFormat(v)}, nil
	case *AgentImage:
		return v, nil
	default:
		return nil, fmt.Errorf("AgentImage: 不支持的类型 %T，支持 image.Image / string(path) / []byte / *AgentImage", src)
	}
}

// String 返回图像的文本表示。
//
// 对于 LLM 上下文，返回文件路径（如果有）或内存标识。
// 对标 smolagents：to_string() 返回 self._path。
func (ai *AgentImage) String() string {
	if ai.path != "" {
		return ai.path
	}
	// 内存中的图像没有路径，返回一个标识
	return fmt.Sprintf("<AgentImage:%s>", ai.formatOrDefault())
}

// ToRaw 返回底层原始图像数据。
//
// 如果图像已持有（构造时传入了 image.Image），直接返回。
// 如果只有路径或字节流，触发延迟加载并解码。
// 多次调用 ToRaw() 只会在第一次时执行解码（sync.Once 保护）。
//
// 返回：
//   - image.Image 类型的原始图像
//   - 解码失败时返回 nil + error
func (ai *AgentImage) ToRaw() (any, error) {
	var loadErr error
	ai.once.Do(func() {
		// 如果已持有原始数据，无需加载
		if ai.raw != nil {
			return
		}

		// 从路径加载
		if ai.path != "" {
			ai.raw, loadErr = loadImage(ai.path)
			return
		}

		// 从字节解码
		if len(ai.data) > 0 {
			ai.raw, loadErr = decodeImage(ai.data)
			return
		}
	})

	if loadErr != nil {
		return nil, fmt.Errorf("AgentImage: 加载失败: %w", loadErr)
	}
	if ai.raw == nil {
		return nil, fmt.Errorf("AgentImage: 无可用图像数据")
	}
	return ai.raw, nil
}

// Save 将图像保存到指定文件路径。
//
// 参数：
//   - dst: 目标文件路径，扩展名决定格式（.png → PNG, .jpg/.jpeg → JPEG）
//
// 可能的错误：
//   - 图像数据不可用（未加载且无数据源）
//   - 编码或写入文件失败
func (ai *AgentImage) Save(dst string) error {
	raw, err := ai.ToRaw()
	if err != nil {
		return fmt.Errorf("AgentImage.Save: %w", err)
	}
	img, ok := raw.(image.Image)
	if !ok {
		return fmt.Errorf("AgentImage.Save: 内部数据不是 image.Image")
	}

	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("AgentImage.Save: 创建文件失败: %w", err)
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(dst))
	switch ext {
	case ".png":
		return png.Encode(f, img)
	case ".jpg", ".jpeg":
		return jpeg.Encode(f, img, nil)
	default:
		// 默认使用 PNG
		return png.Encode(f, img)
	}
}

// formatOrDefault 返回格式名，未设置时返回 "png"。
func (ai *AgentImage) formatOrDefault() string {
	if ai.format != "" {
		return ai.format
	}
	return "png"
}

// formatFromExt 根据文件扩展名推断格式。
func formatFromExt(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return "png"
	case ".jpg", ".jpeg":
		return "jpeg"
	default:
		return ""
	}
}

// detectFormat 从字节流魔术字检测图片格式。
func detectFormat(data []byte) string {
	if len(data) < 4 {
		return ""
	}
	// PNG: 89 50 4E 47
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "png"
	}
	// JPEG: FF D8 FF
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "jpeg"
	}
	return ""
}

// loadImage 从文件路径加载并解码图片。
func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()
	return decodeImageReader(f)
}

// decodeImage 从字节流解码图片。
func decodeImage(data []byte) (image.Image, error) {
	return decodeImageReader(strings.NewReader(string(data)))
}

// decodeImageReader 从 io.Reader 解码图片（image.Decode 自动检测格式）。
func decodeImageReader(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("解码图片失败: %w", err)
	}
	return img, nil
}
