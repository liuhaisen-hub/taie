package data

import (
	"context"
	"taie/internal/po"
	"taie/internal/services"
)

const DefaultMaxHistory = 40

type sessionRepo struct {
	data *Data
}

func NewSessionRepo(data *Data) services.SessionRepo {
	return &sessionRepo{
		data: data,
	}
}
func (s *sessionRepo) Create(ctx context.Context, ss *po.ChatSession) (*po.ChatSession, error) {
	if err := s.data.db.WithContext(ctx).Create(&ss).Error; err != nil {
		return nil, err
	}
	return ss, nil

}
func (s *sessionRepo) Update(ctx context.Context, id uint, title string) error {
	return s.data.db.WithContext(ctx).Model(po.ChatSession{}).
		Where("id = ?", id).Update("title = ?", title).Error
}

// Delete 按主键软删除会话记录
func (s *sessionRepo) Delete(ctx context.Context, id uint) error {
	return s.data.db.WithContext(ctx).Delete(&po.ChatSession{}, id).Error
}

// GetByID 按主键查询单条会话记录
func (s *sessionRepo) GetByID(ctx context.Context, id uint) (*po.ChatSession, error) {
	var session po.ChatSession
	if err := s.data.db.WithContext(ctx).First(&session, id).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// List 按标题模糊搜索分页查询会话列表，title 为空时查全部，返回当页数据和总条数
func (s *sessionRepo) List(ctx context.Context, title string, page, pageSize int) ([]*po.ChatSession, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := s.data.db.WithContext(ctx).Model(&po.ChatSession{})
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
