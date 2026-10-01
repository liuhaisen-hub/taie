package services

import (
	"context"

	"taie/internal/po"
)

type SystemRepo interface {
	// Get 返回系统设置；system.json 不存在时直接报错
	Get(ctx context.Context) (*po.SystemConfig, error)
	// Update 整体覆盖写回系统设置
	Update(ctx context.Context, cfg *po.SystemConfig) error
}

type SystemServices struct {
	repo SystemRepo
}

func NewSystemServices(repo SystemRepo) *SystemServices {
	return &SystemServices{
		repo: repo,
	}
}

// Get 返回系统设置
func (s *SystemServices) Get(ctx context.Context) (*po.SystemConfig, error) {
	return s.repo.Get(ctx)
}

type UpdateSystemReq struct {
	Config *po.SystemConfig `json:"config"`
}

// Update 整体覆盖写回系统设置
func (s *SystemServices) Update(ctx context.Context, req *UpdateSystemReq) error {
	return s.repo.Update(ctx, req.Config)
}
