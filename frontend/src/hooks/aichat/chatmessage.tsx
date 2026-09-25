import { AIChatDialogue } from "@douyinfe/semi-ui";
import type { Message } from "@douyinfe/semi-ui/lib/es/aiChatDialogue";
import { useCallback, useState } from "react";
import { useAgentChat } from "./provider/provider";
const roleConfig = {
    user: {
        name: 'User',
        avatar: 'https://lf3-static.bytednsdoc.com/obj/eden-cn/ptlz_zlp/ljhwZthlaukjlkulzlp/docs-icon.png'
    },
    assistant: {
        name: 'Assistant',
        avatar: 'https://lf3-static.bytednsdoc.com/obj/eden-cn/ptlz_zlp/ljhwZthlaukjlkulzlp/other/logo.png'
    },
    system: {
        name: 'System',
        avatar: 'https://lf3-static.bytednsdoc.com/obj/eden-cn/ptlz_zlp/ljhwZthlaukjlkulzlp/other/logo.png'
    }
};
const defaultMessages = [
    {
        role: 'system',
        id: '1',
        createAt: 1715676751919,
        content: "Hello, I'm your AI assistant.",
    },
    {
        role: 'user',
        id: '2',
        createAt: 1715676751919,
        content: "给一个 Semi Design 的 Button 组件的使用示例",
    },
    {
        role: 'assistant',
        id: '3',
        createAt: 1715676751919,
        content: "以下是一个 Semi 代码的使用示例：\n\`\`\`jsx \nimport React from 'react';\nimport { Button } from '@douyinfe/semi-ui';\n\nconst MyComponent = () => {\n  return (\n    <Button>Click me</Button>\n );\n};\nexport default MyComponent;\n\`\`\`\n",
    }
];

export const ChatMessage = () => {
    const { chat } = useAgentChat()

    return <AIChatDialogue
        align="leftRight"
        mode="bubble"
        chats={chat}
        roleConfig={roleConfig}
    />

}
