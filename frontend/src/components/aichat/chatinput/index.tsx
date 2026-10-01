import { AIChatInput } from '@douyinfe/semi-ui';
import type { MessageContent } from '@douyinfe/semi-ui/lib/es/aiChatInput';
import type { customRequestArgs } from '@douyinfe/semi-ui/lib/es/upload';
import * as ModelServices from '../../../../bindings/taie/internal/services/modelservices'
import { useCallback, useEffect, useRef, useState } from 'react';
import { useAgentChat } from '../provider';
import type { ConfigureState } from '../types';
import styles from './index.module.css';

const { Configure } = AIChatInput;

/**
 * 只本地预览的上传：objectURL 挂到 FileItem.url（chatInputToMessage 会把它映射为
 * input_image.image_url / input_file.file_url），不发任何网络请求。
 * TODO: 消息删除/会话重置时 revoke objectURL
 */
const uploadProps = {
    // action 为 UploadProps 必填字段，走 customRequest 时不生效，置空即可
    action: "",
    customRequest: ({ fileInstance, file, onSuccess }: customRequestArgs) => {
        file.url = URL.createObjectURL(fileInstance);
        onSuccess({});
    },
};

export const ChatInputer = () => {
    // generating false→true 时 Semi 会清空输入并切换为停止按钮
    const { sendMessage, stop, generate } = useAgentChat();
    /** 配置区当前值：发送时整体透传进 ChatRequest.extra */
    const configureRef = useRef<ConfigureState>({});
    /** 模型下拉选项：来自后端 ai_model 表 */
    const [modelOptions, setModelOptions] = useState<{ value: string; label: string }[]>([]);

    useEffect(() => {
        let cancelled = false;
        // 注意 Wails 绑定的请求字段就是 modeName（Go 侧 json tag 拼写已固化）
        ModelServices.List({ modeName: "", page: 1, size: 50 })
            .then(resp => {
                if (cancelled || !resp?.items) return;
                setModelOptions(resp.items
                    .filter((m): m is NonNullable<typeof m> => Boolean(m?.modelName))
                    .map(m => ({ value: m.modelName, label: m.modelName })));
            })
            .catch(() => { /* 模型列表拉取失败不阻塞输入框 */ });
        return () => { cancelled = true; };
    }, []);

    const onMessageSend = useCallback((props: MessageContent) => {
        // 文本提取与 user 气泡回显都在 provider 内经 chatInputToMessage 官方转换完成
        sendMessage(props, { ...configureRef.current });
    }, [sendMessage]);

    const onConfigureChange = useCallback((value: Record<string, unknown>) => {
        configureRef.current = { ...configureRef.current, ...value } as ConfigureState;
    }, []);

    const renderConfigureArea = useCallback((className?: string) => (
        <div className={[styles.configure, className].filter(Boolean).join(" ")}>
            <Configure.Select field="model" optionList={modelOptions} placeholder="模型" filter />
        </div>
    ), [modelOptions]);

    return (
        <AIChatInput
            className={styles.input}
            placeholder={'输入内容或者上传内容...'}
            uploadProps={uploadProps}
            renderConfigureArea={renderConfigureArea}
            onConfigureChange={onConfigureChange}
            onMessageSend={onMessageSend}
            onStopGenerate={stop}
            generating={generate}
        />
    );
}
