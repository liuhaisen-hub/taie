package data

import (
	"context"
	"time"

	"gorm.io/gorm"

	"taie/internal/po"
	"taie/internal/services"
)

type tokenUseRepo struct {
	data *Data
}

func NewTokenUseRepo(data *Data) services.TokenUseRepo {
	return &tokenUseRepo{
		data: data,
	}
}

// sinceDays 返回最近 days 天的本地零点起始时间；days<=0 时返回 false 表示不限时间
func sinceDays(days int) (time.Time, bool) {
	if days <= 0 {
		return time.Time{}, false
	}
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(days - 1))
	return start, true
}

// applySince 给聚合查询追加最近 days 天的时间过滤；days<=0 不过滤
func (t *tokenUseRepo) applySince(query *gorm.DB, days int) *gorm.DB {
	if start, ok := sinceDays(days); ok {
		query = query.Where("created_at >= ?", start)
	}
	return query
}

// SumDaily 按天聚合 token 用量；created_at 落库为带时区的时间文本，localtime 修饰符转本地日期
func (t *tokenUseRepo) SumDaily(ctx context.Context, days int) ([]*services.TokenDailyStat, error) {
	query := t.data.db.WithContext(ctx).Model(&po.SessionTokenUsage{}).
		Select("date(created_at, 'localtime') AS date, "+
			"SUM(prompt_tokens) AS prompt_tokens, "+
			"SUM(completion_tokens) AS completion_tokens, "+
			"SUM(total_tokens) AS total_tokens").
		Group("date").
		Order("date ASC")
	query = t.applySince(query, days)

	var stats []*services.TokenDailyStat
	if err := query.Scan(&stats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

// SumBySession 按会话聚合 token 用量，左联会话表带出标题（会话被删后用量仍保留）
func (t *tokenUseRepo) SumBySession(ctx context.Context, days int) ([]*services.TokenSessionStat, error) {
	query := t.data.db.WithContext(ctx).Model(&po.SessionTokenUsage{}).
		Select("session_token_usage.session_id AS session_id, "+
			"COALESCE(chat_session.title, '') AS title, "+
			"SUM(prompt_tokens) AS prompt_tokens, "+
			"SUM(completion_tokens) AS completion_tokens, "+
			"SUM(total_tokens) AS total_tokens").
		Joins("LEFT JOIN chat_session ON chat_session.id = session_token_usage.session_id").
		Group("session_token_usage.session_id, chat_session.title").
		Order("total_tokens DESC")
	query = t.applySince(query, days)

	var stats []*services.TokenSessionStat
	if err := query.Scan(&stats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

// SumByModel 按模型名称聚合 token 用量
func (t *tokenUseRepo) SumByModel(ctx context.Context, days int) ([]*services.TokenModelStat, error) {
	query := t.data.db.WithContext(ctx).Model(&po.SessionTokenUsage{}).
		Select("model_name, "+
			"SUM(prompt_tokens) AS prompt_tokens, "+
			"SUM(completion_tokens) AS completion_tokens, "+
			"SUM(total_tokens) AS total_tokens").
		Group("model_name").
		Order("total_tokens DESC")
	query = t.applySince(query, days)

	var stats []*services.TokenModelStat
	if err := query.Scan(&stats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}
