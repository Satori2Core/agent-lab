// Knowledge: AGENT-TYPE-ABSTRACTION — Agent 输出类型抽象
// 测试 AgentType 接口是否能被正确实现。
package types

import (
	"fmt"
	"testing"
)

// mockAgentType 是一个测试用的最小实现。
type mockAgentType struct {
	text string
	raw  any
}

// String 返回 mock 的文本表示。
func (m mockAgentType) String() string {
	return m.text
}

// ToRaw 返回 mock 的原始数据。
func (m mockAgentType) ToRaw() (any, error) {
	return m.raw, nil
}

// TestAgentTypeInterface 验证 mock 实现正确满足 AgentType 接口。
// 这是一个编译期检查 —— 如果 mockAgentType 不满足 AgentType，
// 下面的赋值语句会导致编译错误。
func TestAgentTypeInterface(t *testing.T) {
	var at AgentType = mockAgentType{
		text: "hello",
		raw:  "world",
	}

	if at.String() != "hello" {
		t.Errorf("String() = %q, want %q", at.String(), "hello")
	}

	raw, err := at.ToRaw()
	if err != nil {
		t.Fatalf("ToRaw() unexpected error: %v", err)
	}
	if raw != "world" {
		t.Errorf("ToRaw() = %q, want %q", raw, "world")
	}
}

// TestAgentTypeInterfaceContract 验证接口契约：
// 1. String() 和 ToRaw() 可以返回不同的值（文本表示 ≠ 原始数据）
// 2. 这是设计意图 —— String() 给 LLM 看，ToRaw() 给程序用
func TestAgentTypeInterfaceContract(t *testing.T) {
	// 模拟 AgentImage 的场景：String() 返回路径，ToRaw() 返回原始数据
	at := mockAgentType{
		text: "/tmp/image.png", // LLM 看到路径
		raw:  []byte{0x89, 0x50, 0x4E, 0x47}, // 程序拿到 PNG 字节
	}

	if at.String() != "/tmp/image.png" {
		t.Error("String() should return the path for LLM context")
	}

	raw, _ := at.ToRaw()
	png, ok := raw.([]byte)
	if !ok {
		t.Fatal("ToRaw() should return []byte for image data")
	}
	if len(png) != 4 {
		t.Errorf("ToRaw() image bytes length = %d, want 4", len(png))
	}
}

// TestAgentTypeNilSafety 验证接口的 nil 安全性。
func TestAgentTypeNilSafety(t *testing.T) {
	// AgentType 的零值实现者 —— 空字符串是安全的默认值
	at := mockAgentType{}
	if at.String() != "" {
		t.Error("zero value String() should return empty string")
	}

	raw, err := at.ToRaw()
	if err != nil {
		t.Fatalf("zero value ToRaw() should not error: %v", err)
	}
	if raw != nil {
		t.Errorf("zero value ToRaw() = %v, want nil", raw)
	}
}

// compileTimeCheck 确保 mockAgentType 满足 AgentType 和 fmt.Stringer。
// 如果编译失败，说明接口设计有误。
var (
	_ AgentType   = mockAgentType{}
	_ fmt.Stringer = mockAgentType{}
)
