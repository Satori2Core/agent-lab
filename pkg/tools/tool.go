// Knowledge: AGENT-TOOL-INTERFACE — Tool 接口
// Tool 是 Agent 的"手"——定义工具的元数据（名称、描述、参数 Schema）和执行逻辑。
// Reference: smolagents tools.py → Tool / BaseTool 类
//
// Python 用 @tool 装饰器从函数签名和 docstring 自动生成工具描述。
// Go 用泛型 + reflect 实现等价的"从类型信息生成 JSON Schema"。
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/Satori2Core/agent-lab/pkg/types"
)

// ToolFunc 是工具执行函数的通用签名。
//
// 对标 smolagents 中 Tool.forward() 的签名。
// 每个 Tool 的内部执行函数都符合此签名：
//   输入：json.RawMessage — LLM 传入的 JSON 参数
//   输出：types.AgentType — 统一的多模态返回类型（Week 1）
type ToolFunc func(ctx context.Context, input json.RawMessage) (types.AgentType, error)

// Tool 表示一个 Agent 可调用的工具。
//
// Tool 是 smolagents 中 Tool 类的 Go 等价实现。
// 它包含两部分信息：
//   1. 元数据（Name、Description、Parameters）— 发送给 LLM，帮助模型决定是否调用
//   2. 执行逻辑（Fn）— LLM 决定调用后，真正执行的代码
//
// 对标 smolagents：
//   Python: class Tool(BaseTool): name, description, inputs, output_type, forward()
//   Go:     type Tool struct { Name, Description, Parameters, Fn }
type Tool struct {
	// Name 工具名称，用于 LLM function calling 的 function.name。
	// 命名规则：小写字母+下划线，如 "get_weather", "search_web"
	Name string `json:"name"`

	// Description 工具描述，告诉 LLM 这个工具做什么、什么时候用。
	// 这是 prompt engineering 的关键部分——描述的质量直接影响 Agent 的工具选择。
	Description string `json:"description"`

	// Parameters 参数的 JSON Schema 定义。
	// 直接兼容 OpenAI function calling 的 parameters 字段。
	// 由 NewTool() 在构造时通过反射自动生成。
	Parameters json.RawMessage `json:"parameters"`

	// Fn 工具的执行逻辑。
	// 接收 LLM 传入的 JSON 参数，返回统一的 AgentType 结果。
	Fn ToolFunc `json:"-"`
}

// NewTool 创建一个类型安全的 Tool 实例。
//
// 这是 smolagents @tool 装饰器的 Go 等价实现。
// 利用 Go 1.18+ 泛型捕获 TInput 的类型信息，通过反射自动生成 JSON Schema。
//
// 类型参数：
//   - TInput: 工具参数的结构体类型（字段的 json/desc tag 用于生成 Schema）
//   - TOutput: 工具返回值类型，必须实现 types.AgentType 接口
//
// 参数：
//   - name: 工具名称（function calling 标识）
//   - description: 工具描述（告诉 LLM 何时调用）
//   - fn: 工具的实现函数，签名 func(context.Context, TInput) (TOutput, error)
//
// 示例：
//
//	type WeatherInput struct {
//	    City string `json:"city" desc:"城市名称"`
//	}
//	tool, _ := NewTool("get_weather", "查询天气", func(ctx context.Context, input WeatherInput) (*types.AgentText, error) {
//	    return types.NewAgentText("晴天")
//	})
func NewTool[TInput, TOutput any](name string, description string, fn func(context.Context, TInput) (TOutput, error)) (*Tool, error) {
	// 1. 从 TInput 生成 JSON Schema
	schema := generateJSONSchema[TInput]()

	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("Tool %q: JSON Schema 序列化失败: %w", name, err)
	}

	// 2. 构建包装函数——将 json.RawMessage 反序列化为 TInput，调用 fn，转换 TOutput
	wrapper := func(ctx context.Context, raw json.RawMessage) (types.AgentType, error) {
		var input TInput
		if err := json.Unmarshal(raw, &input); err != nil {
			return nil, fmt.Errorf("Tool %q: 参数解析失败: %w", name, err)
		}

		output, err := fn(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("Tool %q: 执行失败: %w", name, err)
		}

		return toAgentType(output)
	}

	return &Tool{
		Name:        name,
		Description: description,
		Parameters:  schemaBytes,
		Fn:          wrapper,
	}, nil
}

// ─── JSON Schema 生成 ───

// generateJSONSchema 从类型 T 的 struct 字段生成 JSON Schema。
//
// 对标 smolagents 中 Tool.inputs 属性的自动生成逻辑。
// Go 版本用 reflect 遍历 struct 字段，根据字段类型和 tag 生成 Schema。
func generateJSONSchema[T any]() map[string]any {
	t := reflect.TypeOf((*T)(nil)).Elem()

	// 非 struct 类型 → 简单类型 Schema
	if t.Kind() != reflect.Struct {
		return simpleTypeSchema(t)
	}

	properties := make(map[string]any)
	required := make([]string, 0)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 非导出字段跳过
		if !field.IsExported() {
			continue
		}

		jsonName := fieldJSONName(field)
		propSchema := fieldSchema(field)

		properties[jsonName] = propSchema

		// omitempty 标记的字段不作为必填
		if !fieldHasOmitempty(field) {
			required = append(required, jsonName)
		}
	}

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

