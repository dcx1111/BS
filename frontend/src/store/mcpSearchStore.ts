import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { ImageMeta } from '../types'

/**
 * MCP搜索结果缓存Store
 * 用于缓存AI搜索的结果，路由切换时不会清空
 */
interface MCPSearchState {
  query: string                    // 搜索查询
  images: ImageMeta[]              // 搜索结果图片列表
  total: number                    // 总数量
  page: number                     // 当前页码
  pageSize: number                 // 每页数量
  filters: Record<string, string>  // 搜索过滤器
  // 更新搜索结果
  setSearchResult: (result: {
    query: string
    images: ImageMeta[]
    total: number
    page: number
    pageSize: number
    filters: Record<string, string>
  }) => void
  // 清空搜索结果
  clear: () => void
}

export const useMCPSearchStore = create<MCPSearchState>()(
  persist(
    (set) => ({
      query: '',
      images: [],
      total: 0,
      page: 1,
      pageSize: 20,
      filters: {},
      setSearchResult: (result) => set({
        query: result.query,
        images: result.images,
        total: result.total,
        page: result.page,
        pageSize: result.pageSize,
        filters: result.filters,
      }),
      clear: () => set({
        query: '',
        images: [],
        total: 0,
        page: 1,
        filters: {},
      }),
    }),
    {
      name: 'image-manager-mcp-search',  // localStorage key
    },
  ),
)

