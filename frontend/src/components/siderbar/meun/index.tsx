import { IconList } from "@douyinfe/semi-icons"
import { Button } from "@douyinfe/semi-ui"
import styles from './index.module.css'
import type { ReactNode } from "react"
import { IconChat } from "@douyinfe/semi-icons-lab"

export type MenuAction = {
    key: string
    icon: ReactNode
    text: string
}

// 顶部功能按钮，后续在这里追加即可
const actions: MenuAction[] = [
    { key: 'new_session', icon: <IconChat />, text: '新建会话' },
    { key: 'new_project', icon: <IconList />, text: '项目' },
]

export const MenuList = ({ onAction }: { onAction?: (key: string) => void }) => {
    return <div className={styles.root}>
        {actions.map(item => (
            <Button
                key={item.key}
                theme="borderless"
                type="tertiary"
                icon={item.icon}
                className={styles.action}
                onClick={() => onAction?.(item.key)}
            >
                {item.text}
            </Button>
        ))}
    </div>
}
