// 按窗口宽度相对设计基准等比缩放整个界面（zoom 作用于 html，
// 自有样式与 Semi 组件的固定 px 样式会一并缩放）。
// zoom 会让子树布局与视口单位错位：应用内宽度一律用 100%/auto；
// 确需视口尺寸的场景（如弹窗），用 calc(100vw / var(--app-zoom)) 换算回布局空间。
const DESIGN_WIDTH = 1440
const MIN_SCALE = 0.85
const MAX_SCALE = 1.2

const applyScreenScale = () => {
    const scale = Math.min(Math.max(window.innerWidth / DESIGN_WIDTH, MIN_SCALE), MAX_SCALE)
    document.documentElement.style.zoom = scale.toFixed(3)
    document.documentElement.style.setProperty('--app-zoom', scale.toFixed(3))
}

export const initScreenScale = () => {
    applyScreenScale()
    window.addEventListener('resize', applyScreenScale)
}
