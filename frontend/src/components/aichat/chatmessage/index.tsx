import { AIChatDialogue, Badge, Card } from "@douyinfe/semi-ui";
import type AIChatDialogueClass from "@douyinfe/semi-ui/lib/es/aiChatDialogue";
import type { Message } from "@douyinfe/semi-ui/lib/es/aiChatDialogue";
import { useEffect, useRef, useState } from "react";
import { useAgentChat } from "../provider";
import type { ContentItem } from "../types";
import styles from './index.module.css';
import { ApprovalCard } from "./approvalCard";
import { ToolCallCard } from "./toolcallCard";

const roleConfig = {
    user: {
        name: 'User',
        avatar: '/user.png'
    },
    assistant: {
        name: 'Assistant',
        avatar: '/ai.png'
    },
    system: {
        name: 'System',
        avatar: 'https://lf3-static.bytednsdoc.com/obj/eden-cn/ptlz_zlp/ljhwZthlaukjlkulzlp/other/logo.png'
    }
};

/** ContentItem[] → 纯文本（复制/反馈等操作取内容用） */
const messageToText = (message?: Message): string => {
    const content = message?.content;
    if (typeof content === "string") return content;
    if (!Array.isArray(content)) return "";
    return content.map(item => {
        const it = item as ContentItem;
        if (it.type === "output_text") return it.text ?? "";
        if (it.type === "message") {
            const inner = (it.content as ContentItem[] | undefined) ?? [];
            return inner
                .filter(c => c.type === "input_text" || c.type === "output_text")
                .map(c => c.text ?? "")
                .join("");
        }
        return "";
    }).join("");
};





// function_call / tool_result / approval 为协议自定义渲染：function_call 注册状态卡
// （上述 ToolCallCard），tool_result 的结果已由状态卡按 call_id 反查展示，这里渲染
// 为空，approval 注册审批卡（上述 ApprovalCard），避免 Semi 对未知类型告警。
// output_text / message / reasoning 等走 Semi 原生渲染。
// 模块级常量保证引用稳定（无需 useCallback——它要求回调是函数，不能包对象）。
const customRenderers = {
    function_call: (item: ContentItem, message?: Message) =>
        <ToolCallCard item={item} message={message} />,
    tool_result: () => null,
    approval: (item: ContentItem) => <ApprovalCard item={item} />,
};

export const ChatMessage = () => {
    const { chat, onChatsChange, resetChat, sessionId } = useAgentChat();
    // Semi 的类组件类型不满足 ComponentRef 约束，直接用实例类型标注
    const dialogueRef = useRef<AIChatDialogueClass>(null);
    /** 上一次会话 ID：识别"会话切换"这条信号 */
    const prevSessionIdRef = useRef(sessionId);

    // Semi 只在"消息条数变多"时自动滚底，切会话是整表替换、条数不一定增长，
    // 不会触发；所以监听 sessionId 变化（loadSession 完成时更新），手动滚底。
    // 注意 MarkdownRender 的正文是异步 evaluate 后挂载的，首帧只有空占位、
    // 高度未定，滚一次会停在半路——因此监听容器 DOM 变更，正文每次挂载都
    // 重新贴底，静默一段时间后停止监听。
    useEffect(() => {
        if (prevSessionIdRef.current === sessionId) return;
        prevSessionIdRef.current = sessionId;
        const dialogue = dialogueRef.current;
        const container = dialogue?.containerRef.current;
        if (!dialogue || !container) return;

        dialogue.scrollToBottom(true);

        let quietTimer = 0;
        const observer = new MutationObserver(() => {
            dialogue.scrollToBottom(false);
            window.clearTimeout(quietTimer);
            quietTimer = window.setTimeout(() => observer.disconnect(), 300);
        });
        observer.observe(container, { childList: true, subtree: true, characterData: true });
        return () => {
            window.clearTimeout(quietTimer);
            observer.disconnect();
        };
    }, [sessionId]);

    return <AIChatDialogue
        ref={dialogueRef}
        className={styles.dialogue}
        align="leftRight"
        mode="bubble"
        chats={chat}
        roleConfig={roleConfig}
        onChatsChange={onChatsChange}
        onMessageCopy={message => { void navigator.clipboard.writeText(messageToText(message)); }}
        onMessageReset={() => resetChat()}
        // TODO: 反馈/删除/分享待后端支持后接入真实行为
        onMessageGoodFeedback={message => console.log("good feedback", message?.id)}
        onMessageBadFeedback={message => console.log("bad feedback", message?.id)}
        onMessageDelete={message => console.log("delete", message?.id)}
        onMessageShare={message => console.log("share", message?.id)}
        onImageClick={image => { if (image?.image_url) window.open(image.image_url, "_blank"); }}
        onFileClick={file => { if (file?.file_url) window.open(file.file_url, "_blank"); }}
        renderDialogueContentItem={customRenderers}
    />
}
