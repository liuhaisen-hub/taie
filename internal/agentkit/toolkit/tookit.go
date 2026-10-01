package toolkit

import (
	"taie/internal/po"

	"github.com/cloudwego/eino/components/tool"
)

// ToolBuilder 由 tools.args 中的 JSON 字符串构建工具实例。
type ToolBuilder func(args []po.CommonJson) (tool.InvokableTool, error)

var registry = map[string]ToolBuilder{
	"web_search": buildWebSearch,
}

func BuildTools(active map[string][]po.CommonJson) []tool.BaseTool {
	if len(active) <= 0 {
		return nil
	}
	tools := make([]tool.BaseTool, 0, len(active))
	for n, a := range active {
		build, ok := registry[n]
		if !ok {
			// 不报错
			continue
		}
		t, err := build(a)
		if err != nil {
			// 不报错
			continue
		}
		tools = append(tools, t)
	}
	return tools
}
