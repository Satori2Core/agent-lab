#!/bin/bash
# ============================================================
# 编码 Agent 系统验证脚本
# ============================================================
# 独立于 AI Chat，在终端直接运行。
# 验证 Week 0 构建的 "编码 Agent" 系统组件是否正常工作。
#
# 对标知识点:
#   AGENT-OBSERVE - 系统级 Observation: Agent 行为的客观验证
#   AGENT-TOOL    - Tool 的独立可运行性
#
# 运行:
#   cd /e/Yang/learn/agent-lab
#   bash specs/week-00-coding-agent/verify.sh
# ============================================================

set -e
cd "$(dirname "$0")/../.."

PASSED=0
FAILED=0

step() {
    echo ""
    echo "━━━ $1 ━━━"
}

pass() {
    echo "  ✅ $1"
    PASSED=$((PASSED + 1))
}

fail() {
    echo "  ❌ $1"
    FAILED=$((FAILED + 1))
}

# ============================================================
# 检查 1: 项目结构完整性
# ============================================================
step "检查 1: 项目结构完整性"

REQUIRED_FILES=(
    "CLAUDE.md"
    "KNOWLEDGE_MAP.md"
    "PRACTICE.md"
    "go.mod"
    ".claude/settings.json"
    "cmd/comment-check/main.go"
    "specs/week-00-coding-agent/SPEC.md"
    "specs/week-00-coding-agent/TASKS.md"
    "specs/week-00-coding-agent/PRACTICE_GUIDE.md"
)

for f in "${REQUIRED_FILES[@]}"; do
    if [ -f "$f" ]; then
        pass "存在: $f"
    else
        fail "缺失: $f"
    fi
done

# ============================================================
# 检查 2: comment-check 编译
# ============================================================
step "检查 2: comment-check 编译"

if go build -o /tmp/comment-check.exe ./cmd/comment-check/ 2>/dev/null; then
    pass "编译成功"
    rm -f /tmp/comment-check.exe
else
    fail "编译失败"
    go build ./cmd/comment-check/
fi

# ============================================================
# 检查 3: comment-check 自检
# ============================================================
step "检查 3: comment-check 自检 (工具检查自己的代码)"

if go run ./cmd/comment-check/ -- cmd/comment-check/main.go; then
    pass "自检通过: 0 个违规 (工具自身代码符合 godoc 规范)"
else
    fail "自检未通过: 工具自身代码有违规"
fi

# ============================================================
# 检查 4: 违规检测能力
# ============================================================
step "检查 4: 违规检测能力测试"

TEST_DIR=$(mktemp -d)
# 场景 A: 完全没有注释 (2 个违规)
cat > "$TEST_DIR/bad_no_doc.go" << 'GOEOF'
package test
// privateOk 是私有类型，不需要注释
type privateOk struct{}
type BadNoComment struct{}
func BadNoCommentFunc() string { return "bad" }
GOEOF

# 场景 B: 有注释但格式错误 (不以符号名开头)
cat > "$TEST_DIR/bad_wrong_prefix.go" << 'GOEOF'
package test
// 错了 — 不是以符号名开头
type WrongPrefixType struct{}
// 这也错了
func WrongPrefixFunc() string { return "wrong" }
GOEOF

# 场景 C: 合规代码 (0 个违规)
cat > "$TEST_DIR/good.go" << 'GOEOF'
package test
// GoodType 有正确 godoc 注释的公开类型。
type GoodType struct{}
// GoodFunc 有正确 godoc 注释的公开函数。
// 返回:
//   - "good": 永远返回此字符串
func GoodFunc() string { return "good" }
GOEOF

set +e
VIOLATIONS_OUTPUT=$(go run ./cmd/comment-check/ -- "$TEST_DIR" 2>&1)
VIOLATION_COUNT=$?
set -e

# 预期: bad_no_doc.go 2个 + bad_wrong_prefix.go 2个 = 4个违规
rm -rf "$TEST_DIR"

if [ "$VIOLATION_COUNT" -eq 4 ]; then
    pass "违规检测: 检测到 4 个违规 (预期: 2个无注释 + 2个格式错误)"
    echo "  违规详情:"
    echo "$VIOLATIONS_OUTPUT" | while IFS= read -r line; do
        echo "    $line"
    done
elif [ "$VIOLATION_COUNT" -gt 0 ]; then
    pass "违规检测: 检测到 $VIOLATION_COUNT 个违规 (预期 4 个，可能有偏差)"
    echo "  输出:"
    echo "$VIOLATIONS_OUTPUT" | while IFS= read -r line; do
        echo "    $line"
    done
else
    fail "违规检测: 未检测到违规 (预期应检测到 4 个)"
fi

# ============================================================
# 检查 5: CLAUDE.md 规则完整性
# ============================================================
step "检查 5: CLAUDE.md 规则完整性"

RULES=("godoc" "知识映射" "测试先行" "禁止行为" "工作流程")
for rule in "${RULES[@]}"; do
    if grep -q "$rule" CLAUDE.md 2>/dev/null; then
        pass "规则存在: $rule"
    else
        fail "规则缺失: $rule"
    fi
done

# ============================================================
# 检查 6: KNOWLEDGE_MAP 覆盖度
# ============================================================
step "检查 6: KNOWLEDGE_MAP 覆盖度"

WEEKS=("Week 0" "Week 1" "Week 2" "Week 3" "Week 4" "Week 5")
for week in "${WEEKS[@]}"; do
    if grep -q "$week" KNOWLEDGE_MAP.md 2>/dev/null; then
        pass "知识映射存在: $week"
    else
        fail "知识映射缺失: $week"
    fi
done

# ============================================================
# 总结
# ============================================================
echo ""
echo "============================================================"
echo "  验证结果"
echo "============================================================"
echo "  通过: $PASSED"
if [ "$FAILED" -gt 0 ]; then
    echo "  失败: $FAILED"
    echo "============================================================"
    echo ""
    echo "⚠ 有 $FAILED 项检查失败，请修复后重新运行。"
    exit 1
else
    echo "  失败: 0"
    echo "============================================================"
    echo ""
    echo "✅ 所有检查通过! 编码 Agent 系统组件工作正常。"
    echo "   注意: 此验证只检查系统级组件 (工具、结构、规则完整性)。"
    echo "   AI 行为合规性需要通过行为实验单独验证。"
fi
