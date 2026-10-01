import type { CommonJson } from '@bindings/taie/internal/po'
import { Form } from '@douyinfe/semi-ui'
import type { FormApi } from '@douyinfe/semi-ui/lib/es/form/interface'

export interface ConfigArgsFormProps {
    /** 配置项数组：label 作为表单字段名（key），name 作为展示名，value 作为初始值 */
    args: CommonJson[]
    getApi: (api: FormApi<Record<string, string>>) => void
}

/**
 * 全局公共组件：按 CommonJson（{label, name, value}）数组动态渲染配置表单。
 * 工具参数存库为该数组的 JSON，不同工具字段不同，新增工具无需改动此组件。
 */
export const ConfigArgsForm = ({ args, getApi }: ConfigArgsFormProps) => {
    const initValues: Record<string, string> = {}
    for (const item of args) {
        initValues[item.label] = item.value
    }

    return (
        <Form<Record<string, string>>
            getFormApi={getApi}
            initValues={initValues}
            labelPosition="left"
            labelWidth={90}
        >
            {args.map(item => (
                <Form.Input
                    key={item.label}
                    field={item.label}
                    label={item.name || item.label}
                    maxLength={500}
                />
            ))}
        </Form>
    )
}
