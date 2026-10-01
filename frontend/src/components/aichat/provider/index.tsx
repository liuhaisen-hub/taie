import { createContext, useCallback, useContext, useEffect, useRef, useState } from "react";
import { chatInputToMessage, Toast } from "@douyinfe/semi-ui";
import { ChatMessage } from "../chatmessage";
import { ChatInputer } from "../chatinput";
import styles from './index.module.css'
import { JSONStream } from "@wailsio/runtime";
import { SessionServices } from "@bindings/taie/internal/services";
import type { ChatMessage as StoredMessage } from "@bindings/taie/internal/po";
import type { Message } from "@douyinfe/semi-ui/lib/es/aiChatDialogue";
import type { MessageContent } from "@douyinfe/semi-ui/lib/es/aiChatInput";
import type { ApprovalRequest, ChatEvent, ChatRequest, ConfigureState, ContentItem } from "../types";
import { useLayout } from "@/hooks/layout";
import { deltaText, extractInputText, sameContent, storedToChat, toItems } from "./utils";

/** 挂起的审批：卡片据此判断自己是否可交互，裁决时回传 interruptId */
export interface ApprovalPending {
    interruptId: string
}

export interface AIChatContextType {
    /** 入参为 AIChatInput onMessageSend 的原始产物，内部经 chatInputToMessage 官方转换 */
    sendMessage: (props: MessageContent, extra?: ConfigureState) => void
    stop: () => void
    /** 清空本地聊天区（AIChatDialogue 重置按钮回调） */
    resetChat: () => void
    /** AIChatDialogue 受控同步：内部编辑/删除等操作回写 */
    onChatsChange: (chats?: Message[]) => void
    generate: boolean
    chat: Message[]
    sessionId: number
    /** 当前挂起的审批（null=无） */
    pendingApproval: ApprovalPending | null
    /** 审批裁决，批准/拒绝都会开 agent/approval 连接续跑 */
    decideApproval: (approved: boolean, reason?: string) => void
}

/** 生成消息唯一 id */
let seq = 0;
const nextId = (prefix: string) => `${prefix}-${Date.now()}-${seq++}`;
export const AIChatContext = createContext<AIChatContextType | undefined>(undefined)



