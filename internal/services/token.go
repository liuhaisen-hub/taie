package services

import (
	"context"
	"time"
)

// TokenUseRepo token 用量统计仓库
type TokenUseRepo interface {
	// SumDaily 按天聚合 token 用量，日期为本地时区；days>0 时只统计最近 days 天
	SumDaily(ctx context.Context, days int) ([]*TokenDailyStat, error)
	// SumBySession 按会话聚合 token 用量；days>0 时只统计最近 days 天
	SumBySession(ctx context.Context, days int) ([]*TokenSessionStat, error)
	// SumByModel 按模型名称聚合 token 用量；days>0 时只统计最近 days 天
	SumByModel(ctx context.Context, days int) ([]*TokenModelStat, error)
}

// TokenDailyStat 单日 token 用量聚合
type TokenDailyStat struct {
	Date             string `json:"date"` // YYYY-MM-DD（本地时区）
	PromptTokens     int64  `json:"promptTokens"`
	CompletionTokens int64  `json:"completionTokens"`
	TotalTokens      int64  `json:"totalTokens"`
}

// TokenSessionStat 按会话聚合的 token 用量
type TokenSessionStat struct {
	SessionID        uint64 `json:"sessionId"`
	Title            string `json:"title"`
	PromptTokens     int64  `json:"promptTokens"`
	CompletionTokens int64  `json:"completionTokens"`
	TotalTokens      int64  `json:"totalTokens"`
}

// TokenModelStat 按模型名称聚合的 token 用量
type TokenModelStat struct {
	ModelName        string `json:"modelName"`
	PromptTokens     int64  `json:"promptTokens"`
	CompletionTokens int64  `json:"completionTokens"`
	TotalTokens      int64  `json:"totalTokens"`
}

type TokenUseServices struct {
	repo TokenUseRepo
}

func NewTokenUseServices(repo TokenUseRepo) *TokenUseServices {
	return &TokenUseServices{
		repo: repo,
	}
}

type TokenUsageReq struct {
	// Days 最近天数；0 表示不限时间
	Days int `json:"days"`
}

// DailyUsage 按天聚合 token 用量；限定了天数时把区间内没有用量的日期补 0，保证图表 x 轴连续
func (s *TokenUseServices) DailyUsage(ctx context.Context, req *TokenUsageReq) ([]*TokenDailyStat, error) {
	stats, err := s.repo.SumDaily(ctx, req.Days)
	if err != nil {
		return nil, err
	}
	if req.Days <= 0 {
		return stats, nil
	}
	byDate := make(map[string]*TokenDailyStat, len(stats))
	for _, st := range stats {
		byDate[st.Date] = st
	}
	now := time.Now()
	filled := make([]*TokenDailyStat, 0, req.Days)
	for i := req.Days - 1; i >= 0; i-- {
		day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -i)
		key := day.Format("2006-01-02")
		if st, ok := byDate[key]; ok {
			filled = append(filled, st)
			continue
		}
		filled = append(filled, &TokenDailyStat{Date: key})
	}
	return filled, nil
}

// SessionUsage 按会话聚合 token 用量
func (s *TokenUseServices) SessionUsage(ctx context.Context, req *TokenUsageReq) ([]*TokenSessionStat, error) {
	return s.repo.SumBySession(ctx, req.Days)
}

// ModelUsage 按模型名称聚合 token 用量
func (s *TokenUseServices) ModelUsage(ctx context.Context, req *TokenUsageReq) ([]*TokenModelStat, error) {
	return s.repo.SumByModel(ctx, req.Days)
}
