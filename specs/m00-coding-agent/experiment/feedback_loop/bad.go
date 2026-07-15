// 这个文件故意有 4 处违规，用于测试反馈闭环
// 实验 C: 运行 comment-check → AI 看到违规 → AI 修正 → 再检查
package bad

// 违规1: 公开类型，没有 godoc 注释
type Calculator struct {
	result int
}

// 违规2: 公开函数，没有 godoc 注释
func NewCalculator() *Calculator {
	return &Calculator{}
}

// 违规3: 有注释但格式错误（不以符号名开头）
// 计算两个数的和
func (c *Calculator) Add(a, b int) int {
	return a + b
}

// 违规4: 有注释但格式错误
// 获取当前结果
func (c *Calculator) Result() int {
	return c.result
}
