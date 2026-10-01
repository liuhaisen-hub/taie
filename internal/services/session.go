package services

import (
	"context"

	"taie/internal/po"
)

type ChatSessionRepo interface {
	Delete(ctx context.Context, id uint) error
	// GetMessageByID 按 session 主键查询其全部聊天消息
	GetMessageByID(ctx context.Context, id uint) ([]*po.ChatMessage, error)
	Update(ctx context.Context, id uint, title string) error
	// List 按标题模糊搜索分页查询会话列表，title 为空时查全部，返回当页数据和总条数
	List(ctx context.Context, title string, page, pageSize int) ([]*po.ChatSession, int64, error)
}
type SessionServices struct {
	repo ChatSessionRepo
}

func NewSessionServices(repo ChatSessionRepo) *SessionServices {
	return &SessionServices{
		repo: repo,
	}
}

// SessionList 是 List 的分页查询结果。
type SessionList struct {
	Items []*po.ChatSession `json:"items"`
	Total int64             `json:"total"`
}

type ListSessionReq struct {
	Title string `json:"title"`
	Page  int    `json:"page"`
	Size  int    `json:"size"`
}

// List 按标题模糊搜索分页查询会话列表，title 为空时查全部
func (s *SessionServices) List(ctx context.Context, req *ListSessionReq) (*SessionList, error) {
	items, total, err := s.repo.List(ctx, req.Title, req.Page, req.Size)
	if err != nil {
		return nil, err
	}
	return &SessionList{
		Items: items,
		Total: total,
	}, nil
}

type ChatSessionReq struct {
	ID uint `json:"id"`
}

func (s *SessionServices) GetMessage(ctx context.Context, req *ChatSessionReq) ([]*po.ChatMessage, error) {
	return s.repo.GetMessageByID(ctx, req.ID)
}

type DeleteSessionReq struct {
	ID uint `json:"id"`
}

// Delete 按主键软删除会话记录。内存中缓存的 agent 由 ChatAgent 的空闲淘汰
// 自行回收（陈旧 entry 不再被命中，空闲超时后被清理）。
func (s *SessionServices) Delete(ctx context.Context, req *DeleteSessionReq) error {
	return s.repo.Delete(ctx, req.ID)
}
