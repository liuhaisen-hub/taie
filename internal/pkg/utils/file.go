package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var fileMu sync.Mutex

type ConfigFile[T any] struct {
	mu   sync.Mutex
	path string
}

func NewConfigFile[T any](path string) *ConfigFile[T] {
	return &ConfigFile[T]{
		path: path,
	}
}

func (c *ConfigFile[T]) Load() (T, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	stat, err := os.Stat(c.path)
	if err != nil {
		var zero T
		return zero, err
	}
	if stat.IsDir() {
		var zero T
		return zero, errors.New("path is directory, not file")
	}
	data, err := os.ReadFile(c.path)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("read file: %w", err)
	}
	var target T
	if err := json.Unmarshal(data, &target); err != nil {
		var zero T
		return zero, fmt.Errorf("json unmarshal: %w", err)
	}
	return target, nil
}
func (c *ConfigFile[T]) Save(v T) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.writeFile(v)
}
func (c *ConfigFile[T]) writeFile(v T) error {
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	// 创建临时文件，防止奔溃损坏数据
	tmpPath := c.path + ".tmp"
	if err := os.WriteFile(tmpPath, bytes, 0644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	if err := os.Rename(tmpPath, c.path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename tmp file: %w", err)
	}
	return nil
}
