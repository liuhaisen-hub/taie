import { AIChatProvider } from "@/hooks/aichat"
import style from './index.module.css'
export const HomePage = () => {
    return <div className={style.home}>
        <AIChatProvider />
    </div>
}