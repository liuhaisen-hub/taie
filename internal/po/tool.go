package po

// ToolConfig 是 .conf/tools.json 数组中的一个工具配置（所有工具集中在这一个文件里）。
type ToolConfig struct {
	ToolName string       `json:"toolName"`
	Desc     string       `json:"desc"`
	Info     string       `json:"info"`
	Args     []CommonJson `json:"args"` // 配置参数
	Enable   bool         `json:"enable"`
}

type CommonJson struct {
	Label string `json:"label"` // 字段名，作为表单的 key
	Name  string `json:"name"`  // 展示名，表单 label 显示用
	Value string `json:"value"`
}
