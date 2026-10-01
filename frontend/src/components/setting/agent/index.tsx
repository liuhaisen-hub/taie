import { IconPlus } from '@douyinfe/semi-icons'
import { Button, Form, Modal, Table, Tag, Toast } from '@douyinfe/semi-ui'
import type { FormApi } from '@douyinfe/semi-ui/lib/es/form/interface'
import type { AIModel } from '@bindings/taie/internal/po'
import { ModelServices } from '@bindings/taie/internal/services'
import { useCallback, useEffect, useState } from 'react'
import { AgentType } from '@/conf/agent'
import styles from './index.module.css'

// 表单值不含后端生成的 id/createdAt/updatedAt；enable 是布尔（Switch），提交时再转 0/1
interface AiModelFormValues {
    modelName: string
    apiKey: string
    baseUrl: string
    type: number
    provider: string
    enable: boolean
}

// 模型类型下拉/表格展示，数据源统一在 conf/agent
const typeOptions = AgentType.map(({ name, icon, value }) => ({
    value,
    label: (
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
            {icon}
            {name}
        </span>
    ),
}))

interface AiModelFormProps {
    initialValues?: Partial<AIModel>
    onSubmit: (values: AiModelFormValues) => void
    getApi: (api: FormApi<AiModelFormValues>) => void
}

const AiModelForm = ({ initialValues, onSubmit, getApi }: AiModelFormProps) => {
    const initValues: AiModelFormValues = {
        modelName: initialValues?.modelName ?? '',
        apiKey: initialValues?.apiKey ?? '',
        baseUrl: initialValues?.baseUrl ?? '',
        // 新增时默认选中第一项，保证提交时 type 一定是 number
        type: initialValues?.type ?? typeOptions[0].value,
        provider: initialValues?.provider ?? '',
        enable: initialValues?.enable === undefined ? true : initialValues.enable === 1,
    }

    return (
        <Form<AiModelFormValues>
            getFormApi={getApi}
            initValues={initValues}
            labelPosition="left"
            labelWidth={90}
            onSubmit={onSubmit}
        >
            <Form.Input
                field="modelName"
                label="模型名称"
                maxLength={50}
                rules={[{ required: true, message: '请输入名称' }]}
            />
            <Form.Input
                field="apiKey"
                label="密钥"
                maxLength={500}
                rules={[{ required: true, message: '请输入 Key' }]}
            />
            <Form.Input
                field="baseUrl"
                label="请求地址"
                maxLength={100}
                rules={[{ required: true, message: '请输入 Base URL' }]}
            />
            <Form.Select
                field="type"
                label="类型"
                placeholder="请选择类型"
                optionList={typeOptions}
                style={{ width: '100%' }}
                rules={[{ required: true, message: '请选择类型' }]}
            />
            <Form.Input
                field="provider"
                label="供应商"
                maxLength={100}
                rules={[{ required: true, message: '请输入供应商' }]}
            />
            <Form.Switch field="enable" label="状态" />
        </Form>
    )
}

export const AgentSetting = () => {
    const [data, setData] = useState<AIModel[]>([])
    const [total, setTotal] = useState(0)
    const [loading, setLoading] = useState(false)
    const [currentPage, setCurrentPage] = useState(1)
    const [pageSize, setPageSize] = useState(10)
    const [modalVisible, setModalVisible] = useState(false)
    const [editing, setEditing] = useState<AIModel>()
    const [formApi, setFormApi] = useState<FormApi<AiModelFormValues>>()

    const load = useCallback(async () => {
        setLoading(true)
        try {
            const result = await ModelServices.List({ modeName: '', page: currentPage, size: pageSize })
            setData((result?.items ?? []).filter((item): item is AIModel => item != null))
            setTotal(result?.total ?? 0)
        } catch (err) {
            Toast.error(`加载模型列表失败：${String(err)}`)
        } finally {
            setLoading(false)
        }
    }, [currentPage, pageSize])

    useEffect(() => {
        void load()
    }, [load])

    const openCreate = () => {
        setEditing(undefined)
        setModalVisible(true)
    }

    const openEdit = (record: AIModel) => {
        setEditing(record)
        setModalVisible(true)
    }

    const handleSubmit = async (values: AiModelFormValues) => {
        try {
            // 表单里 enable 是布尔（Switch），提交时转 0/1
            const payload = { ...values, enable: values.enable ? 1 : 0 }
            if (editing) {
                // editing 自带 id/createdAt/updatedAt，表单字段以提交值为准
                await ModelServices.Update({ ...editing, ...payload })
                Toast.success('修改成功')
            } else {
                // id/createdAt/updatedAt 由后端生成（GORM 自增主键 + 零值自动填充时间戳）；
                // Wails 绑定只生成 interface 没有 class，直接传普通对象。
                // 时间字段传 Go 的零值时间字符串（传空串会导致 time.Time 解析失败）
                await ModelServices.Create({
                    ...payload,
                    id: 0,
                    createdAt: '0001-01-01T00:00:00Z',
                    updatedAt: '0001-01-01T00:00:00Z',
                })
                Toast.success('新增成功')
            }
            setModalVisible(false)
            await load()
        } catch (err) {
            Toast.error(`保存失败：${String(err)}`)
        }
    }

    const columns = [
        { title: '模型名称', dataIndex: 'modelName' },
        // { title: 'apiKey', dataIndex: 'apiKey' },
        // { title: 'baseUrl', dataIndex: 'baseUrl' },
        {
            title: '类型',
            dataIndex: 'type',
            render: (type: number) => typeOptions.find(option => option.value === type)?.label ?? type,
        },
        { title: '协议', dataIndex: 'provider' },
        {
            title: '状态',
            dataIndex: 'enable',
            render: (enable: number) => (
                <Tag color={enable === 1 ? 'green' : 'red'} size="small">
                    {enable === 1 ? '启用' : '禁用'}
                </Tag>
            ),
        },
        {
            title: '操作',
            render: (_: unknown, record: AIModel) => (
                <Button type="primary" theme="borderless" size="small" onClick={() => openEdit(record)}>
                    修改
                </Button>
            ),
        },
    ]

    return (
        <div className={styles.root}>
            <div className={styles.header}>
                <h3 style={{ margin: 0 }}>AI 模型配置</h3>
                <Button theme="solid" icon={<IconPlus />} onClick={openCreate}>
                    新增
                </Button>
            </div>
            <Table<AIModel>
                className={styles.table}
                rowKey="id"
                columns={columns}
                dataSource={data}
                loading={loading}
                pagination={{
                    currentPage,
                    pageSize,
                    total,
                    showSizeChanger: true,
                    pageSizeOpts: [10, 20, 50],
                    onPageChange: setCurrentPage,
                    onPageSizeChange: size => {
                        setPageSize(size)
                        setCurrentPage(1)
                    },
                }}
            />
            <Modal
                title={editing ? '修改配置' : '新增配置'}
                visible={modalVisible}
                footer={
                    <>
                        <Button onClick={() => setModalVisible(false)}>取消</Button>
                        <Button
                            type="primary"
                            theme="solid"
                            onClick={() => formApi?.submitForm()}
                        >
                            确定
                        </Button>
                    </>
                }
                maskClosable={false}
                onCancel={() => setModalVisible(false)}
            >
                <AiModelForm initialValues={editing} onSubmit={handleSubmit} getApi={setFormApi} />
            </Modal>
        </div>
    )
}
