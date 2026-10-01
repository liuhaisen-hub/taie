/* 临时复现页：驱动 AIChatDialogue 走一遍"流式→done"，复刻 provider 的消息形态。
   URL 加 ?done=1 时直接以完成态渲染（对照组）；控制台输出两个时刻的 DOM 快照。 */
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import "@douyinfe/semi-ui/react19-adapter";
import { AIChatDialogue } from "@douyinfe/semi-ui";
import type AIChatDialogueClass from "@douyinfe/semi-ui/lib/es/aiChatDialogue";
import type { Message } from "@douyinfe/semi-ui/lib/es/aiChatDialogue";
import { useEffect, useRef, useState } from "react";

const roleConfig = {
    user: {
        name: 'User',
        avatar: 'https://lf3-static.bytednsdoc.com/obj/eden-cn/ptlz_zlp/ljhwZthlaukjlkulzlp/docs-icon.png'
    },
    assistant: {
        name: 'Assistant',
        avatar: 'https://lf3-static.bytednsdoc.com/obj/eden-cn/ptlz_zlp/ljhwZthlaukjlkulzlp/other/logo.png'
    },
};

const FULL_TEXT = [
    "## 项目结构分析",
    "",
    "这是一个 **前后端分离** 的桌面应用，主要特点如下：",
    "",
    "### 核心模块",
    "",
    "| 模块 | 职责 | 说明 |",
    "| --- | --- | --- |",
    "| frontend | UI 层 | React + Semi Design |",
    "| internal | 服务层 | Go + Wails |",
    "",
    "### 关键代码",
    "",
    "```go",
    "func main() {",
    "\trun := wails.Run(&options.App{})",
    "\tif err != nil { panic(err) }",
    "}",
    "```",
    "",
    "1. 首先初始化配置",
    "2. 然后启动服务",
    "3. 最后监听事件",
    "",
    "> 注意：流式渲染时 Markdown 是不完整的，最后一行可能被截断",
].join("\n");

const CHUNKS = FULL_TEXT.match(/[\s\S]{1,24}/g) ?? [];

/** 与 provider.flush 一致：把增量并入最后一个 message item 的内层 output_text */
const withAppended = (content: Message["content"], chunk: string): Message["content"] => {
    const items = Array.isArray(content) ? [...content] : [];
    const last = items[items.length - 1] as { type?: string; content?: unknown } | undefined;
    if (last?.type === "message") {
        const inner = Array.isArray(last.content) ? [...(last.content as unknown[])] : [];
        const lastInner = inner[inner.length - 1] as { type?: string; text?: string } | undefined;
        if (lastInner?.type === "output_text") {
            inner[inner.length - 1] = { ...lastInner, text: (lastInner.text ?? "") + chunk };
        } else {
            inner.push({ type: "output_text", text: chunk });
        }
        items[items.length - 1] = { ...last, content: inner };
        return items;
    }
    items.push({ type: "message", content: [{ type: "output_text", text: chunk }] });
    return items;
};

const App = () => {
    const [chat, setChat] = useState<Message[]>([
        { id: "u-1", role: "user", content: [{ type: "message", content: [{ type: "input_text", text: "帮我分析下项目结构" }] }], createdAt: 1 },
        { id: "a-1", role: "assistant", content: [], status: "in_progress", createdAt: 2 },
    ]);
    const dialogueRef = useRef<AIChatDialogueClass>(null);
    const idxRef = useRef(0);
    const phaseRef = useRef<"streaming" | "done">("streaming");
    const [phase, setPhase] = useState<"streaming" | "done">("streaming");
    const forceDone = new URLSearchParams(location.search).has("done");

    useEffect(() => {
        if (forceDone) {
            // 对照组：跳过流式，直接以完成态渲染
            setChat([
                { id: "u-1", role: "user", content: [{ type: "message", content: [{ type: "input_text", text: "帮我分析下项目结构" }] }], createdAt: 1 },
                { id: "a-1", role: "assistant", content: [{ type: "message", content: [{ type: "output_text", text: FULL_TEXT }] }], status: "completed", createdAt: 2 },
            ]);
            setPhase("done");
            return;
        }
        const timer = window.setInterval(() => {
            const i = idxRef.current;
            if (i >= CHUNKS.length) {
                window.clearInterval(timer);
                phaseRef.current = "done";
                setPhase("done");
                // done：整体替换为服务端形态 + status completed（与 provider 的 done 分支一致）
                setChat(prev => prev.map(m => m.id === "a-1"
                    ? { ...m, content: [{ type: "message", content: [{ type: "output_text", text: FULL_TEXT }] }], status: "completed" }
                    : m));
                return;
            }
            idxRef.current = i + 1;
            const chunk = CHUNKS[i];
            setChat(prev => prev.map(m => m.id === "a-1" ? { ...m, content: withAppended(m.content, chunk) } : m));
        }, 60);
        return () => window.clearInterval(timer);
    }, [forceDone]);

    // 抓两个时刻的气泡内部 DOM 结构
    useEffect(() => {
        const dump = (tag: string) => {
            const el = dialogueRef.current?.containerRef.current?.querySelector(".semi-ai-chat-dialogue-content");
            console.log(`[${tag}]`, el?.innerHTML?.slice(0, 600));
        };
        if (phase === "streaming" && idxRef.current >= 20 && idxRef.current < 22) dump(`mid-stream-${idxRef.current}`);
        if (phase === "done") window.setTimeout(() => dump("after-done"), 100);
    }, [chat, phase]);

    return (
        <div style={{ height: "100vh", padding: 12, boxSizing: "border-box" }}>
            <div data-phase={phase} style={{ fontFamily: "monospace", fontSize: 12, color: "#888" }}>phase: {phase}</div>
            <AIChatDialogue
                ref={dialogueRef}
                align="leftRight"
                mode="bubble"
                chats={chat}
                roleConfig={roleConfig}
                style={{ height: "90vh" }}
            />
        </div>
    );
};

createRoot(document.getElementById("root")!).render(
    <StrictMode>
        <App />
    </StrictMode>,
);
