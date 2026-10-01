import { Avatar, Modal } from "@douyinfe/semi-ui"
import { IconSetting } from "@douyinfe/semi-icons"
import styles from './index.module.css'
import { MenuList } from "./meun"
import { SessionList } from "./sessionlist"
import { SettingPage } from "@/components/setting"
import { useLayout } from "@/hooks/layout"

export const SiderBard = () => {
    const [setting, setSetting] = useState(false)
    const { createNewSession } = useLayout()
    return <div className={styles.sider_bar}>
        <MenuList onAction={key => {
            if (key === 'new_session') createNewSession()
        }} />
        <SessionList />
        <div className={styles.bottom} onClick={() => setSetting(true)}>
            <Avatar size="small" shape="square" src="/logo.png"></Avatar>
            <IconSetting size="large" />
        </div>
        <Modal
            title={null}
            visible={setting}
            width="calc((100vw - 120px) / var(--app-zoom, 1))"
            height="calc((100vh - 120px) / var(--app-zoom, 1))"
            footer={null}
            centered
            // Semi 默认 .semi-modal-body-wrapper 不撑满（无 flex:1）、body 高度为
            // 内容高，内容超出时会被 content 的 overflow:hidden 直接裁掉。
            // 这里让 wrapper 吃满剩余高度、body 定高，设置页内部的 height:100%
            // 链路才能成立，滚动发生在各页自己的容器里。
            modalContentClass={styles.setting_modal_content}
            bodyStyle={{ height: '100%', overflowY: 'auto' }}
            onCancel={() => setSetting(false)}
        >
            <SettingPage />
        </Modal>
    </div>
}
