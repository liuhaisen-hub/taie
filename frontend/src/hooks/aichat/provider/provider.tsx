import { createContext } from "react";
import { ChatMessage } from "../chatmessage";
import { ChatInputer } from "../chatInput";
import styles from './index.module.css'
import type { MessageContent } from "@douyinfe/semi-ui/lib/es/aiChatInput";
import { JSONStream } from "@wailsio/runtime";
import type { Message } from "@douyinfe/semi-ui/lib/es/aiChatDialogue";
interface ChatEvent {
    type: "delta" | "tool" | "done" | "error";
    delta?: string;
    tool?: string;
    tool_status?: string;
    content?: string;
    error?: string;
}
interface ChatRequest {
    user_input: string;
}
export interface AIChatContextType {
    sendMessage: (text: string) => void
    generate: boolean
    chat: Message[]
}

/** 生成消息唯一 id */
let seq = 0;
const nextId = (prefix: string) => `${prefix}-${Date.now()}-${seq++}`;
export const AIChatContext = createContext<AIChatContextType | undefined>(undefined)


export function AIChatProvider() {
    const [chat, setChats] = useState<Message[]>([])
    const [generate, setGenerating] = useState(false)
    /** 当前流连接；stop 与组件卸载时用来关闭 */
    const socketRef = useRef<ReturnType<typeof JSONStream> | null>(null);
    /** delta 缓冲区 + rAF 句柄：把高频增量合并到每一帧一次 setState */
    const bufRef = useRef("");
    const rafRef = useRef(0);
    /** 把缓冲中的增量追加到指定消息，status 切为 in_progress */
    const flush = useCallback((id: string) => {
        rafRef.current = 0;
        const chunk = bufRef.current;
        if (!chunk) return;
        bufRef.current = "";
        setChats(prev => prev.map(m =>
            m.id === id
                ? { ...m, content: (typeof m.content === "string" ? m.content : "") + chunk, status: "in_progress" }
                : m,
        ));
    }, []);
    /** 结束一次生成：清缓冲、关连接、复位状态 */
    const finish = useCallback((id?: string) => {
        if (rafRef.current) {
            cancelAnimationFrame(rafRef.current);
            rafRef.current = 0;
        }
        if (id) flush(id); // 收尾把残留增量刷出去
        bufRef.current = "";
        socketRef.current?.close();
        socketRef.current = null;
        setGenerating(false);

    }, [flush]);

    /** 停止生成：关闭连接 → Go 侧 ctx 取消 → 模型推理中断 */
    const stop = useCallback(() => {
        finish();
    }, [finish]);
    const scheduleFlush = useCallback((id: string) => {
        if (rafRef.current) return; // 已有排程，等下一帧
        rafRef.current = requestAnimationFrame(() => flush(id));
    }, [flush]);
    /** 发送一条用户消息，建立流式连接并增量接收回答 */
    const sendMessage = useCallback((text: string) => {
        const input = text.trim();
        if (!input || generate) return;

        const userId = nextId("u");
        const botId = nextId("a");

        // 先插入用户消息 + 空的 assistant 占位消息
        setChats(prev => [
            ...prev,
            { id: userId, role: "user", content: input, createAt: Date.now() },
            { id: botId, role: "assistant", content: "", status: "loading", createAt: Date.now() },
        ]);
        setGenerating(true);

        // 一次问答 = 一个连接，生命周期清晰、取消天然生效
        const socket = JSONStream("agent/chat");
        socketRef.current = socket;

        socket.onopen = () => {
            socket.send({ user_input: input } satisfies ChatRequest);
        };

        socket.onmessage = (ev: MessageEvent) => {
            const evt = ev.data as ChatEvent;
            switch (evt.type) {
                case "delta":
                    bufRef.current += evt.delta ?? "";
                    scheduleFlush(botId);
                    break;
                case "tool":
                    break;
                case "done":
                    // 用服务端全文兜底覆盖，避免丢帧导致的文本缺失
                    setChats(prev => prev.map(m =>
                        m.id === botId
                            ? { ...m, content: evt.content ?? m.content, status: "completed" }
                            : m,
                    ));
                    finish();
                    break;
                case "error":
                    setChats(prev => prev.map(m =>
                        m.id === botId
                            ? { ...m, content: `出错了：${evt.error ?? "未知错误"}`, status: "completed" }
                            : m,
                    ));
                    finish();
                    break;
            }
        };

        // 连接异常断开（非 done/error 路径）也要收尾
        socket.onclose = () => {
            if (socketRef.current === socket) finish(botId);
        };
    }, [generate, finish, scheduleFlush]);
    const ctxValue: AIChatContextType = {
        generate,
        chat,
        sendMessage

    }
    return <AIChatContext.Provider value={ctxValue}>
        <div className={styles.root}>
            <div className={styles.top}>
                <ChatMessage />
            </div>
            <div className={styles.bottom}>
                <ChatInputer />
            </div>
        </div>
    </AIChatContext.Provider>
}

export function useAgentChat() {
    const ctx = useContext(AIChatContext)
    if (!ctx) {
        throw new Error("useAgent 需要在AIChatProvider中使用")
    }
    return ctx
}