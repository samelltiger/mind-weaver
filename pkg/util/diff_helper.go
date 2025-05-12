package util

import (
	"strings"
)

// ProcessDiffText 处理 diff 文本
// 对第一个代码块保留删除操作（只去掉减号），对后续代码块执行真正的增删操作
func ProcessDiffText(diffText string) string {
	lines := strings.Split(diffText, "\n")
	output := make([]string, 0)
	firstBlockProcessed := false

	for _, line := range lines {
		if strings.HasPrefix(line, "-") && !firstBlockProcessed {
			// 第一个代码块中的删除行，只去掉减号（保留缩进）
			output = append(output, line[1:])
		} else if strings.HasPrefix(line, "+") && !firstBlockProcessed {
			// 第一个代码块中的增加行，忽略
			continue
		} else if strings.HasPrefix(line, "?") && !firstBlockProcessed {
			// 第一个代码块中的差异标记，忽略
			continue
		} else {
			// 标记第一个代码块已处理完毕
			if !firstBlockProcessed && !strings.HasPrefix(line, "-") &&
				!strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "?") {
				firstBlockProcessed = true
			}

			// 处理后续代码块
			if firstBlockProcessed {
				if strings.HasPrefix(line, "-") {
					// 后续代码块的删除行，跳过
					continue
				} else if strings.HasPrefix(line, "+") {
					// 后续代码块的增加行，保留（去掉加号）
					output = append(output, line[1:])
				} else if !strings.HasPrefix(line, "?") {
					// 普通行保留
					output = append(output, line)
				}
			}
		}
	}

	return strings.Join(output, "\n")
}
