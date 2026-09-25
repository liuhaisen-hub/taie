import { AIChatInput } from '@douyinfe/semi-ui';
import type { MessageContent } from '@douyinfe/semi-ui/lib/es/aiChatInput';
import { useAgentChat } from './provider/provider';
const uploadProps = { action: "https://api.semi.design/upload" };
const outerStyle = { margin: 12 };
export const ChatInputer = () => {
    const { sendMessage } = useAgentChat()
    const onMessageSend = (props: MessageContent) => {
        const text = (props.inputContents ?? [])
            .filter(c => c.type === "text")
            .map(c => (c as { text?: string }).text ?? "")
            .join("");
        if (text.trim()) sendMessage(text);
    }
    return (
        <AIChatInput
            placeholder={'输入内容或者上传内容...'}
            uploadProps={uploadProps}
            onMessageSend={onMessageSend}
            style={outerStyle}
        />
    );
}
