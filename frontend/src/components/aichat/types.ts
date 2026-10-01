/**
 * 协议 v3：payload 对齐 OpenAI Response Object / Semi ContentItem，
 * 与 Go 侧 internal/pkg/events 一一对应。
 */

/** ContentItem 带索引签名：可直接赋给 Semi Message.content（其联合类型含 CustomObject） */
export type ContentItem = {
    [key: string]: any;
    id?: string;
    /** message | output_text | function_call（自定义，前端注册状态卡渲染） | tool_result（自定义，不单独渲染） | approval（自定义，审批卡：extra.interrupt_id 为续跑定位键） */
    type?: string;
    /** in_progress | completed | failed */
    status?: string;
    role?: string;
    /** output_text 文本：delta 事件里是增量，done 事件里是全文 */
    text?: string;
    /** function_call / tool_result 归位键 */
    call_id?: string;
    name?: string;
    /** tool_result 结果文本；approval 卡复用为入参 JSON（审批前展示给用户） */
    result?: string;
};

export interface ChatMessage {
    role: string;
    id?: string;
    createdAt?: number;
    status?: string;
    content?: ContentItem[];
}

export type ChatEventKind =
    | "begin" | "delta" | "tool_call" | "tool_result"
    | "approval"   // 工具等待审批，extra.interrupt_id 需原样回传
    | "done" | "error";

export interface ChatEvent {
    kind: ChatEventKind;
    message: ChatMessage;
}

export interface ChatRequest {
    session_id: number; // 0 = 首条消息，后端新建会话后经 begin 事件回推 ID
    user_input: string;
    /** 配置区透传（深度思考/模型选择等），后端暂不消费 */
    extra?: Record<string, unknown>;
}

/** 审批裁决请求（agent/approval 通道）：interrupt_id 来自 approval 事件，必须原样回传 */
export interface ApprovalRequest {
    session_id: number;
    interrupt_id: string;
    approved: boolean;
    /** 拒绝理由，approved=false 时生效，会作为工具结果喂回模型 */
    reason?: string;
}

/** 配置区取值（Configure field -> value），整体透传进 ChatRequest.extra。
    带索引签名的类型别名：可直接赋给 Record<string, unknown> */
export type ConfigureState = {
    [key: string]: unknown;
    /** standard | deep */
    deepThink?: string;
    model?: string;
};
