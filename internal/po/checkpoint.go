package po

import "time"

type Checkpoint struct {
	CheckpointID string    `gorm:"column:checkpoint_id;primaryKey"` // through Model
	CreatedAt    time.Time // through Model
	UpdatedAt    time.Time // through Model
	SessionID    uint64    `gorm:"index;comment:归属会话ID"`
	SnapshotJSON []byte    `gorm:"type:TEXT;not null;comment:快照json bytes"`
}

func (Checkpoint) TableName() string {
	return "checkpoint"
}
