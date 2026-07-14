// Knowledge: AGENT-TOOL-REGISTRY — 工具注册中心
// ToolRegistry 管理 Agent 可用的所有工具：注册、查找、导出为 OpenAI 格式。
// Reference: smolagents tools.py → ToolCollection
package tools

import (
	"encoding/json"
	"fmt"
	"sync"
)

// ToolRegistry 是工具的注册中心。
//
// Agent 通过 ToolRegistry 知道自己有哪些工具可用。
// 对标 smolagents 的 ToolCollection —— 两者的职责完全相同。
//
// 线程安全：所有方法都用 sync.RWMutex 保护，支持并发读写。
//
// 对标 smolagents：
//   Python: class ToolCollection: __init__(tools), add_tool(), to_openai()
//   Go:     type ToolRegistry struct { Register(), Get(), List(), ToOpenAI() }
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]*Tool
}

// NewToolRegistry 创建一个空的 ToolRegistry。
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]*Tool),
	}
}

// Register 注册一个工具。
//
// 如果已存在同名工具，返回错误（防止意外覆盖）。
//
// 参数：
//   - t: 要注册的工具
//
// 可能的错误：
//   - 工具名已存在
func (r *ToolRegistry) Register(t *Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[t.Name]; exists {
		return fmt.Errorf("ToolRegistry: 工具 %q 已注册，不允许重复注册", t.Name)
	}
	r.tools[t.Name] = t
	return nil
}

// Get 按名称查找工具。
//
// 返回：
//   - 找到的工具指针和 true
//   - nil 和 false（未找到）
func (r *ToolRegistry) Get(name string) (*Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

// List 返回所有已注册工具的列表。
//
// 返回顺序不保证稳定。
func (r *ToolRegistry) List() []*Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]*Tool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, t)
	}
	return tools
}

// ToOpenAI 将所有工具转为 OpenAI function calling 格式。
//
// 返回格式为 []map[string]any，可直接放入 Chat Completions API 的 tools 字段。
// 每个工具的 Parameters（JSON Schema）保持不变。
//
// 对标 smolagents ToolCollection.to_openai()。
func (r *ToolRegistry) ToOpenAI() []map[string]any {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]map[string]any, 0, len(r.tools))
	for _, t := range r.tools {
		// 解析 Parameters 为 map（避免嵌套的 json.RawMessage）
		var paramsMap map[string]any
		json.Unmarshal(t.Parameters, &paramsMap)

		result = append(result, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        t.Name,
				"description": t.Description,
				"parameters":  paramsMap,
			},
		})
	}
	return result
}
