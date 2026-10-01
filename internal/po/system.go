package po

// SystemConfig.Mode 的合法取值（前端设置页同款字符串）
const (
	ModeManual = "manual" // 手动：按权限开关，关=审批
	ModePlan   = "plan"   // 计划：只读放行，写/执行必须审批
	ModeAuto   = "auto"   // 自动：全部放行，减少询问
)

// SystemConfig 是 .conf/system.json 中的应用系统设置（语言、工作模式、通知与文件权限开关）。
type SystemConfig struct {
	Language      string              `json:"language"` // agent 回答使用的语言，如 zh-CN / en
	Mode          string              `json:"mode"`     // 工作模式：manual 手动 / plan 计划 / auto 自动（减少询问）
	FontSize      string              `json:"fontSize"` // 界面字号
	Notifications NotificationSetting `json:"notifications"`
	Permissions   PermissionSetting   `json:"permissions"`
}

// NotificationSetting 各类事件的通知开关。
type NotificationSetting struct {
	ResponseCompletions bool `json:"responseCompletions"` // 响应补全完成时通知
	ScheduledTasks      bool `json:"scheduledTasks"`      // 计划任务执行时通知
	Notifications       bool `json:"notifications"`       // 事件通知
	PermissionRequests  bool `json:"permissionRequests"`  // 需要授权时通知
	Emails              bool `json:"emails"`              // 云会话邮件
}

// PermissionSetting agent 内置工具的授权开关，key 与工具名一一对应。
type PermissionSetting struct {
	ReadFile  bool `json:"read_file"`  // read_file：读取文件内容
	WriteFile bool `json:"write_file"` // write_file：写入文件内容
	EditFile  bool `json:"edit_file"`  // edit_file：编辑文件内容
	Glob      bool `json:"glob"`       // glob：根据 glob 模式查找文件
	Grep      bool `json:"grep"`       // grep：在文件中搜索内容
	Execute   bool `json:"execute"`    // execute：执行 shell 命令
}

// Lookup 返回工具对应的授权开关；不在权限清单中的工具 ok=false（不受权限管辖）
func (p *PermissionSetting) Lookup(toolName string) (allowed, ok bool) {
	switch toolName {
	case "read_file":
		return p.ReadFile, true
	case "write_file":
		return p.WriteFile, true
	case "edit_file":
		return p.EditFile, true
	case "glob":
		return p.Glob, true
	case "grep":
		return p.Grep, true
	case "execute":
		return p.Execute, true
	}
	return false, false
}

// IsPermissionTool 工具是否在权限清单管辖范围内
func IsPermissionTool(toolName string) bool {
	var p PermissionSetting
	_, ok := p.Lookup(toolName)
	return ok
}

// IsWriteTool 写/执行类工具：plan 模式下无论开关一律人工审批
func IsWriteTool(toolName string) bool {
	switch toolName {
	case "write_file", "edit_file", "execute":
		return true
	}
	return false
}
