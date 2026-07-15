// Knowledge: AGENT-MODEL-INTERFACE — 模型接口抽象
// 测试 Model 接口 + ChatMessage 类型。
package models

import (
	"context"
	"testing"
)

// mockModel 是一个测试用的最小 Model 实现。
type mockModel struct {
	response string
}

func (m *mockModel) Generate(ctx context.Context, messages []ChatMessage, tools []map[string]any) (*Response, error) {
	return &Response{Content: m.response, FinishReason: "stop"}, nil
}

func (m *mockModel) GenerateStream(ctx context.Context, messages []ChatMessage, tools []map[string]any) (<-chan Delta, error) {
	ch := make(chan Delta, 2)
	go func() {
		defer close(ch)
		ch <- Delta{Content: m.response[:len(m.response)/2]}
		ch <- Delta{Content: m.response[len(m.response)/2:], Done: true}
	}()
	return ch, nil
}

// TestMessageRoleValid 验证 MessageRole 的有效性检查。
func TestMessageRoleValid(t *testing.T) {
	tests := []struct {
		role  MessageRole
		valid bool
	}{
		{RoleSystem, true},
		{RoleUser, true},
		{RoleAssistant, true},
		{RoleTool, true},
		{MessageRole(""), false},
		{MessageRole("invalid"), false},
	}
	for _, tt := range tests {
		if got := tt.role.IsValid(); got != tt.valid {
			t.Errorf("MessageRole(%q).IsValid() = %v, want %v", tt.role, got, tt.valid)
		}
	}
}

// TestModelInterface 验证 mock 实现满足 Model 接口。
func TestModelInterface(t *testing.T) {
	var m Model = &mockModel{response: "hello"}

	resp, err := m.Generate(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Generate() unexpected error: %v", err)
	}
	if resp.Content != "hello" {
		t.Errorf("Content = %q, want %q", resp.Content, "hello")
	}
}

// TestModelInterfaceStream 验证流式接口。
func TestModelInterfaceStream(t *testing.T) {
	m := &mockModel{response: "hello world"}

	ch, err := m.GenerateStream(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("GenerateStream() unexpected error: %v", err)
	}

	var content string
	for delta := range ch {
		if delta.Error != nil {
			t.Fatalf("stream error: %v", delta.Error)
		}
		content += delta.Content
	}
	if content != "hello world" {
		t.Errorf("streamed content = %q, want %q", content, "hello world")
	}
}

// compileTimeCheck
var _ Model = (*mockModel)(nil)
