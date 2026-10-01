import type { CommonJson, ToolConfig } from '@bindings/taie/internal/po'
import { ToolsServices } from '@bindings/taie/internal/services'
import { IconEdit, IconHelpCircle } from '@douyinfe/semi-icons'
import { Button, List, Modal, Switch, Toast, Tooltip } from '@douyinfe/semi-ui'
import type { FormApi } from '@douyinfe/semi-ui/lib/es/form/interface'
import { useCallback, useEffect, useState } from 'react'
import { ConfigArgsForm } from '@/components/configArgsForm'
import styles from './index.module.css'

// 工具为内置固定项：不能增删，只能启用/禁用与编辑配置参数
export const ToolsConfig = () => {
    const [tools, setTools] = useState<ToolConfig[]>([])
    const [loading, setLoading] = useState(false)
    const [editing, setEditing] = useState<ToolConfig>()
    /** 弹窗中的配置项（editing.args 的解析结果），Label 决定表单字段 */
    const [editArgs, setEditArgs] = useState<CommonJson[]>([])
    const [visible, setVisible] = useState(false)
    const [formApi, setFormApi] = useState<FormApi<Record<string, string>>>()

    const load = useCallback(async () => {
        setLoading(true)
        try {
            // 内置工具数量固定，一次拉全量
            const result = await ToolsServices.List({ page: 1, size: 100 })
            setTools((result?.items ?? []).filter((item): item is ToolConfig => item != null))
        } catch (err) {
            Toast.error(`加载工具列表失败：${String(err)}`)
        } finally {
            setLoading(false)
        }
    }, [])

    useEffect(() => {
        void load()
    }, [load])

    const handleEnable = async (tool: ToolConfig, enable: boolean) => {
        try {
            await ToolsServices.EnableTools({ toolName: tool.toolName, enable })
            await load()
        } catch (err) {
            // 后端会拦截参数未配置就启用的工具（如未填 apiKey）
            Toast.error(`「${tool.toolName}」${enable ? '启用' : '禁用'}失败：${String(err)}`)
        }
    }

    const openEdit = (tool: ToolConfig) => {
        setEditArgs(tool.args ?? [])
        setEditing(tool)
        setVisible(true)
    }

    const handleOk = async () => {
        if (!editing) return
        try {
            const values = formApi?.getValues() ?? {}
            // 表单以 label 为字段名，提交时还原成 CommonJson 数组（name 展示名原样保留）
            const args = editArgs.map(item => ({ label: item.label, name: item.name, value: values[item.label] ?? '' }))
            await ToolsServices.Update({ toolName: editing.toolName, args })
            Toast.success('保存成功')
            setVisible(false)
            await load()
        } catch (err) {
            Toast.error(`保存失败：${String(err)}`)
        }
    }

    return (
        <div className={styles.root}>
            <div className={styles.section}>
                <div className={styles.sectionTitle}>工具</div>
                <List
                    bordered
                    loading={loading}
                    dataSource={tools}
                    emptyContent="暂无工具"
                    renderItem={tool => (
                        <List.Item
                            main={
                                <div>
                                    <div className={styles.itemTitle}>{tool.toolName}</div>
                                    {tool.desc && <div className={styles.itemDesc}>{tool.desc}</div>}
                                </div>
                            }
                            extra={
                                <div className={styles.itemControl}>
                                    <Tooltip content={tool.info}>
                                        <IconHelpCircle className={styles.itemHelp} />
                                    </Tooltip>
                                    <Button
                                        theme="borderless"
                                        icon={<IconEdit />}
                                        onClick={() => openEdit(tool)}
                                    >
                                        编辑
                                    </Button>
                                    <Switch
                                        checked={tool.enable}
                                        onChange={checked => handleEnable(tool, checked)}
                                    />
                                </div>
                            }
                        />
                    )}
                />
            </div>
            <Modal
                title={`编辑配置：${editing?.toolName ?? ''}`}
                visible={visible}
                footer={
                    <>
                        <Button onClick={() => setVisible(false)}>取消</Button>
                        <Button type="primary" theme="solid" onClick={handleOk}>
                            确定
                        </Button>
                    </>
                }
                maskClosable={false}
                onCancel={() => setVisible(false)}
            >
                <ConfigArgsForm args={editArgs} getApi={setFormApi} />
            </Modal>
        </div>
    )
}
