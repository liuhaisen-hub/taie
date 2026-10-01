import { AIChatProvider } from "@/components/aichat"
import style from './index.module.css'
export const HomePage = () => {
    return <div className={style.home}>
        <AIChatProvider />
    </div>
}