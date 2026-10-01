import { IconBox, IconEyeOpened, IconMicrophone, IconRotate, IconTextRectangle } from "@douyinfe/semi-icons";

export const AgentType = [
    {
        name: '多模态模型',
        icon: <IconBox />,
        value: 1
    },
    {
        name: '语言模型',
        icon: <IconTextRectangle />,
        value: 2
    },
    {
        name: '向量模型',
        icon: <IconRotate />,
        value: 3
    },
    {
        name: '视觉模型',
        icon: <IconEyeOpened />,
        value: 4
    },
    {
        name: '语音模型',
        icon: <IconMicrophone />,
        value: 5
    }
]