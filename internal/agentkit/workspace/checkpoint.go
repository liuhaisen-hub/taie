package workspace

import (
	"context"
	"errors"
	"sync"
	"taie/internal/po"

	"github.com/cloudwego/eino/adk"
	"gorm.io/gorm"
)

type CheckPointRepo interface {
	GetById(ctx context.Context, ID string) (*po.Checkpoint, error)
	Save(ctx context.Context, data *po.Checkpoint) error
	Delete(ctx context.Context, ID string) error
}

// checkpoint store 接口实现
type Cpstore struct {
	mu    sync.Mutex
	store adk.CheckPointStore
	repo  CheckPointRepo
}

func NewCheckpointStore(repo CheckPointRepo) adk.CheckPointStore {
	return &Cpstore{
		repo: repo,
	}
}

// 实现checkpoint 的接口 Get Set
//
//	type CheckPointStore interface {
//		Get(ctx context.Context, checkPointID string) ([]byte, bool, error)
//		Set(ctx context.Context, checkPointID string, checkPoint []byte) error
//	}
func (c *Cpstore) Get(ctx context.Context, checkpointId string) ([]byte, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	data, err := c.repo.GetById(ctx, checkpointId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil // 不存在不是错误,接口语义如此
		}
		return nil, false, err
	}
	return data.SnapshotJSON, true, nil
}

func (c *Cpstore) Set(ctx context.Context, checkpointId string, checkpoint []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.repo.Save(ctx, &po.Checkpoint{CheckpointID: checkpointId, SnapshotJSON: checkpoint})
}

// 实现
// type CheckPointDeleter interface {
// 	Delete(ctx context.Context, checkPointID string) error
// }

func (c *Cpstore) Delete(ctx context.Context, checkpointID string) error {
	return c.repo.Delete(ctx, checkpointID)
}
