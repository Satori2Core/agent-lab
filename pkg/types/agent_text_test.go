// Knowledge: AGENT-TYPE-TEXT — 文本类型
// 测试 AgentText：对 string 的封装，实现 AgentType 接口。
// Reference: smolagents agent_types.py → AgentText 类
package types

import (
	"strings"
	"testing"
)

// TestAgentTextFromString 验证从 string 构造。
func TestAgentTextFromString(t *testing.T) {
	at, err := NewAgentText("hello world")
	if err != nil {
		t.Fatalf("NewAgentText(string) unexpected error: %v", err)
	}

	if at.String() != "hello world" {
		t.Errorf("String() = %q, want %q", at.String(), "hello world")
	}

	raw, err := at.ToRaw()
	if err != nil {
		t.Fatalf("ToRaw() unexpected error: %v", err)
	}
	if raw != "hello world" {
		t.Errorf("ToRaw() = %q, want %q", raw, "hello world")
	}
}

// TestAgentTextFromBytes 验证从 []byte 构造。
func TestAgentTextFromBytes(t *testing.T) {
	at, err := NewAgentText([]byte("from bytes"))
	if err != nil {
		t.Fatalf("NewAgentText([]byte) unexpected error: %v", err)
	}
	if at.String() != "from bytes" {
		t.Errorf("String() = %q, want %q", at.String(), "from bytes")
	}
}

// TestAgentTextFromReader 验证从 io.Reader 构造。
func TestAgentTextFromReader(t *testing.T) {
	r := strings.NewReader("from reader")
	at, err := NewAgentText(r)
	if err != nil {
		t.Fatalf("NewAgentText(io.Reader) unexpected error: %v", err)
	}
	if at.String() != "from reader" {
		t.Errorf("String() = %q, want %q", at.String(), "from reader")
	}
}

// TestAgentTextFromUnsupported 验证不支持的类型返回错误。
func TestAgentTextFromUnsupported(t *testing.T) {
	_, err := NewAgentText(42) // int 不支持
	if err == nil {
		t.Error("NewAgentText(int) should return error")
	}
}

// TestAgentTextEmpty 验证空字符串的零值行为。
func TestAgentTextEmpty(t *testing.T) {
	at, err := NewAgentText("")
	if err != nil {
		t.Fatalf("NewAgentText(empty) unexpected error: %v", err)
	}
	if at.String() != "" {
		t.Errorf("Empty AgentText String() = %q, want empty", at.String())
	}
}

// compileTimeCheck 确保 AgentText 满足 AgentType 接口。
var _ AgentType = (*AgentText)(nil)
