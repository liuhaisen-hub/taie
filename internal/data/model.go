package data

import (
	"context"
	"taie/internal/po"
	"taie/internal/services"
)

type modelRepo struct {
	data *Data
}

func NewModelRepo(data *Data) services.ModelRepo {
	return &modelRepo{
		data: data,
	}
}

// Create 新增一条模型记录
func (m modelRepo) Create(ctx context.Context, model *po.AIModel) error {
	return m.data.db.WithContext(ctx).Create(model).Error
}

// Update 按主键更新业务字段，不改动 created_at；Select 指定列使 enable=0（禁用）等零值也能更新
func (m modelRepo) Update(ctx context.Context, model *po.AIModel) error {
	return m.data.db.WithContext(ctx).Model(model).
		Select("name", "key", "base_url", "type", "provider", "enable").
		Updates(model).Error
}

// Delete 按主键软删除模型记录
func (m modelRepo) Delete(ctx context.Context, id uint) error {
	return m.data.db.WithContext(ctx).Delete(&po.AIModel{}, id).Error
}

// GetByID 按主键查询单条模型记录
func (m modelRepo) GetByID(ctx context.Context, id uint) (*po.AIModel, error) {
	var model po.AIModel
	if err := m.data.db.WithContext(ctx).First(&model, id).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

// List 按名字模糊搜索分页查询模型列表，name 为空时查全部，返回当页数据和总条数
func (m modelRepo) List(ctx context.Context, name string, page, pageSize int) ([]*po.AIModel, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := m.data.db.WithContext(ctx).Model(&po.AIModel{})
	if name != "" {
		query = query.Where("model_name LIKE ?", "%"+name+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []*po.AIModel
	if err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	return models, total, nil
}

func (m *modelRepo) GetByType(ctx context.Context, tp int32) ([]*po.AIModel, error) {
	var models []*po.AIModel
	err := m.data.db.WithContext(ctx).Where("type = ?", tp).Find(&models).Error
	if err != nil {
		return nil, err
	}
	return models, nil
}
