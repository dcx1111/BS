import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'

// 错误处理：捕获全局未处理的错误
window.addEventListener('error', (event) => {
  console.error('全局错误:', event.error)
})

window.addEventListener('unhandledrejection', (event) => {
  console.error('未处理的Promise拒绝:', event.reason)
})

// 检查浏览器兼容性
const rootElement = document.getElementById('root')
if (!rootElement) {
  throw new Error('找不到root元素')
}

// 检查是否支持createRoot（React 18+）
try {
  createRoot(rootElement).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
} catch (error) {
  console.error('应用初始化失败:', error)
  rootElement.innerHTML = '<div style="padding: 20px; text-align: center;"><h2>浏览器不支持</h2><p>请使用现代浏览器（Chrome、Firefox、Safari、Edge等）</p></div>'
}
