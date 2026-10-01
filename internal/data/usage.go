package data

import (
	"context"
	"taie/internal/agentkit/hooks"
	"taie/internal/po"
)

type useRepo struct {
	data *Data
}

func NewUseRepo(data *Data) hooks.UsageRepo {
	return &useRepo{
		data: data,
	}
}

// AppendTokenUsage 落库一行 token 消耗明细（每次模型调用一条）
func (u *useRepo) AppendTokenUsage(ctx context.Context, data *po.SessionTokenUsage) error {
	return u.data.db.WithContext(ctx).Create(data).Error
}
