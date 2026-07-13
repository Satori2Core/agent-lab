# ============================================================
# 编码 Agent 系统验证脚本
# ============================================================
# 此脚本独立于 AI Chat，在终端直接运行。
# 用于验证 Week 0 构建的 "编码 Agent" 系统组件是否正常工作。
#
# 对标知识点：
#   AGENT-OBSERVE — 系统级 Observation：Agent 行为的客观验证
#   AGENT-TOOL    — Tool 的独立可运行性
#
# 运行方式：
#   cd E:\Yang\learn\agent-lab
#   powershell -ExecutionPolicy Bypass -File specs\week-00-coding-agent\verify.ps1
# ============================================================

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent (Split-Path -Parent (Split-Path -Parent $PSScriptRoot))
Set-Location $ProjectRoot

$Passed = 0
$Failed = 0

function Write-Step {
    param([string]$Message)
    Write-Host "`n━━━ $Message ━━━" -ForegroundColor Cyan
}

function Write-Pass {
    param([string]$Message)
    Write-Host "  ✅ $Message" -ForegroundColor Green
    $script:Passed++
}

function Write-Fail {
    param([string]$Message)
    Write-Host "  ❌ $Message" -ForegroundColor Red
    $script:Failed++
}

# ============================================================
# 检查 1: 项目结构完整性
# ============================================================
Write-Step "检查 1: 项目结构完整性"

$RequiredFiles = @(
    "CLAUDE.md",
    "KNOWLEDGE_MAP.md",
    "PRACTICE.md",
    "go.mod",
    ".claude\settings.json",
    "cmd\comment-check\main.go",
    "specs\week-00-coding-agent\SPEC.md",
    "specs\week-00-coding-agent\TASKS.md",
    "specs\week-00-coding-agent\PRACTICE_GUIDE.md"
)

foreach ($file in $RequiredFiles) {
    $fullPath = Join-Path $ProjectRoot $file
    if (Test-Path $fullPath) {
        Write-Pass "存在: $file"
    } else {
        Write-Fail "缺失: $file"
    }
}

# ============================================================
# 检查 2: comment-check 工具编译
# ============================================================
Write-Step "检查 2: comment-check 编译"

$compileResult = go build -o "$env:TEMP\comment-check.exe" ./cmd/comment-check/ 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Pass "编译成功"
} else {
    Write-Fail "编译失败"
    Write-Host "  $compileResult" -ForegroundColor Red
}

# ============================================================
# 检查 3: comment-check 自检
# ============================================================
Write-Step "检查 3: comment-check 自检（工具检查自己的代码）"

$selfCheck = go run ./cmd/comment-check/ -- cmd/comment-check/main.go 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Pass "自检通过: comment-check 自己的代码符合规范（0 个违规）"
} else {
    Write-Fail "自检未通过: comment-check 自己的代码有违规"
    Write-Host $selfCheck
}

# ============================================================
# 检查 4: 用违规代码测试检测能力
# ============================================================
Write-Step "检查 4: 违规检测能力测试"

# 创建一个临时的违规 Go 文件（无注释的公开符号）
$testDir = "$env:TEMP\agent-lab-test"
New-Item -ItemType Directory -Force -Path $testDir | Out-Null

@'
// 这个文件故意没有 godoc 注释，用于测试 comment-check 的检测能力
package test

// 私有类型 — 不应该报违规
type privateType struct{}

// 公开类型 — 没有 godoc 注释，应该被检测到
type BadType struct{}

// 公开函数 — 没有 godoc 注释，应该被检测到
func BadFunc() string {
    return "bad"
}
'@ | Set-Content -Path "$testDir\bad.go" -Encoding UTF8

# 另一个文件：有合规注释的代码
@'
// GoodType 是一个有正确 godoc 注释的公开类型。
//
// 它演示了合规的注释格式：以类型名开头。
type GoodType struct{}

// GoodFunc 是一个有正确 godoc 注释的公开函数。
//
// 返回：
//   - "good": 永远返回这个字符串
func GoodFunc() string {
    return "good"
}
'@ | Set-Content -Path "$testDir\good.go" -Encoding UTF8

$violationCheck = go run ./cmd/comment-check/ -- $testDir 2>&1
$exitCode = $LASTEXITCODE

# 清理
Remove-Item -Recurse -Force $testDir

if ($exitCode -gt 0) {
    Write-Pass "违规检测: 检测到 $exitCode 个违规（预期 = 2: BadType + BadFunc）"
    Write-Host "  输出:"
    foreach ($line in ($violationCheck -split "`n" | Select-Object -First 5)) {
        Write-Host "    $line" -ForegroundColor DarkGray
    }
} else {
    Write-Fail "违规检测: 未检测到违规（预期应检测到 2 个）"
}

# ============================================================
# 检查 5: CLAUDE.md 规则完整性
# ============================================================
Write-Step "检查 5: CLAUDE.md 规则完整性"

$claudeContent = Get-Content "CLAUDE.md" -Raw
$requiredRules = @("godoc", "知识映射", "测试先行", "禁止行为", "工作流程")
foreach ($rule in $requiredRules) {
    if ($claudeContent -match $rule) {
        Write-Pass "规则存在: $rule"
    } else {
        Write-Fail "规则缺失: $rule"
    }
}

# ============================================================
# 检查 6: KNOWLEDGE_MAP 覆盖度
# ============================================================
Write-Step "检查 6: KNOWLEDGE_MAP 覆盖度"

$knowledgeContent = Get-Content "KNOWLEDGE_MAP.md" -Raw
$requiredWeeks = @("Week 0", "Week 1", "Week 2", "Week 3", "Week 4", "Week 5")
foreach ($week in $requiredWeeks) {
    if ($knowledgeContent -match [regex]::Escape($week)) {
        Write-Pass "知识映射存在: $week"
    } else {
        Write-Fail "知识映射缺失: $week"
    }
}

# ============================================================
# 总结
# ============================================================
Write-Host "`n" -NoNewline
Write-Host "============================================================" -ForegroundColor DarkGray
Write-Host "  验证结果" -ForegroundColor White
Write-Host "============================================================" -ForegroundColor DarkGray
Write-Host "  通过: $Passed" -ForegroundColor Green
if ($Failed -gt 0) {
    Write-Host "  失败: $Failed" -ForegroundColor Red
} else {
    Write-Host "  失败: 0" -ForegroundColor Green
}
Write-Host "============================================================" -ForegroundColor DarkGray

if ($Failed -gt 0) {
    Write-Host "`n⚠ 有 $Failed 项检查失败，请修复后重新运行。" -ForegroundColor Yellow
    exit 1
} else {
    Write-Host "`n✅ 所有检查通过！编码 Agent 系统组件工作正常。" -ForegroundColor Green
    Write-Host "   注意：此验证只检查系统级组件（工具、结构、规则完整性）。" -ForegroundColor DarkGray
    Write-Host "   AI 行为合规性需要通过行为实验单独验证。" -ForegroundColor DarkGray
    exit 0
}
