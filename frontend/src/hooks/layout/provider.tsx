import { createContext, useCallback, useContext, useMemo, useState } from "react";

interface LayoutCtxValue {
    /** 当前会话 ID；undefined 表示新会话（尚未产生后端 ID） */
    activeId: number | undefined
    /** 选中会话：写入 activeId，聊天区由 AIChatProvider 监听并加载落库记录 */
    onSessionSelect: (sessionId: number) => void
    /** 新建会话：清空选中态，聊天区复位；下一条消息发送时后端才创建会话 */
    createNewSession: () => void
    /** 会话列表版本号：会话落库/变更时 +1，SessionList 依赖它重拉列表 */
    sessionsVersion: number
    /** 通知侧边栏会话列表已变更（新会话落库、标题更新等） */
    notifySessionsChanged: () => void
}

export const layoutContext = createContext<LayoutCtxValue | undefined>(undefined)

// App.tsx 把 SiderBard 和路由 Outlet 都包在本 Provider 内，两侧都能用 useLayout 通信
export function LayoutProvider({ children }: { children: React.ReactNode }) {
    const [activeId, setActiveId] = useState<number | undefined>(undefined)
    const [sessionsVersion, setSessionsVersion] = useState(0)
    const onSessionSelect = useCallback((sessionId: number) => {
        setActiveId(sessionId)
    }, [])
    const createNewSession = useCallback(() => {
        setActiveId(undefined)
    }, [])
    const notifySessionsChanged = useCallback(() => {
        setSessionsVersion(v => v + 1)
    }, [])
    const ctxVal = useMemo(() => ({
        activeId,
        onSessionSelect,
        createNewSession,
        sessionsVersion,
        notifySessionsChanged
    }), [activeId, onSessionSelect, createNewSession, sessionsVersion, notifySessionsChanged])
    return <layoutContext.Provider value={ctxVal}>
        {children}
    </layoutContext.Provider>
}

export function useLayout() {
    const ctx = useContext(layoutContext)
    if (!ctx) {
        throw Error("useLayout 必须在 LayoutProvider中使用")
    }
    return ctx
}
