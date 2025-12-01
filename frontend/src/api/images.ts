import api from './client'
import type { ImageMeta, PaginatedResponse } from '../types'

export const fetchImages = async (params: Record<string, string | number | undefined>) => {
  const { data } = await api.get<PaginatedResponse<ImageMeta>>('/images', {
    params,
  })
  return data
}

export const uploadImage = async (file: File, tags: string[]) => {
  const formData = new FormData()
  formData.append('file', file)
  tags.forEach((tag) => formData.append('tags[]', tag))
  const { data } = await api.post<ImageMeta>('/images/upload', formData)
  return data
}

export const fetchImageDetail = async (id: string) => {
  const { data } = await api.get<ImageMeta>(`/images/${id}`)
  return data
}

export const deleteImage = async (id: string) => {
  await api.delete(`/images/${id}`)
}

