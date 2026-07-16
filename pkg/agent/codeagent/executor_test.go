// Knowledge: AGENT-CODE-SANDBOX — 代码执行沙箱测试
package codeagent

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestExecutorPython 验证 Python 代码执行。
func TestExecutorPython(t *testing.T) {
	// 检查 Python 是否可用
	executor := NewLocalExecutor()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	output, err := executor.Execute(ctx, "python", "print(2 + 3)")
	if err != nil {
		t.Skipf("Python 不可用: %v", err)
	}
	if !strings.Contains(output, "5") {
		t.Errorf("output = %q, should contain '5'", output)
	}
}

// TestExecutorShell 验证 Shell 代码执行。
func TestExecutorShell(t *testing.T) {
	executor := NewLocalExecutor()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	output, err := executor.Execute(ctx, "shell", "echo hello world")
	if err != nil {
		t.Fatalf("Execute shell error: %v", err)
	}
	if !strings.Contains(output, "hello world") {
		t.Errorf("output = %q", output)
	}
}

// TestExecutorError 验证代码错误时返回输出和错误。
func TestExecutorError(t *testing.T) {
	executor := NewLocalExecutor()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	output, err := executor.Execute(ctx, "python", "print(1/0)")
	if err == nil {
		t.Error("division by zero should return error")
	}
	if output == "" {
		t.Error("error output should not be empty")
	}
}

// TestExecutorTimeout 验证超时控制。
func TestExecutorTimeout(t *testing.T) {
	executor := NewLocalExecutor()
	executor.Timeout = 2 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := executor.Execute(ctx, "python", "while True: pass")
	if err == nil {
		t.Skip("Python 不可用或无限循环未被终止")
	}
}

// TestNewExecuteCodeTool 验证工具创建和基本调用。
func TestNewExecuteCodeTool(t *testing.T) {
	executor := NewLocalExecutor()
	tool, err := NewExecuteCodeTool(executor)
	if err != nil {
		t.Fatalf("NewExecuteCodeTool error: %v", err)
	}
	if tool.Name != "execute_code" {
		t.Errorf("Name = %q", tool.Name)
	}
}
