import { IconHistogram, IconHistory, IconSetting, IconSync } from '@douyinfe/semi-icons';
import { Tabs, TabPane } from '@douyinfe/semi-ui';
import styles from './index.module.css'
import { AgentSetting } from './agent'
import { SystemSetting } from './system'
export const settingMenu = [
    {
        key: 'system',
        icon: <IconSetting />,
        title: '系统设置',
        compoment: <SystemSetting />
    },
    // {
    //     key: 'user',
    //     icon: <IconUser />,
    //     title: '用户设置',

    // },
    {
        key: 'agent',
        icon: <IconSync />,
        title: 'agent设置',
        compoment: <AgentSetting />
    },
    {
        key: 'usage',
        icon: <IconHistogram />,
        title: '用量统计',
    },
    {
        key: 'history',
        icon: <IconHistory />,
        title: '历史'
    }
]
export const SettingPage = () => {
    return <div className={styles.root}>
        <Tabs className={styles.tabs} tabPosition="left" tabPaneMotion defaultActiveKey='system'>
            {settingMenu.map(item => {
                return <TabPane key={item.key}
                    itemKey={item.key}
                    tab={
                        <span>
                            {item.icon}
                            {item.title}
                        </span>
                    }
                >
                    {item.compoment}
                </TabPane>
            })}
        </Tabs>
    </div>
}
