package services

import (
	"context"
	"taie/internal/po"
)

type ModelRepo interface {
	// Create 新增一条模型记录
	Create(ctx context.Context, model *po.AIModel) error
	// Update 保存模型的所有字段
	Update(ctx context.Context, model *po.AIModel) error
	// Delete 按主键软删除模型记录
	Delete(ctx context.Context, id uint) error
	// GetByID 按主键查询单条模型记录
	GetByID(ctx context.Context, id uint) (*po.AIModel, error)
	// List 按名字模糊搜索分页查询模型列表，name 为空时查全部，返回当页数据和总条数
	List(ctx context.Context, name string, page, pageSize int) ([]*po.AIModel, int64, error)
	GetByType(ctx context.Context, tp int32) ([]*po.AIModel, error)
}

// ModelList 是 List 的分页查询结果。
type ModelList struct {
	Items []*po.AIModel `json:"items"`
	Total int64         `json:"total"`
}

type ModelServices struct {
	repo ModelRepo
}

func NewModelServices(repo ModelRepo) *ModelServices {
	return &ModelServices{
		repo: repo,
	}
}

// Create 新增一条模型记录
func (m *ModelServices) Create(ctx context.Context, model *po.AIModel) error {
	return m.repo.Create(ctx, model)
}

// Update 保存模型的所有字段
func (m *ModelServices) Update(ctx context.Context, model *po.AIModel) error {
	return m.repo.Update(ctx, model)
}

// Delete 按主键软删除模型记录
func (m *ModelServices) Delete(ctx context.Context, id uint) error {
	return m.repo.Delete(ctx, id)
}

// GetByID 按主键查询单条模型记录
func (m *ModelServices) GetByID(ctx context.Context, id uint) (*po.AIModel, error) {
	return m.repo.GetByID(ctx, id)
}

type ListModelReq struct {
	ModelName string `json:"modeName"`
	Page      int    `json:"page"`
	Size      int    `json:"size"`
}

// List 按名字模糊搜索分页查询模型列表，name 为空时查全部
func (m *ModelServices) List(ctx context.Context, req *ListModelReq) (*ModelList, error) {
	items, total, err := m.repo.List(ctx, req.ModelName, req.Page, req.Size)
	if err != nil {
		return nil, err
	}
	return &ModelList{
		Items: items,
		Total: total,
	}, nil
}
