package data

import (
	"context"

	"taie/internal/pkg/utils"
	"taie/internal/po"
	"taie/internal/services"
)

// systemConfPath 是系统设置配置文件。
const systemConfPath = ".conf/system.json"

// loadSystemConf 读取系统设置；文件不存在时直接报错
func loadSystemConf() (po.SystemConfig, error) {
	return utils.NewConfigFile[po.SystemConfig](systemConfPath).Load()
}

// saveSystemConf 整体写回系统设置
func saveSystemConf(cfg po.SystemConfig) error {
	return utils.NewConfigFile[po.SystemConfig](systemConfPath).Save(cfg)
}

type systemRepo struct {
}

func NewSystemRepo() services.SystemRepo {
	return &systemRepo{}
}

// Get 返回系统设置
func (s *systemRepo) Get(ctx context.Context) (*po.SystemConfig, error) {
	cfg, err := loadSystemConf()
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Update 整体覆盖写回系统设置
func (s *systemRepo) Update(ctx context.Context, cfg *po.SystemConfig) error {
	return saveSystemConf(*cfg)
}
