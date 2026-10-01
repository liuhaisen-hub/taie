import type { TokenModelStat, TokenSessionStat } from '@bindings/taie/internal/services'
import { TokenUseServices } from '@bindings/taie/internal/services'
import { IconRefresh } from '@douyinfe/semi-icons'
import { Button, Radio, RadioGroup, Table, Toast } from '@douyinfe/semi-ui'
import type { ColumnProps } from '@douyinfe/semi-ui/lib/es/table/interface'
import type { UIEvent } from 'react'
import { useCallback, useEffect, useMemo, useState } from 'react'
import styles from './index.module.css'

type UsageMode = 'session' | 'model'

// 无限滚动：初始渲染行数，触底后每次追加的行数
const TABLE_BATCH = 20

const formatNumber = (value?: number | string | null) => Number(value ?? 0).toLocaleString()

const sessionColumns: ColumnProps<TokenSessionStat>[] = [
    { title: '会话', dataIndex: 'title', render: (_text, record) => record.title || `#${record.sessionId}` },
    { title: '输入', dataIndex: 'promptTokens', render: text => formatNumber(text) },
    { title: '输出', dataIndex: 'completionTokens', render: text => formatNumber(text) },
    { title: '总计', dataIndex: 'totalTokens', render: text => formatNumber(text) },
]

const modelColumns: ColumnProps<TokenModelStat>[] = [
    { title: '模型', dataIndex: 'modelName', render: text => text || '-' },
    { title: '输入', dataIndex: 'promptTokens', render: text => formatNumber(text) },
    { title: '输出', dataIndex: 'completionTokens', render: text => formatNumber(text) },
    { title: '总计', dataIndex: 'totalTokens', render: text => formatNumber(text) },
]

// 历史 tab：按会话或模型查看全部时间的 token 用量明细，触底加载更多
export const UsageHistory = () => {
    const [sessions, setSessions] = useState<TokenSessionStat[]>([])
    const [models, setModels] = useState<TokenModelStat[]>([])
    const [loading, setLoading] = useState(false)
    const [mode, setMode] = useState<UsageMode>('session')
    const [visibleCount, setVisibleCount] = useState(TABLE_BATCH)

    const load = useCallback(async () => {
        setLoading(true)
        try {
            // 不限定时间（days=0 表示全部）
            const [sessionRes, modelRes] = await Promise.all([
                TokenUseServices.SessionUsage({ days: 0 }),
                TokenUseServices.ModelUsage({ days: 0 }),
            ])
            setSessions((sessionRes ?? []).filter((item): item is TokenSessionStat => item != null))
            setModels((modelRes ?? []).filter((item): item is TokenModelStat => item != null))
            setVisibleCount(TABLE_BATCH)
        } catch (err) {
            Toast.error(`加载用量明细失败：${String(err)}`)
        } finally {
            setLoading(false)
        }
    }, [])

    useEffect(() => {
        void load()
    }, [load])

    // 数据已全量在前端，切片放出实现无限滚动
    const visibleSessions = useMemo(() => sessions.slice(0, visibleCount), [sessions, visibleCount])
    const visibleModels = useMemo(() => models.slice(0, visibleCount), [models, visibleCount])

    // 触底再放出一批；已全部展示后 setState 同值会被 React 跳过
    const handleTableScroll = (e: UIEvent<HTMLDivElement>) => {
        const el = e.currentTarget
        if (el.scrollTop + el.clientHeight < el.scrollHeight - 32) return
        setVisibleCount(count => Math.min(count + TABLE_BATCH, (mode === 'session' ? sessions : models).length))
    }

    return (
        <div className={styles.root}>
            <div className={styles.toolbar}>
                <RadioGroup
                    type="button"
                    value={mode}
                    onChange={e => {
                        setMode(e.target.value as UsageMode)
                        setVisibleCount(TABLE_BATCH)
                    }}
                >
                    <Radio value="session">按会话</Radio>
                    <Radio value="model">按模型</Radio>
                </RadioGroup>
                <Button icon={<IconRefresh />} theme="borderless" onClick={() => void load()}>
                    刷新
                </Button>
            </div>
            <div className={styles.table_warp} onScroll={handleTableScroll}>
                {mode === 'session' ? (
                    // key 区分两张表：React 19 + Semi 下复用同一 Table 实例切换
                    // columns/dataSource 会导致旧行残留（实测行数 20→39→58 累加）
                    <Table<TokenSessionStat>
                        key="session"
                        columns={sessionColumns}
                        dataSource={visibleSessions}
                        rowKey={record => String(record?.sessionId ?? '')}
                        loading={loading}
                        pagination={false}
                        emptyContent="暂无数据"
                    />
                ) : (
                    <Table<TokenModelStat>
                        key="model"
                        columns={modelColumns}
                        dataSource={visibleModels}
                        rowKey={record => record?.modelName ?? ''}
                        loading={loading}
                        pagination={false}
                        emptyContent="暂无数据"
                    />
                )}
            </div>
        </div>
    )
}
