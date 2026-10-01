import type { SystemConfig } from '@bindings/taie/internal/po'
import { SystemServices } from '@bindings/taie/internal/services'
import { IconMoon, IconSun } from '@douyinfe/semi-icons'
import { List, Select, Switch, Toast } from '@douyinfe/semi-ui'
import { useCallback, useEffect, useState } from 'react'
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

const modeOptions = [
    { value: 'manual', label: '手动模式' },
    { value: 'plan', label: '计划模式' },
    { value: 'auto', label: '自动模式' },
]

const languageOptions = [
    { value: 'zh-CN', label: '简体中文' },
    { value: 'en', label: 'English' },
]

const notificationItems = [
    { key: 'responseCompletions', title: '响应完成', description: '响应补全完成时通知我' },
    { key: 'scheduledTasks', title: '计划任务', description: '计划任务执行时通知我' },
    { key: 'notifications', title: '通知', description: '事件通知' },
    { key: 'permissionRequests', title: '代码权限请求', description: '需要授权时通知我' },
    {
        key: 'emails',
        title: '云会话邮件',
        description: '接收来自 taie 云会话的邮件',
    },
] as const
// 权限 key 与 agent 内置工具名一一对应（read_file/write_file/edit_file/glob/grep/execute）
const permissitionItems = [
    { key: 'read_file', title: '读取文件', description: '允许agent读取文件内容' },
    { key: 'write_file', title: '写入文件', description: '允许agent写入文件内容' },
    { key: 'edit_file', title: '编辑文件', description: '允许agent编辑文件内容' },
    { key: 'glob', title: '查找文件', description: '允许agent根据glob模式查找文件' },
    { key: 'grep', title: '搜索内容', description: '允许agent在文件中搜索内容' },
    { key: 'execute', title: '执行命令', description: '允许agent执行shell命令' }
] as const
type PermissionKey = (typeof permissitionItems)[number]['key']
type NotificationKey = (typeof notificationItems)[number]['key']
type Mode = 'auto' | 'manual' | 'plan'

// 后端读取失败/字段缺失时的兜底默认值
const defaultNotifications: Record<NotificationKey, boolean> = {
    responseCompletions: true,
    scheduledTasks: true,
    notifications: false,
    permissionRequests: true,
    emails: false,
}
const defaultPermissions: Record<PermissionKey, boolean> = {
    read_file: false,
    write_file: false,
    edit_file: false,
    glob: false,
    grep: false,
    execute: false,
}

export const SystemSetting = () => {
    const { theme, toggleTheme } = useTheme()
    const [fontSize, setFontSize] = useState('default')
    const [language, setLanguage] = useState('zh-CN')
    const [notifications, setNotifications] = useState<Record<NotificationKey, boolean>>(defaultNotifications)
    const [permissions, setPermissions] = useState<Record<PermissionKey, boolean>>(defaultPermissions)
    const [mode, setMode] = useState<Mode>('manual')

    const load = useCallback(async () => {
        try {
            const cfg = await SystemServices.Get()
            if (!cfg) return
            setLanguage(cfg.language || 'zh-CN')
            setMode(cfg.mode === 'auto' || cfg.mode === 'plan' ? cfg.mode : 'manual')
            setFontSize(cfg.fontSize || 'default')
            setNotifications({ ...defaultNotifications, ...cfg.notifications })
            setPermissions({ ...defaultPermissions, ...cfg.permissions })
        } catch (err) {
            Toast.error(`加载系统设置失败：${String(err)}`)
        }
    }, [])

    useEffect(() => {
        void load()
    }, [load])

    // 即时生效型 UI：每次变更把当前快照加 patch 整体写回 system.json；
    // 保存失败只提示不回滚，下次进入设置页会以后端为准
    const saveConfig = (patch: Partial<SystemConfig>) => {
        const config: SystemConfig = { language, mode, fontSize, notifications, permissions, ...patch }
        SystemServices.Update({ config }).catch(err => Toast.error(`保存系统设置失败：${String(err)}`))
    }
    const changeLanguage = (value: string) => {
        setLanguage(value)
        saveConfig({ language: value })
    }
    const changeMode = (value: Mode) => {
        setMode(value)
        saveConfig({ mode: value })
    }
    const toggleNotification = (key: NotificationKey, checked: boolean) => {
        const next = { ...notifications, [key]: checked }
        setNotifications(next)
        saveConfig({ notifications: next })
    }
    const togglePermission = (key: PermissionKey, checked: boolean) => {
        const next = { ...permissions, [key]: checked }
        setPermissions(next)
        saveConfig({ permissions: next })
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
                        title="语言"
                        description="agent回答时使用的语言"
                        control={
                            <Select
                                value={language}
                                onChange={value => changeLanguage(value as string)}
                                optionList={languageOptions}
                                style={{ width: 140 }}
                            />
                        }
                    />
                    <SettingItem
                        title="工作模式"
                        description="自动模式下agent将减少询问你的意见"
                        control={
                            <Select
                                value={mode}
                                onChange={value => changeMode(value as Mode)}
                                optionList={modeOptions}
                                style={{ width: 140 }}
                            />
                        }
                    />
                </List>
            </div>
            <div className={styles.section}>
                <div className={styles.sectionTitle}>权限</div>
                <List bordered>
                    {permissitionItems.map(item => (
                        <SettingItem
                            key={item.key}
                            title={item.title}
                            description={item.description}
                            control={
                                <Switch
                                    checked={permissions[item.key]}
                                    onChange={checked => togglePermission(item.key, checked)}
                                />
                            }
                        />
                    ))}
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
