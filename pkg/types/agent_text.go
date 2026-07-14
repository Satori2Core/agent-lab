// Knowledge: AGENT-TYPE-TEXT — 文本类型
// AgentText 是对 string 的轻量封装，实现 AgentType 接口。
// Reference: smolagents agent_types.py → AgentText 类
//
// 对标关系：
//   Python: class AgentText(AgentType, str)  # 多重继承，表现得像 str
//   Go:     type AgentText struct { value string }  # 组合优于继承
//
// 在 Go 中不能同时嵌入 string 和实现 AgentType（接口冲突），
// 因此用组合方式，暴露 String() 和 ToRaw() 方法。
package types

import (
	"fmt"
	"io"
)

// AgentText 表示 Agent 输出的文本类型。
//
// AgentText 封装了一个 string 值，同时实现 AgentType 接口。
// 它是 Agent 最基础的输出类型 —— 大多数 Tool 返回的就是 AgentText。
//
// 对标 smolagents：
//   Python: class AgentText(AgentType, str)
//     可以直接当 str 用：text[0], text.upper(), f"{text}"
//   Go:
//     通过 String() 获得文本内容
//     通过 ToRaw() 获得底层 string
type AgentText struct {
	value string
}

// NewAgentText 从多种输入源构造 AgentText。
//
// 支持的输入类型：
//   - string: 直接持有
//   - []byte: 转为 string
//   - io.Reader: 读取全部内容转为 string
//
// 参数：
//   - src: 输入源，支持 string / []byte / io.Reader
//
// 返回：
//   - 构造好的 *AgentText
//   - 不支持的类型返回 nil + error
//
// 可能的错误：
//   - io.Reader 读取失败
//   - 不支持的输入类型
func NewAgentText(src any) (*AgentText, error) {
	switch v := src.(type) {
	case string:
		return &AgentText{value: v}, nil
	case []byte:
		return &AgentText{value: string(v)}, nil
	case io.Reader:
		data, err := io.ReadAll(v)
		if err != nil {
			return nil, fmt.Errorf("AgentText: 读取 io.Reader 失败: %w", err)
		}
		return &AgentText{value: string(data)}, nil
	case *AgentText:
		// 避免重复包装
		return v, nil
	default:
		return nil, fmt.Errorf("AgentText: 不支持的类型 %T，支持 string / []byte / io.Reader / *AgentText", src)
	}
}

// String 返回文本内容。
// 这是 fmt.Stringer 接口的实现，也是 AgentType.String() 的实现。
// Agent 框架会将此返回值放入 LLM 的上下文窗口。
func (t *AgentText) String() string {
	return t.value
}

// ToRaw 返回底层原始数据。
// 对于 AgentText，原始数据就是 string 本身。
//
// 返回：
//   - string 类型的原始文本
//   - 永远不会返回 error（文本类型的 ToRaw 不会失败）
func (t *AgentText) ToRaw() (any, error) {
	return t.value, nil
}
