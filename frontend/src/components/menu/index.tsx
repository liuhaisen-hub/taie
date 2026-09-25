import { IconSemiLogo } from "@douyinfe/semi-icons"
import { Modal, Nav } from "@douyinfe/semi-ui"
import styles from './index.module.css'
import { MenuItems } from "@/conf/menu"
import type { ItemKey } from "@douyinfe/semi-ui/lib/es/navigation"
import { useState } from "react"
import { SettingPage } from "@/page/setting"
export const Menu: React.FC = () => {
    const [setting, setSetting] = useState(false)
    const onItemClick = (data: {
        itemKey?: ItemKey | undefined;
        domEvent?: MouseEvent;
        isOpen?: boolean;
    }) => {
        if (data.itemKey == 'Setting') {
            setSetting(true)
        }
    }
    return <>
        <Nav
            defaultSelectedKeys={['Home']}
            className={styles.menu}
            style={{ maxWidth: 220, }}
            items={[
                ...MenuItems
            ]}
            onClick={onItemClick}
            header={{
                logo: <IconSemiLogo style={{ fontSize: 36 }} />,
                text: 'Semi Design',
            }}
            footer={{
                collapseButton: true,
            }}
        />
        <Modal
            title={null}
            visible={setting}
            width={800}
            footer={null}
            centered
            // Cap the modal within the viewport so the whole box never
            // overflows; overflow scrolls inside the body instead (the
            // default scrolls the entire modal via .semi-modal-wrap).
            bodyStyle={{ maxHeight: 'calc(100vh - 120px)', overflowY: 'auto' }}
            onCancel={() => setSetting(false)}
        >
            <SettingPage />
        </Modal>
    </>
}