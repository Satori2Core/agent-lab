// Knowledge: AGENT-TYPE-ABSTRACTION — Agent 输出类型抽象
// Agent 的输出不只有文本，需要统一的多模态类型接口。
// Reference: smolagents agent_types.py → AgentType 类
//
// 在 smolagents 中，AgentType 是一个抽象基类，定义了三个职责：
//   1. 表现得像底层原始类型（文本像 string，图片像 PIL.Image）
//   2. 可被转成字符串（String() 用于塞进 LLM 的上下文窗口）
//   3. 可返回原始数据（ToRaw() 用于下游程序消费）
//
// Go 版本用 interface 替代 Python 的抽象基类 + 多重继承。
// 这比 Python 版本更简洁 —— Go 的隐式接口满足意味着任何实现了
// String() 和 ToRaw() 的类型自动成为 AgentType。
package types

import "fmt"

// AgentType 是所有 Agent 输出类型的统一接口。
//
// 任何实现了 AgentType 的类型都可以被 Agent 的 Tool 返回，
// 并在 LLM 上下文（通过 String()）和程序消费（通过 ToRaw()）之间无缝切换。
//
// 实现者必须满足两个行为契约：
//   1. String() — 返回适合 LLM 上下文的文本表示
//      例如：AgentText 返回文本内容，AgentImage 返回文件路径/URI
//   2. ToRaw() — 返回底层原始数据，供程序进一步处理
//      例如：AgentText 返回 string，AgentImage 返回 image.Image
//
// 对标 smolagents：agent_types.py → AgentType 类
//   Python: class AgentType: to_raw(), to_string()
//   Go:     interface AgentType { String() + ToRaw() }
type AgentType interface {
	fmt.Stringer

	// ToRaw 返回该类型的底层原始数据。
	//
	// 返回：
	//   - 原始数据的具体类型取决于实现类型（见各类型的文档）
	//   - 如果延迟加载失败（如文件不存在），返回 nil + error
	//
	// 线程安全：
	//   ToRaw() 可能被多个 goroutine 并发调用。
	//   实现者应使用 sync.Once 或其他机制保证延迟加载的线程安全。
	ToRaw() (any, error)
}
