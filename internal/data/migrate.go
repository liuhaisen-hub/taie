package data

import (
	"fmt"
	"os"
	"path/filepath"
)

// MigrationsDir 是 SQL 迁移文件所在目录（相对于项目根目录）。
const MigrationsDir = "migrations"

// Migrate 打开 dsn 指向的数据库，按文件名字母序执行 MigrationsDir 下的全部
// .sql 文件。迁移脚本要求幂等（使用 IF NOT EXISTS 等写法），可重复执行。
// 空的 dsn 回退到 DefaultDSN。
func Migrate(dsn string) error {
	files, err := filepath.Glob(filepath.Join(MigrationsDir, "*.sql"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no migration files found in %s", MigrationsDir)
	}

	d, cleanup, err := NewData(dsn)
	if err != nil {
		return err
	}
	defer cleanup()

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read %s: %w", file, err)
		}
		if err := d.db.Exec(string(content)).Error; err != nil {
			return fmt.Errorf("exec %s: %w", file, err)
		}
		fmt.Println("applied", file)
	}
	return nil
}
