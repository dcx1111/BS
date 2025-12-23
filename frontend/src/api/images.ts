/**
 * images.ts - 图片相关API接口
 * 提供图片的查询、上传、删除、编辑等HTTP请求方法
 */

import api from './client'
import type { ImageMeta, PaginatedResponse } from '../types'

/**
 * fetchImages - 获取图片列表（支持分页和筛选）
 * @param params - 查询参数对象，包含分页信息和筛选条件
 *   - keyword: 关键词搜索（匹配文件名）
 *   - page: 页码（从1开始）
 *   - pageSize: 每页数量
 *   - start/end: 创建时间范围（ISO格式字符串）
 *   - width_min/width_max: 宽度范围
 *   - height_min/height_max: 高度范围
 *   - size_min/size_max: 文件大小范围（MB，可以是小数，如1.5表示1.5MB）
 *   - taken_start/taken_end: 拍摄时间范围
 *   - tags: 标签筛选（逗号分隔的标签名）
 * @returns Promise<PaginatedResponse<ImageMeta>> 分页响应数据，包含图片列表和总数
 */
export const fetchImages = async (params: Record<string, string | number | undefined>) => {
  const { data } = await api.get<PaginatedResponse<ImageMeta>>('/images', {
    params,
  })
  return data
}

/**
 * uploadImage - 上传图片
 * @param file - 要上传的图片文件对象
 * @param tags - 图片标签名称数组
 * @param useAI - 是否使用AI自动生成标签（默认true）
 * @returns Promise<ImageMeta> 上传成功后返回的图片元数据
 */
export const uploadImage = async (file: File, tags: string[], useAI: boolean = true) => {
  // 创建FormData对象用于multipart/form-data格式的文件上传
  const formData = new FormData()
  // 添加文件到表单数据
  formData.append('file', file)
  // 添加所有标签到表单数据，使用tags[]数组格式以便后端解析
  tags.forEach((tag) => formData.append('tags[]', tag))
  // 添加是否使用AI的标志
  formData.append('use_ai', useAI ? 'true' : 'false')
  // 发送POST请求到 /images/upload 端点
  // api.post 会自动添加认证token（在axios拦截器中处理）
  const { data } = await api.post<ImageMeta>('/images/upload', formData)
  return data
}

/**
 * fetchImageDetail - 获取图片详细信息
 * @param id - 图片ID（字符串格式）
 * @returns Promise<ImageMeta> 包含完整信息的图片元数据（包括EXIF、标签、缩略图等）
 */
export const fetchImageDetail = async (id: string) => {
  const { data } = await api.get<ImageMeta>(`/images/${id}`)
  return data
}

/**
 * deleteImage - 删除图片
 * @param id - 要删除的图片ID
 * @returns Promise<void> 删除操作不返回数据
 */
export const deleteImage = async (id: string) => {
  await api.delete(`/images/${id}`)
}

/**
 * addImageTag - 为图片添加标签
 * 如果标签不存在，会自动创建（颜色为空）
 * @param imageId - 图片ID
 * @param tagName - 要添加的标签名称
 * @returns Promise<void>
 */
export const addImageTag = async (imageId: string, tagName: string) => {
  await api.post(`/images/${imageId}/tags/add`, { tagName })
}

/**
 * updateImageTag - 更新图片的标签（将旧标签替换为新标签）
 * 如果新标签不存在，会使用旧标签的颜色创建新标签
 * @param imageId - 图片ID
 * @param oldTagId - 旧标签的ID
 * @param newTagName - 新标签的名称
 * @returns Promise<void>
 */
export const updateImageTag = async (imageId: string, oldTagId: number, newTagName: string) => {
  await api.put(`/images/${imageId}/tags/update`, { oldTagId, newTagName })
}

/**
 * removeImageTag - 删除图片的标签关联
 * @param imageId - 图片ID
 * @param tagId - 要删除的标签ID
 * @returns Promise<void>
 */
export const removeImageTag = async (imageId: string, tagId: string) => {
  await api.delete(`/images/${imageId}/tags/${tagId}`)
}

/**
 * verifyImportAccount - 验证要导入的账户凭据并获取其图片列表
 * @param username - 用户名或邮箱
 * @param password - 密码
 * @returns Promise<{images: ImageMeta[], userId: number}>
 */
export const verifyImportAccount = async (username: string, password: string) => {
  const { data } = await api.post<{ images: ImageMeta[]; userId: number }>('/images/import/verify', {
    username,
    password,
  })
  return data
}

/**
 * importImages - 导入其他用户的图片
 * @param username - 用户名或邮箱
 * @param password - 密码
 * @param imageIds - 要导入的图片ID列表
 * @returns Promise<{message: string, importedImages: ImageMeta[]}>
 */
export const importImages = async (username: string, password: string, imageIds: number[]) => {
  const { data } = await api.post<{ message: string; importedImages: ImageMeta[] }>('/images/import', {
    username,
    password,
    imageIds,
  })
  return data
}

