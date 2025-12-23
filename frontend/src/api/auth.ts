/**
 * auth.ts - 认证相关API接口
 * 提供用户登录和注册的HTTP请求方法
 */

import api from './client'
import type { AuthResponse } from '../types'

/**
 * 认证请求载荷接口
 */
interface AuthPayload {
  username: string  // 用户名
  email?: string    // 邮箱（注册时必填，登录时不需要）
  password: string  // 密码（明文，后端会进行哈希验证）
}

/**
 * register - 用户注册
 * @param payload - 注册信息，包含用户名、邮箱和密码
 * @returns Promise<AuthResponse['user']> 注册成功后返回的用户信息
 */
export const register = async (payload: AuthPayload) => {
  console.log('[Auth API] 发送注册请求:', { url: '/auth/register', payload: { ...payload, password: '***' } })
  try {
    const { data } = await api.post<{ user: AuthResponse['user'] }>('/auth/register', payload)
    console.log('[Auth API] 注册成功:', data)
    return data.user
  } catch (error: any) {
    console.error('[Auth API] 注册失败:', error)
    console.error('[Auth API] 错误详情:', {
      message: error.message,
      response: error.response?.data,
      status: error.response?.status,
      config: {
        url: error.config?.url,
        baseURL: error.config?.baseURL,
        method: error.config?.method,
      },
    })
    throw error
  }
}

/**
 * login - 用户登录
 * @param payload - 登录信息，包含用户名和密码
 * @returns Promise<AuthResponse> 登录成功后返回的数据，包含token和用户信息
 */
export const login = async (payload: AuthPayload) => {
  console.log('[Auth API] 发送登录请求:', { url: '/auth/login', payload: { ...payload, password: '***' } })
  try {
    const { data } = await api.post<AuthResponse>('/auth/login', payload)
    console.log('[Auth API] 登录成功:', data)
    return data
  } catch (error: any) {
    console.error('[Auth API] 登录失败:', error)
    console.error('[Auth API] 错误详情:', {
      message: error.message,
      response: error.response?.data,
      status: error.response?.status,
      config: {
        url: error.config?.url,
        baseURL: error.config?.baseURL,
        method: error.config?.method,
      },
    })
    throw error
  }
}

