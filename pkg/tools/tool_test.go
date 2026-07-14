// Knowledge: AGENT-TOOL-INTERFACE — Tool 接口
// 测试 Tool 创建、JSON Schema 生成、参数校验。
// Reference: smolagents tools.py → Tool 类
package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Satori2Core/agent-lab/pkg/types"
)

// WeatherInput 天气查询工具的参数。
type WeatherInput struct {
	City string `json:"city" desc:"城市名称，如 Beijing"`
	Days int    `json:"days" desc:"查询未来几天，1-7"`
}

// TestNewToolSchema 验证 JSON Schema 自动生成。
func TestNewToolSchema(t *testing.T) {
	fn := func(ctx context.Context, input WeatherInput) (*types.AgentText, error) {
		return types.NewAgentText(input.City + " 晴天")
	}

	tool, err := NewTool("get_weather", "查询指定城市的天气", fn)
	if err != nil {
		t.Fatalf("NewTool() error: %v", err)
	}

	if tool.Name != "get_weather" {
		t.Errorf("Name = %q, want %q", tool.Name, "get_weather")
	}
	if tool.Description != "查询指定城市的天气" {
		t.Errorf("Description = %q", tool.Description)
	}

	// 验证 JSON Schema 包含关键字段
	var schema map[string]any
	if err := json.Unmarshal(tool.Parameters, &schema); err != nil {
		t.Fatalf("Parameters 不是合法的 JSON: %v", err)
	}

	if schema["type"] != "object" {
		t.Errorf("schema type = %v, want object", schema["type"])
	}

	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("schema 缺少 properties")
	}

	cityProp, ok := props["city"].(map[string]any)
	if !ok {
		t.Fatal("schema properties 缺少 city")
	}
	if cityProp["type"] != "string" {
		t.Errorf("city type = %v, want string", cityProp["type"])
	}
	if cityProp["description"] != "城市名称，如 Beijing" {
		t.Errorf("city description = %v", cityProp["description"])
	}

	// 验证 required 字段
	required, ok := schema["required"].([]any)
	if !ok || len(required) == 0 {
		t.Error("schema 缺少 required 字段")
	}
}

// TestToolExecute 验证工具执行流程。
func TestToolExecute(t *testing.T) {
	fn := func(ctx context.Context, input WeatherInput) (*types.AgentText, error) {
		result := input.City + " 晴天，未来" + string(rune('0'+input.Days)) + "天"
		return types.NewAgentText(result)
	}

	tool, _ := NewTool("get_weather", "查询天气", fn)

	// 模拟 LLM 传入的 JSON 参数
	params := json.RawMessage(`{"city":"Beijing","days":3}`)
	result, err := tool.Fn(context.Background(), params)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	text, ok := result.(*types.AgentText)
	if !ok {
		t.Fatalf("result type = %T, want *AgentText", result)
	}
	if text.String() == "" {
		t.Error("result should not be empty")
	}
}

// TestNewToolJSONSchemaFormat 验证生成的 JSON Schema 是标准格式。
func TestNewToolJSONSchemaFormat(t *testing.T) {
	type TestInput struct {
		Name    string  `json:"name" desc:"用户姓名"`
		Age     int     `json:"age" desc:"年龄"`
		Height  float64 `json:"height,omitempty" desc:"身高(可选)"`
		IsAdmin bool    `json:"is_admin" desc:"是否管理员"`
	}

	fn := func(ctx context.Context, input TestInput) (*types.AgentText, error) {
		return types.NewAgentText("ok")
	}

	tool, _ := NewTool("test", "测试工具", fn)
	var schema map[string]any
	json.Unmarshal(tool.Parameters, &schema)

	props := schema["properties"].(map[string]any)

	// string
	if props["name"].(map[string]any)["type"] != "string" {
		t.Error("name type should be string")
	}
	// int
	if props["age"].(map[string]any)["type"] != "integer" {
		t.Error("age type should be integer")
	}
	// float64
	if props["height"].(map[string]any)["type"] != "number" {
		t.Error("height type should be number")
	}
	// bool
	if props["is_admin"].(map[string]any)["type"] != "boolean" {
		t.Error("is_admin type should be boolean")
	}

	// omitempty → 不在 required 中
	required := schema["required"].([]any)
	for _, r := range required {
		if r.(string) == "height" {
			t.Error("height 有 omitempty，不应在 required 中")
		}
	}
}

// compileTimeCheck — ToolFunc 签名检查
var _ ToolFunc = func(context.Context, json.RawMessage) (types.AgentType, error) {
	return nil, nil
}

// TestValidateValid 验证合法参数通过校验。
func TestValidateValid(t *testing.T) {
	tool := makeTestTool("test")
	params := json.RawMessage(`{"query":"hello"}`)
	if err := tool.Validate(params); err != nil {
		t.Errorf("合法参数不应报错: %v", err)
	}
}

// TestValidateMissingRequired 验证缺少必填字段被检测。
func TestValidateMissingRequired(t *testing.T) {
	tool := makeTestTool("test")
	params := json.RawMessage(`{}`)
	if err := tool.Validate(params); err == nil {
		t.Error("缺少必填字段应报错")
	}
}

// TestValidateWrongType 验证类型错误被检测。
func TestValidateWrongType(t *testing.T) {
	tool := makeTestTool("test")
	params := json.RawMessage(`{"query": 123}`)
	if err := tool.Validate(params); err == nil {
		t.Error("字段类型错误应报错")
	}
}

// TestValidateInvalidJSON 验证非法 JSON 被检测。
func TestValidateInvalidJSON(t *testing.T) {
	tool := makeTestTool("test")
	params := json.RawMessage(`not json`)
	if err := tool.Validate(params); err == nil {
		t.Error("非法 JSON 应报错")
	}
}
