import { IconBox, IconHistogram, IconHistory, IconSetting, IconWrench } from '@douyinfe/semi-icons';
import { Tabs, TabPane } from '@douyinfe/semi-ui';
import styles from './index.module.css'
import { AgentSetting } from './agent'
import { UsageHistory } from './history'
import { SystemSetting } from './system'
import { ToolsConfig } from './tools'
import { UsageStats } from './usage'
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
        key: 'tools',
        icon: <IconWrench />,
        title: '工具设置',
        compoment: <ToolsConfig />
    },
    {
        key: 'agent',
        icon: <IconBox />,
        title: 'agent设置',
        compoment: <AgentSetting />
    },
    {
        key: 'usage',
        icon: <IconHistogram />,
        title: '用量统计',
        compoment: <UsageStats />
    },
    {
        key: 'history',
        icon: <IconHistory />,
        title: '历史',
        compoment: <UsageHistory />
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
