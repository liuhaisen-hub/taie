package data

import (
	"context"
	"taie/internal/po"
	"taie/internal/services"
)

type chatSession struct {
	data *Data
}

func NewChatSessionRepo(data *Data) services.ChatSessionRepo {
	return &chatSession{
		data: data,
	}
}

func (c *chatSession) Delete(ctx context.Context, id uint) error {
	return c.data.db.WithContext(ctx).Delete(&po.ChatSession{}, id).Error
}

// GetMessageByID 按主键查询单条会话记录
func (c *chatSession) GetMessageByID(ctx context.Context, id uint) ([]*po.ChatMessage, error) {

	var list []*po.ChatMessage
	err := c.data.db.WithContext(ctx).Where("session_id = ?", id).Order("id ASC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, err
}

// 只允许更新标题
func (c *chatSession) Update(ctx context.Context, id uint, title string) error {
	return c.data.db.WithContext(ctx).Model(po.ChatSession{}).
		Where("id = ?", id).Update("title = ?", title).Error
}

// List 按标题模糊搜索分页查询会话列表，title 为空时查全部，返回当页数据和总条数
func (c *chatSession) List(ctx context.Context, title string, page, pageSize int) ([]*po.ChatSession, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := c.data.db.WithContext(ctx).Model(&po.ChatSession{})
	if title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var sessions []*po.ChatSession
	if err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&sessions).Error; err != nil {
		return nil, 0, err
	}
	return sessions, total, nil
}