export const AIChatProvider = () => {
    const { activeId, onSessionSelect, notifySessionsChanged } = useLayout()
    const [chat, setChats] = useState<Message[]>([])
    const [generate, setGenerating] = useState(false)
    /** 当前会话 ID（begin 事件回推）；0 表示尚未建立 */
    const [sessionId, setSessionId] = useState(0)
    /** sessionId 的同步副本：sendMessage 的 onopen 闭包要立即读到最新 id，不等一轮 setState 生效 */
    const sessionIdRef = useRef(0)
    /** 当前流连接；stop 与组件卸载时用来关闭 */
    const socketRef = useRef<ReturnType<typeof JSONStream> | null>(null);
    /** 当前流式消息 id：占位 botId，首个带服务端 id 的事件到达后替换 */
    const activeIdRef = useRef("");
    /** 服务端 id 是否已采纳（每轮只做一次，否则 rAF 按旧 id 定位会丢帧） */
    const idAdoptedRef = useRef(false);
    /** delta 缓冲区 + rAF 句柄：把高频增量合并到每一帧一次 setState */
    const bufRef = useRef("");
    const rafRef = useRef(0);
    // 挂起审批（state 供渲染 + ref 供闭包立即读到最新值）
    const [pendingApproval, setPendingApproval] = useState<ApprovalPending | null>(null);
    const pendingApprovalRef = useRef<ApprovalPending | null>(null);
    // 当前连接是审批续跑——值为本次裁决回传的 interrupt_id（done 据此定位续跑段
    // 起点），空串表示普通轮、done 整体替换
    const resumingRef = useRef("");

    /** 首个携带服务端 id 的事件到达时，把占位气泡 id 替换为服务端 id */
    const adoptServerId = useCallback((serverId?: string) => {
        if (!serverId || idAdoptedRef.current) return;
        idAdoptedRef.current = true;
        const localId = activeIdRef.current;
        activeIdRef.current = serverId;
        setChats(prev => prev.map(m => (m.id === localId ? { ...m, id: serverId } : m)));
    }, []);

    /** 把缓冲中的增量并入最后一个 message item 的内层 output_text（不存在则新建），
        Semi 只内置渲染 message 包裹形态，裸 output_text 放顶层不显示 */
    const flush = useCallback(() => {
        rafRef.current = 0;
        const chunk = bufRef.current;
        if (!chunk) return;
        bufRef.current = "";
        const id = activeIdRef.current;
        setChats(prev => prev.map(m => {
            if (m.id !== id) return m;
            const items = toItems(m.content);
            const last = items[items.length - 1];
            if (last?.type === "message") {
                const inner = Array.isArray(last.content) ? [...(last.content as ContentItem[])] : [];
                const lastInner = inner[inner.length - 1];
                if (lastInner?.type === "output_text") {
                    inner[inner.length - 1] = { ...lastInner, text: (lastInner.text ?? "") + chunk };
                } else {
                    inner.push({ type: "output_text", text: chunk });
                }
                items[items.length - 1] = { ...last, content: inner };
            } else {
                // 上一项是工具卡片或还没有内容：新开一个 message item
                items.push({ type: "message", content: [{ type: "output_text", text: chunk }] });
            }
            return { ...m, content: items, status: "in_progress" };
        }));
    }, []);

    const scheduleFlush = useCallback(() => {
        if (rafRef.current) return; // 已有排程，等下一帧
        rafRef.current = requestAnimationFrame(flush);
    }, [flush]);

    /** 真正收口：清缓冲、关连接、清挂起审批、复位状态；mark 存在时给本轮消息盖上收口状态 */
    const finish = useCallback((mark?: string) => {
        if (rafRef.current) {
            cancelAnimationFrame(rafRef.current);
            rafRef.current = 0;
        }
        bufRef.current = "";
        socketRef.current?.close();
        socketRef.current = null;
        // 任何收口都终结挂起审批（停止/断连/发新消息 = 放弃裁决）
        pendingApprovalRef.current = null;
        setPendingApproval(null);
        if (mark && activeIdRef.current) {
            const id = activeIdRef.current;
            setChats(prev => prev.map(m => (m.id === id ? { ...m, status: mark } : m)));
        }
        setGenerating(false);
    }, []);

    /** 审批暂停——关连接、generate=false，但不动气泡状态、不清挂起审批
        （回合没结束，只是挂起等裁决） */
    const pauseForApproval = useCallback(() => {
        if (rafRef.current) {
            cancelAnimationFrame(rafRef.current);
            rafRef.current = 0;
        }
        bufRef.current = "";
        socketRef.current?.close();
        socketRef.current = null;
        setGenerating(false);
    }, []);

    /** 停止生成：关闭连接 → Go 侧 ctx 取消 → 模型推理中断，气泡置 cancelled */
    const stop = useCallback(() => {
        finish("cancelled");
    }, [finish]);

    /** 事件处理：agent/chat 与 agent/approval 两个连接共用同一套消费逻辑 */
    const handleEvent = useCallback((evt: ChatEvent) => {
        const msg = evt.message ?? ({} as ChatEvent["message"]);
        switch (evt.kind) {
            case "begin":
                // 新会话此时才拿到 ID，后续请求必须回传；同步回布局层作为当前会话，
                // 否则布局层仍是"新建会话"态，再次点新建不会触发清空
                sessionIdRef.current = Number(msg.id ?? 0);
                setSessionId(sessionIdRef.current);
                onSessionSelect(sessionIdRef.current);
                break;
            case "delta":
                adoptServerId(msg.id);
                bufRef.current += deltaText(msg.content);
                scheduleFlush();
                break;
            case "tool_call":
                // 先刷掉残留 delta，保证卡片顺序与事件顺序一致；function_call 由 Semi 原生渲染
                adoptServerId(msg.id);
                flush();
                setChats(prev => prev.map(m => m.id === activeIdRef.current
                    ? { ...m, content: [...toItems(m.content), ...(msg.content ?? [])] }
                    : m));
                break;
            case "tool_result":
                // 按 call_id 把对应 function_call 置 completed——审批续跑时命中
                // 的是中断前那张卡片（单气泡策略的关键收益）
                adoptServerId(msg.id);
                setChats(prev => prev.map(m => {
                    if (m.id !== activeIdRef.current) return m;
                    const items = toItems(m.content);
                    for (const inc of msg.content ?? []) {
                        for (let i = 0; i < items.length; i++) {
                            if (items[i].type === "function_call" && items[i].call_id === inc.call_id) {
                                items[i] = { ...items[i], status: "completed" };
                            }
                        }
                        items.push({ ...inc });
                    }
                    return { ...m, content: items };
                }));
                break;
            case "approval": {
                // 工具等待审批——追加审批卡片，记录挂起态后暂停（不是结束）
                adoptServerId(msg.id);
                flush();
                setChats(prev => prev.map(m => m.id === activeIdRef.current
                    ? { ...m, content: [...toItems(m.content), ...(msg.content ?? [])] }
                    : m));
                const item = (msg.content ?? [])[0] as ContentItem | undefined;
                const interruptId =
                    (item?.extra as { interrupt_id?: string } | undefined)?.interrupt_id ?? "";
                if (interruptId) {
                    pendingApprovalRef.current = { interruptId };
                    setPendingApproval({ interruptId });
                }
                pauseForApproval();
                break;
            }
            case "done": {
                // 服务端全文兜底 + 状态收口；incomplete = 后端中断收口，气泡置 cancelled。
                if (rafRef.current) {
                    cancelAnimationFrame(rafRef.current);
                    rafRef.current = 0;
                }
                bufRef.current = "";
                adoptServerId(msg.id);
                const status = msg.status === "incomplete" ? "cancelled" : "completed";
                if (resumingRef.current) {
                    // 审批续跑轮：续跑段事件已实时归并进气泡，直接追加会和 done 重组的
                    // 同段内容整体重复（正文最明显：答案出现两遍）；整体替换又会抹掉
                    // 审批前的正文与卡片——因此只替换挂起审批卡之后的段落。替换源与
                    // 服务端落库同源，顺带修复流式尾部未及 rAF 刷帧被丢弃造成的截断。
                    const interruptId = resumingRef.current;
                    resumingRef.current = "";
                    setChats(prev => prev.map(m => {
                        if (m.id !== activeIdRef.current) return m;
                        const items = toItems(m.content);
                        let base = -1;
                        for (let i = items.length - 1; i >= 0; i--) {
                            const it = items[i];
                            if (it.type === "approval" &&
                                (it.extra as { interrupt_id?: string } | undefined)?.interrupt_id === interruptId) {
                                base = i + 1;
                                break;
                            }
                        }
                        if (base < 0) {
                            // 兜底：找不到挂起审批卡（气泡被外力重载等），退回追加去重
                            const seen = new Set(
                                items.filter(i => i.call_id).map(i => `${i.type}:${i.call_id}`));
                            const fresh = (msg.content ?? [])
                                .filter(i => !(i.call_id && seen.has(`${i.type}:${i.call_id}`)));
                            return { ...m, content: [...items, ...fresh], status };
                        }
                        return { ...m, content: [...items.slice(0, base), ...(msg.content ?? [])], status };
                    }));
                } else {
                    // 原逻辑：服务端全文兜底（整体替换消除流式归并漂移）+ 状态收口。
                    // 重组结果与流式归并等价时保留原 content 引用（sameContent）：
                    // MarkdownRender 的 raw 不变即不触发整树重编译，消除收口闪动
                    setChats(prev => prev.map(m => {
                        if (m.id !== activeIdRef.current) return m;
                        const items = toItems(m.content);
                        return sameContent(items, msg.content)
                            ? { ...m, status }
                            : { ...m, content: msg.content ?? items, status };
                    }));
                }
                finish();
                // 本轮已落库，通知侧边栏重拉列表（新会话露脸 / 标题更新）
                notifySessionsChanged();
                break;
            }
            case "error": {
                adoptServerId(msg.id);
                const errorText = deltaText(msg.content) || "未知错误";
                setChats(prev => prev.map(m => m.id === activeIdRef.current
                    ? {
                        ...m,
                        // 错误文案同样按 message 包裹形态追加，Semi 才会渲染
                        content: [...toItems(m.content), {
                            type: "message",
                            content: [{ type: "output_text", text: `出错了：${errorText}` }],
                        }],
                        status: "failed",
                    }
                    : m));
                finish();
                break;
            }
        }
    }, [adoptServerId, flush, scheduleFlush, pauseForApproval, finish, onSessionSelect, notifySessionsChanged]);

    /** 发送一条用户消息，建立流式连接并增量接收回答 */
    const sendMessage = useCallback((props: MessageContent, extra?: ConfigureState) => {
        if (generate) return;
        // 官方转换：attachments → input_image/input_file item（本地预览回显），文本 → input_text
        const userMsg = chatInputToMessage(props);
        const text = extractInputText(userMsg);
        if (!text.trim()) return;

        const userId = nextId("u");
        const botId = nextId("a");
        activeIdRef.current = botId;
        idAdoptedRef.current = false;
        resumingRef.current = ""; // 普通轮 done 走整体替换模式

        // 先插入用户消息 + 空的 assistant 占位消息（content 为 item 数组，等事件归并）
        setChats(prev => [
            ...prev,
            { id: userId, ...userMsg, createdAt: Date.now() },
            { id: botId, role: "assistant", content: [], status: "in_progress", createdAt: Date.now() },
        ]);
        setGenerating(true);

        // 一次问答 = 一个连接，生命周期清晰、取消天然生效
        const socket = JSONStream("agent/chat");
        socketRef.current = socket;

        socket.onopen = () => {
            socket.send({ session_id: sessionIdRef.current, user_input: text, extra } satisfies ChatRequest);
        };

        socket.onmessage = (ev: MessageEvent) => handleEvent(ev.data as ChatEvent);

        // 连接异常断开（非 done/error 路径）也要收口
        socket.onclose = () => {
            if (socketRef.current === socket) finish("cancelled");
        };
    }, [generate, handleEvent, finish]);

    /** 审批裁决。批准/拒绝都开 agent/approval 连接续跑；
        事件继续归并进当前气泡（activeIdRef 不变、idAdoptedRef 置 true 不再采纳新 id） */
    const decideApproval = useCallback((approved: boolean, reason?: string) => {
        const pending = pendingApprovalRef.current;
        if (!pending || generate) return;
        pendingApprovalRef.current = null;
        setPendingApproval(null);

        // 卡片落定：批准 completed / 拒绝 failed，按钮随之消失
        const interruptId = pending.interruptId;
        setChats(prev => prev.map(m => {
            if (m.id !== activeIdRef.current) return m;
            const items = toItems(m.content).map(i =>
                i.type === "approval" && (i.extra as { interrupt_id?: string } | undefined)?.interrupt_id === interruptId
                    ? { ...i, status: approved ? "completed" : "failed" }
                    : i);
            return { ...m, content: items };
        }));

        resumingRef.current = interruptId;
        setGenerating(true);
        const socket = JSONStream("agent/approval");
        socketRef.current = socket;
        socket.onopen = () => {
            socket.send({
                session_id: sessionIdRef.current,
                interrupt_id: interruptId,
                approved,
                reason: reason ?? "",
            } satisfies ApprovalRequest);
        };
        socket.onmessage = (ev: MessageEvent) => handleEvent(ev.data as ChatEvent);
        socket.onclose = () => {
            // done/error 已 finish 收口时 socketRef 已置 null；这里兜底异常断开
            if (socketRef.current === socket) {
                resumingRef.current = "";
                finish("cancelled");
            }
        };
    }, [generate, handleEvent, finish]);

    const resetChat = useCallback(() => setChats([]), []);
    const onChatsChange = useCallback((chats?: Message[]) => setChats(chats ?? []), []);

    /** 侧边栏点击会话：拉取落库消息整体替换聊天区，并续接 sessionId，
        后续 sendMessage 继续写入该会话 */
    const loadSession = useCallback(async (id: number) => {
        if (!id) return;
        if (generate) stop(); // 切会话前中断进行中的生成，避免旧流回写新会话视图
        try {
            const rows = (await SessionServices.GetMessage({ id })) ?? [];
            setChats(rows.filter((r): r is StoredMessage => r != null).map(storedToChat));
            sessionIdRef.current = id;
            setSessionId(id);
        } catch (err) {
            Toast.error(`加载会话记录失败：${String(err)}`);
        }
    }, [generate, stop]);

    /** 响应布局层的会话状态：activeId 指向某会话时加载其落库记录；
        activeId 为空（新建会话）时中断生成、清空聊天区并复位 sessionId，
        下一条消息由后端创建新会话（begin 事件回推新 ID 后同步回布局层）。
        注意"无会话"有两种写法：activeId=undefined 与 sessionIdRef=0，
        必须归一化比较，否则新建会话态下每次 generate 翻转都会误入清空路径，
        把刚发出的消息气泡和连接一起掐掉。 */
    useEffect(() => {
        if (activeId == null) {
            if (sessionIdRef.current === 0) return; // 已在新建会话态（含发送触发的重跑），勿动
            if (generate) stop();
            setChats([]);
            sessionIdRef.current = 0;
            setSessionId(0);
            return;
        }
        if (activeId === sessionIdRef.current) return; // 已是目标会话，无需加载
        void loadSession(activeId);
    }, [activeId, generate, stop, loadSession]);

    const ctxValue: AIChatContextType = {
        generate,
        chat,
        sessionId,
        sendMessage,
        stop,
        resetChat,
        onChatsChange,
        pendingApproval,
        decideApproval,
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
