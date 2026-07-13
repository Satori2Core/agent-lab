// Knowledge: Agent Tool 设计 — 给编码 Agent 一个可调用的验证工具
// 对标: AGENT-TOOL — 输入=Go文件路径, 输出=违规列表, 退出码=违规数量
//
// comment-check 是编码 Agent 的"工具"之一。
// 在 Agent 工程中，Tool 需要明确的输入（文件路径）、输出（违规报告）和错误处理。
// 这个工具的简单性体现了好的 Tool 设计原则：单一职责、可组合、机器可读的输出。
//
// 用法：
//
//	go run cmd/comment-check/main.go ./pkg/types/...
//	go run cmd/comment-check/main.go -json ./pkg/types/agent_text.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// Violation 表示一个注释违规。
type Violation struct {
	File    string `json:"file"`    // 文件路径
	Line    int    `json:"line"`    // 行号
	Symbol  string `json:"symbol"`  // 违规符号名（类型/函数/方法）
	Kind    string `json:"kind"`    // 符号种类：func, type, method, const, var
	Reason  string `json:"reason"`  // 违规原因：missing doc / wrong format
}

// checkFile 检查单个 Go 文件中的公开符号是否有符合规范的 godoc 注释。
//
// 参数：
//   - filename: Go 源文件路径
//
// 返回：
//   - 违规列表，空列表表示检查通过
//   - 解析错误时返回 nil + error
func checkFile(filename string) ([]Violation, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", filename, err)
	}

	var violations []Violation

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			// 检查类型定义、常量、变量
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.IsExported() {
						v := checkDoc(fset, filename, s.Name.Name, "type", d.Doc, s.Doc)
						if v != nil {
							violations = append(violations, *v)
						}
					}
				case *ast.ValueSpec:
					for _, name := range s.Names {
						if name.IsExported() {
							v := checkDoc(fset, filename, name.Name, "var/const", d.Doc, s.Doc)
							if v != nil {
								violations = append(violations, *v)
							}
						}
					}
				}
			}
		case *ast.FuncDecl:
			if d.Name.IsExported() {
				v := checkDoc(fset, filename, d.Name.Name, "func", d.Doc, nil)
				if v != nil {
					violations = append(violations, *v)
				}
			}
		}
	}

	return violations, nil
}

// checkDoc 检查一个公开符号是否有合格的 godoc 注释。
//
// 合格的 godoc 注释要求：
//  1. 必须有注释文档
//  2. 注释必须以符号名开头（Go 官方惯例）
//
// 参数：
//   - fset: token 文件集，用于计算行号
//   - filename: 源文件路径
//   - name: 符号名称
//   - kind: 符号种类
//   - docs: 可能的文档来源（GenDecl 级别的 doc 或 TypeSpec/ValueSpec 级别的 doc）
//
// 返回：
//   - 违规信息，nil 表示检查通过
func checkDoc(fset *token.FileSet, filename, name, kind string, docs ...*ast.CommentGroup) *Violation {
	// 收集所有层级的文档注释
	var doc *ast.CommentGroup
	for _, d := range docs {
		if d != nil {
			doc = d
		}
	}

	if doc == nil {
		return &Violation{
			File:   filename,
			Line:   0,
			Symbol: name,
			Kind:   kind,
			Reason: "缺少 godoc 注释",
		}
	}

	// 检查注释是否以符号名开头
	firstLine := strings.TrimSpace(doc.List[0].Text)
	expectedPrefix := "// " + name
	if !strings.HasPrefix(firstLine, expectedPrefix) {
		pos := fset.Position(doc.Pos())
		return &Violation{
			File:   filename,
			Line:   pos.Line,
			Symbol: name,
			Kind:   kind,
			Reason: fmt.Sprintf("godoc 注释应以 %q 开头，实际: %s", expectedPrefix, firstLine),
		}
	}

	return nil
}

// checkDir 递归检查目录下所有 .go 文件（跳过 _test.go 和 vendor）。
//
// 参数：
//   - root: 要检查的根目录路径
//
// 返回：
//   - 所有文件违规的汇总列表
//   - 首个目录访问错误
func checkDir(root string) ([]Violation, error) {
	var allViolations []Violation

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录、非 Go 文件、测试文件、vendor
		if info.IsDir() {
			if info.Name() == "vendor" || info.Name() == ".git" || info.Name() == ".codegraph" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		violations, err := checkFile(path)
		if err != nil {
			return fmt.Errorf("check %s: %w", path, err)
		}
		allViolations = append(allViolations, violations...)
		return nil
	})

	return allViolations, err
}

// printViolations 以人类可读的格式打印违规列表。
func printViolations(violations []Violation) {
	for _, v := range violations {
		if v.Line > 0 {
			fmt.Printf("%s:%d: %s %q: %s\n", v.File, v.Line, v.Kind, v.Symbol, v.Reason)
		} else {
			fmt.Printf("%s: %s %q: %s\n", v.File, v.Kind, v.Symbol, v.Reason)
		}
	}
}

// printViolationsJSON 以 JSON 格式打印违规列表。
func printViolationsJSON(violations []Violation) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(violations)
}

func main() {
	jsonFlag := flag.Bool("json", false, "以 JSON 格式输出")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		args = []string{"."}
	}

	var allViolations []Violation
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %s: %v\n", arg, err)
			os.Exit(2)
		}

		var violations []Violation
		if info.IsDir() {
			violations, err = checkDir(arg)
		} else {
			violations, err = checkFile(arg)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(2)
		}
		allViolations = append(allViolations, violations...)
	}

	if *jsonFlag {
		printViolationsJSON(allViolations)
	} else {
		printViolations(allViolations)
	}

	// Tool 设计原则：退出码 = 违规数量（0 = 通过，>0 = 有违规）
	os.Exit(len(allViolations))
}
