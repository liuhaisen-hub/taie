import { Badge, Button, Card, Collapse, Input, Modal } from "@douyinfe/semi-ui";
import { useAgentChat } from "../../provider";
import type { ContentItem } from "../../types";
import styles from './index.module.css'
import type { BadgeType } from "@douyinfe/semi-ui/lib/es/badge";

const ApprovalTitle = (props: { name: string, approved: boolean, rejected: boolean }) => {
    let text: string
    let type: BadgeType
    if (props.approved) {
        text = `已批准：${props.name}`
        type = 'success'
    } else if (props.rejected) {
        text = `已拒绝：${props.name}`
        type = 'danger'
    } else {
        text = `工具 ${props.name} 等待审批`
        type = 'warning'
    }
    // const text = approved? `已批准：${name}`: rejected? `已拒绝：${name}`: `工具 ${name} 等待审批`
    return <span><Badge dot type={type} style={{ marginRight: '12px' }} />{text}</span>
}
/** approval 审批卡：工具执行前等待人工裁决。
    仅当"本卡片的 interrupt_id == 当前挂起审批 && 不在生成中"时可交互；
    裁决后 status 由 provider 置 completed(批准)/failed(拒绝)，转为结果态。 */
export const ApprovalCard = ({ item }: { item: ContentItem }) => {
    const { pendingApproval, decideApproval, generate } = useAgentChat();
    const [rejecting, setRejecting] = useState(false); // 拒绝理由输入区
    const [reason, setReason] = useState("");
    const interruptId =
        (item.extra as { interrupt_id?: string } | undefined)?.interrupt_id ?? "";
    const active = !!pendingApproval
        && pendingApproval.interruptId === interruptId && !generate;
    const approved = item.status === "completed";
    const rejected = item.status === "failed";
    const name = item.name ?? "工具";
    // 原始入参 JSON：尽量美化展示，解析失败按原文兜底——批准前先看清要执行什么
    let args = item.result ?? "";
    if (args) {
        try {
            args = JSON.stringify(JSON.parse(args), null, 2);
        } catch { /* 非 JSON 入参，原文展示 */ }
    }

    return (
        <Card
            title={<ApprovalTitle name={name} approved={approved} rejected={rejected} />}
        >
            {args && <pre className={styles.approvalArgs}>{args}</pre>}
            {active && !rejecting && (
                <div className={styles.approvalActions}>
                    <Button type="primary"
                        onClick={() => decideApproval(true)}>批准</Button>
                    <Button type="danger"
                        onClick={() => setRejecting(true)}>拒绝</Button>
                </div>
            )}
            {active && rejecting && (
                <div className={styles.approvalReason}>
                    <Input
                        value={reason}
                        placeholder="拒绝原因（可选，会告知 agent）"
                        onChange={v => setReason(v)}
                    />
                    <Button type="danger"
                        onClick={() => decideApproval(false, reason)}>确认拒绝</Button>
                    <Button type="secondary"
                        onClick={() => { setRejecting(false); setReason(""); }}>取消</Button>
                </div>
            )}
        </Card>
    );
};