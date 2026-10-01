package services

import (
	"context"

	"taie/internal/po"
)

type ToolsRepo interface {
	// List 分页返回工具配置，按 tools.json 中的顺序返回当页数据和总条数
	List(ctx context.Context, page, size int64) ([]*po.ToolConfig, int64, error)
	// Update 只更新工具的配置参数，其余字段原样保留
	Update(ctx context.Context, toolName string, args []po.CommonJson) error
	// Enable 启用/禁用工具；启用前要求已配置参数
	Enable(ctx context.Context, toolName string, enable bool) error
}

// ToolList 是 List 的分页查询结果。
type ToolList struct {
	Items []*po.ToolConfig `json:"items"`
	Total int64            `json:"total"`
}

type ToolsServices struct {
	repo ToolsRepo
}

func NewToolsServices(repo ToolsRepo) *ToolsServices {
	return &ToolsServices{
		repo: repo,
	}
}

type ListToolReq struct {
	Page int64 `json:"page"`
	Size int64 `json:"size"`
}

// List 分页返回工具配置
func (t *ToolsServices) List(ctx context.Context, req *ListToolReq) (*ToolList, error) {
	items, total, err := t.repo.List(ctx, req.Page, req.Size)
	if err != nil {
		return nil, err
	}
	return &ToolList{
		Items: items,
		Total: total,
	}, nil
}

type UpdateToolReq struct {
	ToolName string          `json:"toolName"`
	Args     []po.CommonJson `json:"args"`
}

// Update 只更新工具的配置参数，其余字段原样保留
func (t *ToolsServices) Update(ctx context.Context, req *UpdateToolReq) error {
	return t.repo.Update(ctx, req.ToolName, req.Args)
}

type EnableReq struct {
	ToolName string `json:"toolName"`
	Enable   bool   `json:"enable"`
}

// EnableTools 启用/禁用工具；启用前要求已配置参数
func (t *ToolsServices) EnableTools(ctx context.Context, req *EnableReq) error {
	return t.repo.Enable(ctx, req.ToolName, req.Enable)
}
