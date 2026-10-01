import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import '@douyinfe/semi-ui/react19-adapter';
import { RouterProvider } from 'react-router-dom'
import { router } from './router'
import { initScreenScale } from './conf/scale'

initScreenScale()

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>,
)
