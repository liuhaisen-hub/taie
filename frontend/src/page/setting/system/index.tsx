import { IconMoon, IconSun } from '@douyinfe/semi-icons'
import { List, Select, Switch } from '@douyinfe/semi-ui'
import { useState } from 'react'
import { useTheme } from '@/hooks/useTheme'
import styles from './index.module.css'

interface SettingItemProps {
    title: string
    description?: string
    control: React.ReactNode
}

const SettingItem = ({ title, description, control }: SettingItemProps) => {
    return (
        <List.Item
            main={
                <div>
                    <div className={styles.itemTitle}>{title}</div>
                    {description && <div className={styles.itemDesc}>{description}</div>}
                </div>
            }
            extra={control}
        />
    )
}

const fontOptions = [
    { value: 'small', label: '小号' },
    { value: 'default', label: '标准' },
    { value: 'large', label: '大号' },
]

const languageOptions = [
    { value: 'zh-CN', label: '简体中文' },
    { value: 'en', label: 'English' },
]

const notificationItems = [
    { key: 'responseCompletions', title: '响应完成', description: '响应补全完成时通知我' },
    { key: 'scheduledTasks', title: '计划任务', description: '计划任务执行时通知我' },
    { key: 'codeNotifications', title: '代码通知', description: '代码相关事件通知' },
    { key: 'codePermissionRequests', title: '代码权限请求', description: '需要代码授权时通知我' },
    {
        key: 'cloudSessionEmails',
        title: 'Claude Code 云会话邮件',
        description: '接收来自 Claude Code 云会话的邮件',
    },
] as const

type NotificationKey = (typeof notificationItems)[number]['key']

export const SystemSetting = () => {
    const { theme, toggleTheme } = useTheme()
    const [fontSize, setFontSize] = useState('default')
    const [language, setLanguage] = useState('zh-CN')
    const [notifications, setNotifications] = useState<Record<NotificationKey, boolean>>({
        responseCompletions: true,
        scheduledTasks: true,
        codeNotifications: false,
        codePermissionRequests: true,
        cloudSessionEmails: false,
    })

    const toggleNotification = (key: NotificationKey, checked: boolean) => {
        setNotifications(prev => ({ ...prev, [key]: checked }))
    }

    return (
        <div className={styles.root}>
            <div className={styles.section}>
                <div className={styles.sectionTitle}>基本</div>
                <List bordered>
                    <SettingItem
                        title="外观模式"
                        description="切换暗黑或明亮模式"
                        control={
                            <Switch
                                checked={theme === 'dark'}
                                onChange={toggleTheme}
                                checkedText={<IconMoon />}
                                uncheckedText={<IconSun />}
                            />
                        }
                    />
                    <SettingItem
                        title="聊天字体"
                        description="设置聊天内容的字体大小"
                        control={
                            <Select
                                value={fontSize}
                                onChange={value => setFontSize(value as string)}
                                optionList={fontOptions}
                                style={{ width: 140 }}
                            />
                        }
                    />
                    <SettingItem
                        title="语言"
                        description="选择界面显示语言"
                        control={
                            <Select
                                value={language}
                                onChange={value => setLanguage(value as string)}
                                optionList={languageOptions}
                                style={{ width: 140 }}
                            />
                        }
                    />
                </List>
            </div>
            <div className={styles.section}>
                <div className={styles.sectionTitle}>通知</div>
                <List bordered>
                    {notificationItems.map(item => (
                        <SettingItem
                            key={item.key}
                            title={item.title}
                            description={item.description}
                            control={
                                <Switch
                                    checked={notifications[item.key]}
                                    onChange={checked => toggleNotification(item.key, checked)}
                                />
                            }
                        />
                    ))}
                </List>
            </div>
        </div>
    )
}
