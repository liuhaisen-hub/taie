/* 临时复现页：无 Wails 后端跑设置弹窗里的用量统计。
   Wails 浏览器模式走 POST /wails/runtime，这里用 fetch 拦截按 methodID
   返回 mock 数据（ES Module 命名空间冻结，不能直接改 TokenUseServices），
   渲染与 siderbar 完全相同参数的 Modal + SettingPage。 */
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { Modal } from '@douyinfe/semi-ui'
import { useState } from 'react'
import './index.css'
import '@douyinfe/semi-ui/react19-adapter'
import { initScreenScale } from './conf/scale'
import { SettingPage } from './components/setting'
import type { TokenDailyStat, TokenModelStat, TokenSessionStat } from '@bindings/taie/internal/services'
import styles from './components/siderbar/index.module.css'

// tokenuseservices.ts 生成的方法 ID
const BINDING_IDS = {
    DailyUsage: 3438064097,
    ModelUsage: 2606885645,
    SessionUsage: 4259221598,
} as const

const SESSIONS: TokenSessionStat[] = Array.from({ length: 87 }, (_, i) => ({
    sessionId: i + 1,
    title: `调试会话 ${i + 1}`,
    promptTokens: 1000 + i * 37,
    completionTokens: 500 + i * 19,
    totalTokens: 1500 + i * 56,
}))

const MODELS: TokenModelStat[] = Array.from({ length: 23 }, (_, i) => ({
    modelName: `doubao-model-${i + 1}-pro`,
    promptTokens: 2000 + i * 111,
    completionTokens: 800 + i * 43,
    totalTokens: 2800 + i * 154,
}))

const DAILY: TokenDailyStat[] = Array.from({ length: 5 }, (_, i) => {
    const d = new Date(2026, 8, 26 + i)
    const promptTokens = 20000 + i * 1500
    const completionTokens = 9000 + i * 700
    return {
        date: `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`,
        promptTokens,
        completionTokens,
        totalTokens: promptTokens + completionTokens,
    }
})

const mockData = new Map<Number, unknown>([
    [BINDING_IDS.DailyUsage, DAILY],
    [BINDING_IDS.SessionUsage, SESSIONS],
    [BINDING_IDS.ModelUsage, MODELS],
])

const nativeFetch = window.fetch.bind(window)
window.fetch = async (input, init) => {
    const url = typeof input === 'string' ? input : input instanceof URL ? input.href : input.url
    console.log('[mock] fetch', init?.method, url, typeof init?.body)
    if (url.includes('/wails/runtime') && init?.method === 'POST' && typeof init.body === 'string') {
        try {
            const body = JSON.parse(init.body)
            const methodID = body?.args?.methodID
            console.log('[mock] methodID =', methodID, 'hit =', mockData.has(methodID))
            if (mockData.has(methodID)) {
                return new Response(JSON.stringify(mockData.get(methodID)), {
                    status: 200,
                    headers: { 'Content-Type': 'application/json' },
                })
            }
        } catch (e) { console.log('[mock] parse fail', e) }
    }
    return nativeFetch(input, init)
}

initScreenScale()

const ReproApp = () => {
    const [setting] = useState(true)
    return (
        <div className={styles.sider_bar}>
            <Modal
                title={null}
                visible={setting}
                width="calc((100vw - 120px) / var(--app-zoom, 1))"
                height="calc((100vh - 120px) / var(--app-zoom, 1))"
                footer={null}
                centered
                modalContentClass={styles.setting_modal_content}
                bodyStyle={{ height: '100%', overflowY: 'auto' }}
                onCancel={() => undefined}
            >
                <SettingPage />
            </Modal>
        </div>
    )
}

createRoot(document.getElementById('root')!).render(
    <StrictMode>
        <ReproApp />
    </StrictMode>,
)
