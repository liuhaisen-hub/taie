package data

import (
	"context"
	"taie/internal/agentkit/workspace"
	"taie/internal/po"
)

type checkpointRepo struct {
	data *Data
}

func NewCheckPointRepo(data *Data) workspace.CheckPointRepo {
	return &checkpointRepo{
		data: data,
	}
}

// GetById 按主键查询；未命中返回 gorm.ErrRecordNotFound，由 Cpstore 转成 ok=false
// （Find 查不到不报错，无法区分"不存在"和"出错"，必须用 First）
func (c *checkpointRepo) GetById(ctx context.Context, ID string) (*po.Checkpoint, error) {
	var row po.Checkpoint
	if err := c.data.db.WithContext(ctx).First(&row, "checkpoint_id = ?", ID).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (c *checkpointRepo) Save(ctx context.Context, data *po.Checkpoint) error {
	return c.data.db.WithContext(ctx).Save(&data).Error
}

func (c *checkpointRepo) Delete(ctx context.Context, ID string) error {
	return c.data.db.WithContext(ctx).Model(&po.Checkpoint{}).Delete("checkpoint_id = ?", ID).Error
}
