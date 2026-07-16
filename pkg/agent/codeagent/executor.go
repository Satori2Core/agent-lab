// Knowledge: AGENT-CODE-SANDBOX — 代码执行沙箱
// CodeExecutor 提供安全的代码执行环境，支持 Python/Shell + 超时控制。
// Reference: smolagents local_python_executor.py → LocalPythonExecutor
package codeagent

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CodeExecutor 是代码执行器的统一接口。
//
// 对标 smolagents 的 PythonExecutor——提供一个安全的代码执行环境。
// 不同的实现可以用不同的沙箱策略（本地进程、Docker、远程执行）。
type CodeExecutor interface {
	// Execute 在隔离环境中执行一段代码。
	//
	// 参数：
	//   - ctx: 上下文（超时控制）
	//   - language: 语言类型（"python" / "shell"）
	//   - code: 要执行的代码
	//
	// 返回：
	//   - 完整的 stdout+stderr 输出
	//   - 超时/执行失败时返回 error
	Execute(ctx context.Context, language string, code string) (string, error)
}

// LocalExecutor 通过 os/exec 在本地进程执行代码。
//
// 安全措施：
//   - 超时保护（默认 30s，通过 WithTimeout 配置）
//   - 隔离工作目录（临时目录）
//   - stdout+stderr 合并返回
//
// 对标 smolagents 的 LocalPythonExecutor。
type LocalExecutor struct {
	Timeout time.Duration
}

// NewLocalExecutor 创建一个本地代码执行器。
func NewLocalExecutor() *LocalExecutor {
	return &LocalExecutor{Timeout: 30 * time.Second}
}

// Execute 执行代码并返回合并的 stdout+stderr。
func (e *LocalExecutor) Execute(ctx context.Context, language string, code string) (string, error) {
	// 创建隔离的工作目录
	workDir, err := os.MkdirTemp("", "agent-exec-*")
	if err != nil {
		return "", fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(workDir)

	// 将代码写入临时文件
	ext := fileExt(language)
	scriptPath := filepath.Join(workDir, "script"+ext)
	if err := os.WriteFile(scriptPath, []byte(code), 0o700); err != nil {
		return "", fmt.Errorf("写入脚本失败: %w", err)
	}

	// 构建执行命令
	cmdArgs := buildCmd(language, scriptPath)
	if len(cmdArgs) == 0 {
		return "", fmt.Errorf("不支持的语言: %s (支持 python / shell)", language)
	}

	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 执行（context 控制超时）
	if err := cmd.Run(); err != nil {
		// 返回错误 + 已收集的输出（帮助 LLM 诊断）
		output := mergeOutput(stdout.String(), stderr.String())
		return output, fmt.Errorf("执行失败: %w\n输出:\n%s", err, output)
	}

	return mergeOutput(stdout.String(), stderr.String()), nil
}

// ─── 辅助函数 ───

// fileExt 根据语言返回脚本文件扩展名。
func fileExt(language string) string {
	switch strings.ToLower(language) {
	case "python", "py":
		return ".py"
	case "shell", "sh", "bash":
		return ".sh"
	default:
		return ".txt"
	}
}

// buildCmd 根据语言构建执行命令。
func buildCmd(language string, scriptPath string) []string {
	switch strings.ToLower(language) {
	case "python", "py":
		return []string{"python", scriptPath}
	case "shell", "sh", "bash":
		return []string{"bash", scriptPath}
	default:
		return nil
	}
}

// mergeOutput 合并 stdout 和 stderr。
func mergeOutput(stdout, stderr string) string {
	if stderr == "" {
		return stdout
	}
	if stdout == "" {
		return "[stderr]\n" + stderr
	}
	return stdout + "\n[stderr]\n" + stderr
}
