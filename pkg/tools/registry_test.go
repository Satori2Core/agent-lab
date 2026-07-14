// Knowledge: AGENT-TOOL-REGISTRY — 工具注册
// 测试 ToolRegistry：注册、查找、重复检测、OpenAI 格式转换。
// Reference: smolagents tools.py → ToolCollection
package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Satori2Core/agent-lab/pkg/types"
)

// makeTestTool 创建一个简单测试工具。
func makeTestTool(name string) *Tool {
	type Input struct {
		Query string `json:"query" desc:"查询关键词"`
	}
	fn := func(ctx context.Context, input Input) (*types.AgentText, error) {
		return types.NewAgentText("result for " + input.Query)
	}
	t, _ := NewTool(name, "测试工具: "+name, fn)
	return t
}

// TestRegistryRegister 验证注册和查找。
func TestRegistryRegister(t *testing.T) {
	r := NewToolRegistry()

	t1 := makeTestTool("search")
	if err := r.Register(t1); err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	got, ok := r.Get("search")
	if !ok {
		t.Fatal("Get() should find registered tool")
	}
	if got.Name != "search" {
		t.Errorf("Name = %q", got.Name)
	}
}

// TestRegistryDuplicate 验证重复注册检测。
func TestRegistryDuplicate(t *testing.T) {
	r := NewToolRegistry()
	r.Register(makeTestTool("search"))
	err := r.Register(makeTestTool("search"))
	if err == nil {
		t.Error("重复注册应返回错误")
	}
}

// TestRegistryList 验证列出所有工具。
func TestRegistryList(t *testing.T) {
	r := NewToolRegistry()
	r.Register(makeTestTool("a"))
	r.Register(makeTestTool("b"))
	r.Register(makeTestTool("c"))

	tools := r.List()
	if len(tools) != 3 {
		t.Errorf("List() length = %d, want 3", len(tools))
	}
}

// TestRegistryGetMissing 验证查找不存在的工具。
func TestRegistryGetMissing(t *testing.T) {
	r := NewToolRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("Get() should return false for missing tool")
	}
}

// TestRegistryToOpenAI 验证 OpenAI function calling 格式转换。
func TestRegistryToOpenAI(t *testing.T) {
	r := NewToolRegistry()
	r.Register(makeTestTool("search"))
	r.Register(makeTestTool("calculate"))

	result := r.ToOpenAI()
	if len(result) != 2 {
		t.Fatalf("ToOpenAI() length = %d, want 2", len(result))
	}

	// 验证格式: {"type": "function", "function": {...}}
	first := result[0]
	if first["type"] != "function" {
		t.Errorf("type = %v, want function", first["type"])
	}

	fnBlock, ok := first["function"].(map[string]any)
	if !ok {
		t.Fatal("missing function block")
	}
	if _, ok := fnBlock["name"]; !ok {
		t.Error("function block missing name")
	}
	if _, ok := fnBlock["parameters"]; !ok {
		t.Error("function block missing parameters")
	}

	// 验证 parameters 是合法的 JSON Schema
	paramsJSON, _ := json.Marshal(fnBlock["parameters"])
	var schema map[string]any
	json.Unmarshal(paramsJSON, &schema)
	if schema["type"] != "object" {
		t.Error("parameters should be an object schema")
	}
}
