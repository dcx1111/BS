import api from './client'
import type { Tag } from '../types'

export const fetchTags = async () => {
  const { data } = await api.get<Tag[]>('/tags')
  return data
}

export const createTag = async (payload: { name: string; color: string }) => {
  const { data } = await api.post<Tag>('/tags', payload)
  return data
}

export const updateTagColor = async (tagId: number, color: string) => {
  const { data } = await api.put<Tag>(`/tags/${tagId}/color`, { color })
  return data
}

