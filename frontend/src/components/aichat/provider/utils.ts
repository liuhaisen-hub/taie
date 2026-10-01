import type { Message } from "@douyinfe/semi-ui/lib/es/aiChatDialogue";
import type { ContentItem } from "../types";
import type { ChatMessage as StoredMessage } from "@bindings/taie/internal/po";
/** content 统一为 ContentItem[]（兼容 string 形态） */
export const toItems = (content: Message["content"]): ContentItem[] => {
    if (Array.isArray(content)) return content as ContentItem[];
    if (typeof content === "string" && content) {
        return [{ type: "message", content: [{ type: "output_text", text: content }] }];
    }
    return [];
};

/** 取 delta 事件 content 里的增量文本（output_text items） */
export const deltaText = (content?: ContentItem[]): string =>
    (content ?? []).filter(i => i.type === "output_text").map(i => i.text ?? "").join("");

/** 判断服务端重组的内容序列与前端流式归并结果是否等价。等价时 done 保留原
    content 引用：MarkdownRender 的 raw 不变即不重编译，避免收口时刻整树
    unmount/remount 造成的闪动。done 未携带 content（服务端空序列 omitempty）
    且本地也为空时视为等价。 */
export const sameContent = (a: ContentItem[], b?: ContentItem[]): boolean => {
    if (!Array.isArray(b)) return a.length === 0;
    if (a.length !== b.length) return false;
    return a.every((item, i) => {
        const other = b[i];
        if (item.type !== other?.type) return false;
        if (item.type === "message") {
            return sameContent((item.content as ContentItem[]) ?? [], other.content as ContentItem[]);
        }
        return item.text === other.text
            && item.status === other.status
            && item.call_id === other.call_id
            && item.name === other.name
            && item.result === other.result;
    });
};

/** 从 chatInputToMessage 的产物里提取纯文本（嵌套 message item 下的 input_text） */
export const extractInputText = (msg: { content?: unknown }): string => {
    const content = msg.content;
    if (typeof content === "string") return content;
    for (const item of Array.isArray(content) ? content as ContentItem[] : []) {
        if (item.type === "message") {
            const inner = (item.content as ContentItem[] | undefined) ?? [];
            return inner.filter(c => c.type === "input_text").map(c => c.text ?? "").join("");
        }
    }
    return "";
};

/** []byte 经 wails 序列化成 base64 字符串，这里解回 UTF-8 文本 */
export const decodeBase64 = (b64: string): string =>
    new TextDecoder().decode(Uint8Array.from(atob(b64), c => c.charCodeAt(0)));

/** 落库行 Extra（{"status":N} JSON）→ Semi 气泡状态；0=完成 1=中断 3=错误（po 常量） */
export const storedStatus = (extra: string | null | undefined): Message["status"] => {
    if (!extra) return "completed";
    try {
        const status = (JSON.parse(decodeBase64(extra)) as { status?: number }).status;
        if (status === 1) return "cancelled";
        if (status === 3) return "failed";
        return "completed";
    } catch {
        return "completed";
    }
};

/** 落库行（po.ChatMessage）→ Semi Message：v1 只落库纯文本正文，工具调用不回放；
    user/assistant 正文分别包成 input_text/output_text 的 message item（Semi 只内置
    渲染这种形态）。role 防御性兜底为 assistant。 */
export const storedToChat = (row: StoredMessage): Message => {
    const isUser = row.Role === "user";
    const text: ContentItem = {
        type: "message",
        content: [{
            type: isUser ? "input_text" : "output_text",
            text: row.Content,
        }],
    };
    const chat: Message = {
        id: `m-${row.ID}`,
        role: isUser ? "user" : "assistant",
        createdAt: new Date(row.CreatedAt).getTime(),
        content: [text],
    };
    // user 气泡无状态；assistant 收口状态取落库 Extra
    if (!isUser) chat.status = storedStatus(row.Extra);
    return chat;
};