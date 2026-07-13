// Knowledge: AGENT-SYS-PROMPT — System Prompt 实践
// 本文件由编码 Agent 生成，验证 CLAUDE.md 规则是否被遵守
// Reference: KNOWLEDGE_MAP.md → AGENT-SYS-PROMPT（用结构化指令约束 Agent 行为）

package greeter

import "fmt"

// Greeter 是一个问候器，持有一个名字，用于生成个性化问候语。
//
// 这是对 Agent System Prompt 概念的实践：
// Greeter 就像被配置了 System Prompt 的 Agent——它的 Name 字段定义了
// 它的"角色"（向谁问候），Greet() 方法输出符合该角色的行为。
type Greeter struct {
	// Name 是要问候的人的名字。
	Name string
}

// Greet 返回对 Greeter 所持名字的问候语。
//
// 返回：
//   - 格式为 "Hello, {Name}!" 的问候字符串。
func (g *Greeter) Greet() string {
	return fmt.Sprintf("Hello, %s!", g.Name)
}

// NewGreeter 创建一个新的 Greeter 实例。
//
// 参数：
//   - name: 要问候的人的名字。
//
// 返回：
//   - 指向新创建的 Greeter 的指针。
func NewGreeter(name string) *Greeter {
	return &Greeter{Name: name}
}
