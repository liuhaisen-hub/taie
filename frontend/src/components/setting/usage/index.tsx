import type { TokenDailyStat, TokenModelStat } from '@bindings/taie/internal/services'
import { TokenUseServices } from '@bindings/taie/internal/services'
import { IconRefresh } from '@douyinfe/semi-icons'
import { Button, Toast } from '@douyinfe/semi-ui'
import { Area, AreaChart, Bar, BarChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import { useCallback, useEffect, useMemo, useState } from 'react'
import styles from './index.module.css'

const CHART_DAYS = 7

// 图表配色：中性网格在深浅主题下都可读，系列色固定（SVG 属性里不能用 CSS 变量）
const GRID_COLOR = 'rgba(128, 128, 128, 0.2)'
const PROMPT_COLOR = '#4c8bf5'
const COMPLETION_COLOR = '#2fbfa0'

const formatNumber = (value?: number | string | null) => Number(value ?? 0).toLocaleString()

const formatCompact = (value: number) => {
    if (value >= 10000) return `${Math.round(value / 1000)}k`
    if (value >= 1000) return `${(value / 1000).toFixed(1)}k`
    return String(value)
}

// 最近 7 天 token 用量统计：按天面积图看趋势，按模型柱状图看分布
export const UsageStats = () => {
    const [daily, setDaily] = useState<TokenDailyStat[]>([])
    const [models, setModels] = useState<TokenModelStat[]>([])
    const [loading, setLoading] = useState(false)

    const load = useCallback(async () => {
        setLoading(true)
        try {
            // 后端会把区间内没有用量的日期补 0，保证 x 轴连续
            const [dailyRes, modelRes] = await Promise.all([
                TokenUseServices.DailyUsage({ days: CHART_DAYS }),
                TokenUseServices.ModelUsage({ days: CHART_DAYS }),
            ])
            setDaily((dailyRes ?? []).filter((item): item is TokenDailyStat => item != null))
            setModels((modelRes ?? []).filter((item): item is TokenModelStat => item != null))
        } catch (err) {
            Toast.error(`加载用量统计失败：${String(err)}`)
        } finally {
            setLoading(false)
        }
    }, [])

    useEffect(() => {
        void load()
    }, [load])

    const chartTotal = useMemo(
        () => daily.reduce((sum, item) => sum + Number(item.totalTokens ?? 0), 0),
        [daily]
    )

    return (
        <div className={styles.root}>
            <div className={styles.section}>
                <div className={styles.sectionHeader}>
                    <div className={styles.sectionTitle}>最近 {CHART_DAYS} 天 Token 用量</div>
                    <Button icon={<IconRefresh />} theme="borderless" loading={loading} onClick={() => void load()}>
                        刷新
                    </Button>
                </div>
                <div className={styles.summary}>合计 {formatNumber(chartTotal)} tokens</div>
                <div className={styles.legend}>
                    <span className={styles.legendItem}>
                        <i className={styles.dotPrompt} />输入
                    </span>
                    <span className={styles.legendItem}>
                        <i className={styles.dotCompletion} />输出
                    </span>
                </div>
                <div className={`${styles.chart} ${loading ? styles.chartLoading : ''}`}>
                    <ResponsiveContainer width="100%" height="100%">
                        <AreaChart data={daily} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
                            <CartesianGrid strokeDasharray="3 3" stroke={GRID_COLOR} vertical={false} />
                            <XAxis
                                dataKey="date"
                                tick={{ fill: 'currentColor', fontSize: 12 }}
                                tickLine={false}
                                axisLine={{ stroke: GRID_COLOR }}
                            />
                            <YAxis
                                width={44}
                                tick={{ fill: 'currentColor', fontSize: 12 }}
                                tickLine={false}
                                axisLine={false}
                                tickFormatter={formatCompact}
                            />
                            <Tooltip
                                formatter={(value, name) => [formatNumber(value as number), String(name)] as [string, string]}
                            />
                            {/* 输入/输出堆叠成总量；totalTokens 单独一组叠层且不填充，只为在 Tooltip 里带出总计 */}
                            <Area
                                dataKey="promptTokens"
                                name="输入"
                                stackId="usage"
                                stroke={PROMPT_COLOR}
                                fill={PROMPT_COLOR}
                                fillOpacity={0.35}
                                strokeWidth={2}
                            />
                            <Area
                                dataKey="completionTokens"
                                name="输出"
                                stackId="usage"
                                stroke={COMPLETION_COLOR}
                                fill={COMPLETION_COLOR}
                                fillOpacity={0.35}
                                strokeWidth={2}
                            />
                            <Area
                                dataKey="totalTokens"
                                name="总计"
                                stackId="total"
                                stroke="none"
                                fill="none"
                                dot={false}
                                activeDot={false}
                                legendType="none"
                            />
                        </AreaChart>
                    </ResponsiveContainer>
                </div>
            </div>
            <div className={styles.section}>
                <div className={styles.sectionTitle}>按模型统计</div>
                <div className={styles.summary}>共 {models.length} 个模型</div>
                <div className={styles.legend}>
                    <span className={styles.legendItem}>
                        <i className={styles.dotPrompt} />输入
                    </span>
                    <span className={styles.legendItem}>
                        <i className={styles.dotCompletion} />输出
                    </span>
                </div>
                <div className={`${styles.chart} ${loading ? styles.chartLoading : ''}`}>
                    <ResponsiveContainer width="100%" height="100%">
                        <BarChart data={models} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
                            <CartesianGrid strokeDasharray="3 3" stroke={GRID_COLOR} vertical={false} />
                            {/* 模型名较长，斜着排避免互相遮挡 */}
                            <XAxis
                                dataKey="modelName"
                                tick={{ fill: 'currentColor', fontSize: 12 }}
                                tickLine={false}
                                axisLine={{ stroke: GRID_COLOR }}
                                interval={0}
                                angle={-20}
                                textAnchor="end"
                                height={56}
                            />
                            <YAxis
                                width={44}
                                tick={{ fill: 'currentColor', fontSize: 12 }}
                                tickLine={false}
                                axisLine={false}
                                tickFormatter={formatCompact}
                            />
                            <Tooltip
                                cursor={{ fill: 'var(--semi-color-fill-0)' }}
                                formatter={(value, name) => [formatNumber(value as number), String(name)] as [string, string]}
                            />
                            <Bar dataKey="promptTokens" name="输入" stackId="usage" fill={PROMPT_COLOR} fillOpacity={0.85} />
                            <Bar
                                dataKey="completionTokens"
                                name="输出"
                                stackId="usage"
                                fill={COMPLETION_COLOR}
                                fillOpacity={0.85}
                            />
                        </BarChart>
                    </ResponsiveContainer>
                </div>
            </div>
        </div>
    )
}
