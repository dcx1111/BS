import api from './client'
import type { AuthResponse } from '../types'

interface AuthPayload {
  username: string
  email?: string
  password: string
}

export const register = async (payload: AuthPayload) => {
  const { data } = await api.post<{ user: AuthResponse['user'] }>('/auth/register', payload)
  return data.user
}

export const login = async (payload: AuthPayload) => {
  const { data } = await api.post<AuthResponse>('/auth/login', payload)
  return data
}

