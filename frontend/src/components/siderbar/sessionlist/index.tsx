import type { ChatSession } from '@bindings/taie/internal/po'
import { SessionServices } from '@bindings/taie/internal/services'
import { IconComment, IconDelete, IconMore } from '@douyinfe/semi-icons'
import { Button, Dropdown, List, Popconfirm, Skeleton, Toast, Typography } from '@douyinfe/semi-ui'
import { useCallback, useEffect, useState } from 'react'
import { useLayout } from '@/hooks/layout'
import styles from './index.module.css'

const placeholder = (
    <div className={styles.placeholder}>
        <Skeleton.Avatar className={styles.placeholderAvatar} />
        <div>
            <Skeleton.Title className={styles.placeholderTitle} />
            <Skeleton.Paragraph className={styles.placeholderParagraph} rows={2} />
        </div>
    </div>
);

export const SessionList = () => {
    const [sessions, setSessions] = useState<ChatSession[]>([])
    const [loading, setLoading] = useState(false)
    const { activeId, onSessionSelect, createNewSession, sessionsVersion } = useLayout()

    const load = useCallback(async () => {
        setLoading(true)
        try {
            // 侧边栏一次拉全量，标题为空查全部
            const result = await SessionServices.List({ title: '', page: 1, size: 100 })
            setSessions((result?.items ?? []).filter((item): item is ChatSession => item != null))
        } catch (err) {
            Toast.error(`加载会话列表失败：${String(err)}`)
        } finally {
            setLoading(false)
        }
    }, [])

    // 挂载时拉一次；之后依赖布局层的版本号（done 落库等场景）重拉
    useEffect(() => {
        void load()
    }, [load, sessionsVersion])

    const handleDelete = async (session: ChatSession) => {
        try {
            await SessionServices.Delete({ id: session.id })
            // 删除的是当前会话时，聊天区一并回到新建会话态
            if (session.id === activeId) createNewSession()
            Toast.success('删除成功')
            await load()
        } catch (err) {
            Toast.error(`「${session.title}」删除失败：${String(err)}`)
        }
    }

    return (
        <div className={styles.session}>
            <Skeleton placeholder={placeholder} loading={loading} active className={styles.skeleton}>
                <List
                    dataSource={sessions}
                    size="small"
                    emptyContent="暂无会话"
                    renderItem={item => (
                        <List.Item
                            className={item.id === activeId ? `${styles.list} ${styles.active}` : styles.list}
                            key={item.id}
                            // 点击经布局层切换会话，AIChatProvider 监听 activeId 加载历史记录
                            onClick={() => onSessionSelect(item.id)}
                            main={
                                <div className={styles.main}>
                                    <IconComment />
                                    <Typography.Text ellipsis={{
                                        showTooltip: {
                                            opts: { content: item.title }
                                        }
                                    }} className={styles.text}>{item.title}</Typography.Text>
                                    <Dropdown
                                        trigger="click"
                                        // Portal 合成事件沿 React 树冒泡，菜单内点击也要拦住，避免误切会话
                                        stopPropagation
                                        render={
                                            <Dropdown.Menu>
                                                {/* 点击「删除」项弹出确认；点确认/取消落在下拉 portal 外，菜单经外点机制自动收起 */}
                                                <Popconfirm
                                                    title="确定删除该会话吗？"
                                                    onConfirm={() => handleDelete(item)}
                                                >
                                                    <Dropdown.Item icon={<IconDelete />}>删除</Dropdown.Item>
                                                </Popconfirm>
                                            </Dropdown.Menu>
                                        }>
                                        {/* 阻止冒泡，避免点开下拉时误触发会话切换 */}
                                        <Button theme="borderless" type="tertiary" icon={<IconMore />} onClick={e => e.stopPropagation()} />
                                    </Dropdown>
                                </div>
                            }
                        />
                    )}
                />
            </Skeleton>
        </div>
    );
}
