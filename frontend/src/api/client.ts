/**
 * client.ts - Axios HTTP客户端配置
 * 配置API请求的基础URL、请求拦截器（添加认证token）和响应拦截器（处理401错误）
 */

import axios from 'axios'
import { useAuthStore } from '../store/authStore'

/**
 * 创建axios实例，配置基础URL
 * baseURL从环境变量VITE_API_BASE_URL读取
 * 如果未设置，使用相对路径（适用于Docker部署，通过Nginx代理）
 * 如果设置了，使用绝对路径（适用于开发环境或前后端分离部署）
 */
const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? '/api/v1',
})

/**
 * 请求拦截器：在每个HTTP请求发送前执行
 * 从Zustand store中获取token并添加到请求头的Authorization字段
 * 使用Bearer认证方式：Authorization: Bearer <token>
 */
api.interceptors.request.use((config) => {
  // 使用useAuthStore.getState()获取store的当前状态（不在组件中时不能使用hook）
  const token = useAuthStore.getState().token
  if (token) {
    // 添加Bearer token到请求头
    config.headers.Authorization = `Bearer ${token}`
  }
  // 调试日志：记录所有API请求
  console.log('[API Client] 发送请求:', {
    method: config.method?.toUpperCase(),
    url: config.url,
    baseURL: config.baseURL,
    fullURL: `${config.baseURL}${config.url}`,
    hasToken: !!token,
  })
  return config
})

/**
 * 响应拦截器：处理HTTP响应
 * - 成功响应：直接返回
 * - 401未授权：清除登录状态
 * - 其他错误：拒绝Promise，由调用方处理
 */
api.interceptors.response.use(
  (response) => response,  // 成功响应直接返回
  (error) => {
    // 如果响应状态码是401（未授权），说明token过期或无效
    if (error.response?.status === 401) {
      // 调用logout清除本地存储的token和用户信息
      useAuthStore.getState().logout()
    }
    // 将错误继续抛出，由具体的API调用函数处理
    return Promise.reject(error)
  },
)

export default api

