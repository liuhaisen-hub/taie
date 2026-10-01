package data

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"

	"taie/internal/pkg/utils"
	"taie/internal/po"
	"taie/internal/services"
)

// toolsConfPath 是工具配置文件，JSON 数组，所有工具集中在这一个文件里配置。
const toolsConfPath = ".conf/tools.json"

// loadToolConfs 读取全部工具配置；文件不存在时视为无工具，返回空切片
func loadToolConfs() ([]po.ToolConfig, error) {
	cfgs, err := utils.NewConfigFile[[]po.ToolConfig](toolsConfPath).Load()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return cfgs, nil
}

// saveToolConfs 整体写回工具配置
func saveToolConfs(cfgs []po.ToolConfig) error {
	return utils.NewConfigFile[[]po.ToolConfig](toolsConfPath).Save(cfgs)
}

// updateToolConf 按 toolName 定位并用 fn 修改后整体写回；fn 返回 error 时不落盘
func updateToolConf(toolName string, fn func(*po.ToolConfig) error) error {
	cfgs, err := loadToolConfs()
	if err != nil {
		return err
	}
	i := slices.IndexFunc(cfgs, func(c po.ToolConfig) bool { return c.ToolName == toolName })
	if i < 0 {
		return fmt.Errorf("工具不存在: %s", toolName)
	}
	if err := fn(&cfgs[i]); err != nil {
		return err
	}
	return saveToolConfs(cfgs)
}

type toolsRepo struct {
}

func NewToolsRepo() services.ToolsRepo {
	return &toolsRepo{}
}

// List 分页返回 tools.json 中的工具配置，按文件中的顺序返回
func (t *toolsRepo) List(ctx context.Context, page, size int64) ([]*po.ToolConfig, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}

	cfgs, err := loadToolConfs()
	if err != nil {
		return nil, 0, err
	}

	total := int64(len(cfgs))
	start := (page - 1) * size
	if start >= total {
		return []*po.ToolConfig{}, total, nil
	}
	end := min(start+size, total)

	items := make([]*po.ToolConfig, 0, end-start)
	for i := start; i < end; i++ {
		items = append(items, &cfgs[i])
	}
	return items, total, nil
}

// Update 只更新工具的配置参数，其余字段原样保留
func (t *toolsRepo) Update(ctx context.Context, toolName string, args []po.CommonJson) error {
	return updateToolConf(toolName, func(c *po.ToolConfig) error {
		c.Args = args
		return nil
	})
}

// Enable 启用/禁用工具；启用前要求已配置参数
func (t *toolsRepo) Enable(ctx context.Context, toolName string, enable bool) error {
	return updateToolConf(toolName, func(c *po.ToolConfig) error {
		if enable && len(c.Args) == 0 {
			return fmt.Errorf("未设置相关配置")
		}
		c.Enable = enable
		return nil
	})
}