// simpleTypeSchema 为非 struct 类型生成 Schema（如 string、int 等）。
func simpleTypeSchema(t reflect.Type) map[string]any {
	return map[string]any{"type": goTypeToJSONType(t)}
}

// fieldSchema 为单个 struct 字段生成 JSON Schema 片段。
func fieldSchema(field reflect.StructField) map[string]any {
	schema := map[string]any{
		"type": goTypeToJSONType(field.Type),
	}

	// 从 desc tag 提取字段描述
	if desc := field.Tag.Get("desc"); desc != "" {
		schema["description"] = desc
	}

	// 枚举类型（string 类型且定义了 enum tag）
	if field.Type.Kind() == reflect.String {
		if enum := field.Tag.Get("enum"); enum != "" {
			values := strings.Split(enum, ",")
			enumValues := make([]any, len(values))
			for i, v := range values {
				enumValues[i] = strings.TrimSpace(v)
			}
			schema["enum"] = enumValues
		}
	}

	return schema
}

// goTypeToJSONType 将 Go 类型映射到 JSON Schema 类型。
func goTypeToJSONType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Map, reflect.Struct:
		return "object"
	default:
		return "string"
	}
}

// fieldJSONName 提取字段的 JSON 名称（从 json tag 或字段名推导）。
func fieldJSONName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name
	}
	// 去掉 ,omitempty 等后缀
	name, _, _ := strings.Cut(tag, ",")
	if name == "" || name == "-" {
		return field.Name
	}
	return name
}

// fieldHasOmitempty 检查字段的 json tag 是否包含 omitempty。
func fieldHasOmitempty(field reflect.StructField) bool {
	tag := field.Tag.Get("json")
	return strings.Contains(tag, "omitempty")
}

// Validate 校验 LLM 传入的 JSON 参数是否符合工具的 JSON Schema。
//
// 这是 Agent 执行 Tool 前的"安检"——防止 LLM 传入格式错误的参数。
// 对标 smolagents tools.py → validate_tool_arguments()。
//
// 校验规则：
//   1. 必须是合法的 JSON
//   2. 必须是 JSON Object（不是 string/number/array）
//   3. 所有必填字段必须存在且非 null
//   4. 字段类型必须与 Schema 声明的类型匹配
//
// 参数：
//   - input: LLM 传入的 JSON 参数字符串
//
// 返回：
//   - nil 表示参数合法
//   - error 包含具体的违规信息
func (t *Tool) Validate(input json.RawMessage) error {
	// 解析 JSON Schema
	var schema map[string]any
	if err := json.Unmarshal(t.Parameters, &schema); err != nil {
		return fmt.Errorf("Tool %q: 内部错误——JSON Schema 损坏: %w", t.Name, err)
	}

	// 解析输入参数
	var params map[string]any
	if err := json.Unmarshal(input, &params); err != nil {
		return fmt.Errorf("Tool %q: 参数不是合法的 JSON: %w", t.Name, err)
	}

	// 检查必填字段
	required, _ := schema["required"].([]any)
	for _, r := range required {
		fieldName, _ := r.(string)
		val, exists := params[fieldName]
		if !exists || val == nil {
			return fmt.Errorf("Tool %q: 缺少必填字段 %q", t.Name, fieldName)
		}
	}

	// 检查字段类型
	properties, _ := schema["properties"].(map[string]any)
	for fieldName, val := range params {
		propSchema, ok := properties[fieldName].(map[string]any)
		if !ok {
			continue // 未知字段，宽松处理
		}
		expectedType, _ := propSchema["type"].(string)
		if !matchJSONType(val, expectedType) {
			return fmt.Errorf("Tool %q: 字段 %q 类型错误，期望 %s，实际 %T",
				t.Name, fieldName, expectedType, val)
		}
	}

	return nil
}

// matchJSONType 检查一个 Go 值是否匹配 JSON Schema 声明的类型。
func matchJSONType(val any, expected string) bool {
	if val == nil {
		return false
	}
	switch expected {
	case "string":
		_, ok := val.(string)
		return ok
	case "integer":
		// JSON 数字统一为 float64
		if f, ok := val.(float64); ok {
			return f == float64(int64(f))
		}
		return false
	case "number":
		_, ok := val.(float64)
		return ok
	case "boolean":
		_, ok := val.(bool)
		return ok
	case "object":
		_, ok := val.(map[string]any)
		return ok
	case "array":
		_, ok := val.([]any)
		return ok
	default:
		return true // 未知类型，宽松处理
	}
}

// ─── 类型转换 ───

// toAgentType 将 TOutput 转换为 types.AgentType。
//
// 如果 TOutput 已实现 AgentType，直接返回。
// 如果 TOutput 是 string，自动包装为 AgentText。
// 其他类型返回 UnsupportedTypeError。
func toAgentType(output any) (types.AgentType, error) {
	// 已实现 AgentType → 直接返回
	if at, ok := output.(types.AgentType); ok {
		return at, nil
	}

	// string → AgentText（常见的便利转换）
	if s, ok := output.(string); ok {
		return types.NewAgentText(s)
	}
	if s, ok := output.(*string); ok && s != nil {
		return types.NewAgentText(*s)
	}

	return nil, fmt.Errorf("Tool: 不支持将 %T 转换为 AgentType，"+
		"工具函数应返回 types.AgentType 的实现（如 *AgentText）或 string", output)
}
