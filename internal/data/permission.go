package data

import (
	"fmt"
	"taie/internal/agentkit/hooks"
	"taie/internal/po"
)

type permissionRepo struct {
	data *Data
}

func NewPermissionRepo(data *Data) hooks.PermissionRepo {
	return &permissionRepo{
		data: data,
	}
}
func (p *permissionRepo) GetMode() (string, error) {
	cfg, err := loadSystemConf()
	if err != nil {
		return "", err
	}
	mode := cfg.Mode
	if mode == "" {
		return "", fmt.Errorf("模式设置被破坏")
	}
	return mode, nil
}

func (p *permissionRepo) GetPermission() (*po.PermissionSetting, error) {
	cfg, err := loadSystemConf()
	if err != nil {
		return nil, err
	}
	return &cfg.Permissions, nil
}
