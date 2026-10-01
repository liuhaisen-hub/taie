import type { Message } from "@douyinfe/semi-ui/lib/es/aiChatDialogue";
import type { ContentItem } from "../../types";
import { Badge, Card } from "@douyinfe/semi-ui";
import styles from './index.module.css'
/** function_call 状态卡：调用中显示"正在调用xx工具"；tool_result 到达后（status 置
    completed）显示"获取到xx工具结果"，可展开查看结果全文。工具入参后端不下发，不展示。 */
export const ToolCallCard = ({ item, message }: { item: ContentItem; message?: Message }) => {
    const calls = (Array.isArray(message?.content) ? message.content : []) as ContentItem[];
    const result = calls.find(c => c.type === "tool_result" && c.call_id === item.call_id)?.result;
    const name = item.name ?? item.call_id ?? "工具";
    const done = item.status === "completed" || result !== undefined;
    return (
        <Card>
            <div className={styles.root}>
                {done ? <span>
                    <Badge dot type='success' style={{ marginRight: '12px' }} />{`获取到${name}工具结果`}
                </span>
                    : <span><Badge dot type="warning" style={{ marginRight: '12px' }} />{`正在调用${name}工具……`}</span>
                }
            </div>
        </Card>
    );
};