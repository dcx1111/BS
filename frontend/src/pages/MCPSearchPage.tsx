import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { mcpSearch } from '../api/mcp'
import type { ImageMeta } from '../types'
import ImageCard from '../components/ImageCard'
import { useMCPSearchStore } from '../store/mcpSearchStore'
import './MCPSearchPage.css'

const FILTERS_STORAGE_KEY = 'image_list_filters'

/**
 * MCP对话式图片搜索页面
 * 允许用户使用自然语言搜索图片，AI会将查询转换为搜索条件
 * 搜索结果会缓存到localStorage，路由切换时不会清空
 */
const MCPSearchPage = () => {
  const navigate = useNavigate()
  // 从缓存Store中读取搜索结果
  const cachedQuery = useMCPSearchStore((state) => state.query)
  const cachedImages = useMCPSearchStore((state) => state.images)
  const cachedTotal = useMCPSearchStore((state) => state.total)
  const cachedPage = useMCPSearchStore((state) => state.page)
  const cachedPageSize = useMCPSearchStore((state) => state.pageSize)
  const cachedFilters = useMCPSearchStore((state) => state.filters)
  const setSearchResult = useMCPSearchStore((state) => state.setSearchResult)
  
  // 初始化状态，优先使用缓存的值
  const [query, setQuery] = useState(() => cachedQuery || '')
  const [images, setImages] = useState<ImageMeta[]>(() => cachedImages || [])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [total, setTotal] = useState(() => cachedTotal || 0)
  const [page, setPage] = useState(() => cachedPage || 1)
  const [pageSize] = useState(() => cachedPageSize || 20)
  const [filters, setFilters] = useState<Record<string, string>>(() => cachedFilters || {})
  
  // 组件挂载时从缓存恢复数据（只在首次挂载时执行）
  useEffect(() => {
    if (cachedQuery && cachedImages.length > 0) {
      setQuery(cachedQuery)
      setImages(cachedImages)
      setTotal(cachedTotal)
      setPage(cachedPage)
      setFilters(cachedFilters)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const handleSearch = async (pageNum: number = 1) => {
    if (!query.trim()) {
      setError('请输入搜索查询')
      return
    }

    setLoading(true)
    setError(null)
    
    // 开始新搜索时，先清空旧的状态，避免显示上一次搜索的条件
    // 只有在第一页搜索时才清空（翻页时不清空）
    if (pageNum === 1) {
      setFilters({})
      setImages([])
      setTotal(0)
      setPage(1)
    }
    
    try {
      const result = await mcpSearch({
        query: query.trim(),
        page: pageNum,
        pageSize,
      })
      setImages(result.items)
      setTotal(result.total)
      setPage(result.page)
      setFilters(result.filters)
      
      // 缓存搜索结果
      setSearchResult({
        query: query.trim(),
        images: result.items,
        total: result.total,
        page: result.page,
        pageSize,
        filters: result.filters,
      })
    } catch (err: any) {
      setError(err.response?.data?.message ?? '搜索失败')
      setImages([])
      setTotal(0)
      setFilters({})
      // 搜索失败时清空缓存
      setSearchResult({
        query: '',
        images: [],
        total: 0,
        page: 1,
        pageSize,
        filters: {},
      })
    } finally {
      setLoading(false)
    }
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    handleSearch(1)
  }

  const handlePageChange = (newPage: number) => {
    handleSearch(newPage)
  }

  // 将AI生成的筛选条件转换为ImageListPage的筛选格式，并跳转到图片列表页
  const handleGoToImageList = () => {
    if (Object.keys(filters).length === 0) {
      return
    }

    // 将AI返回的filters格式转换为ImageListPage的filters格式
    const imageListFilters: Record<string, string> = {
      keyword: filters.keyword || '',
      start_date: filters.start_date || '',
      end_date: filters.end_date || '',
      taken_start: filters.taken_start || '',
      taken_end: filters.taken_end || '',
      width_min: filters.width_min || '',
      width_max: filters.width_max || '',
      height_min: filters.height_min || '',
      height_max: filters.height_max || '',
      size_min_mb: filters.size_min || '', // AI返回的size_min已经是MB格式
      size_max_mb: filters.size_max || '', // AI返回的size_max已经是MB格式
      tags: filters.tags || '',
      keyword_mode: filters.keyword_mode || 'or',
      tag_mode: filters.tag_mode || 'or',
    }

    // 保存筛选条件到localStorage
    try {
      localStorage.setItem(FILTERS_STORAGE_KEY, JSON.stringify(imageListFilters))
    } catch (err) {
      console.error('Failed to save filters to storage:', err)
    }

    // 跳转到图片列表页，ImageListPage会自动从localStorage恢复筛选条件并查询
    navigate('/images')
  }

  return (
    <div className="mcp-search-page">
      <h1>AI对话式搜索</h1>
      <p className="mcp-description">
        使用自然语言搜索图片，例如："找一些风景照片"、"显示上个月拍的猫的照片"、"查找大尺寸的图片"
      </p>

      <form className="mcp-search-form" onSubmit={handleSubmit}>
        <div className="search-input-wrapper">
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="输入搜索查询..."
            className="mcp-search-input"
            disabled={loading}
          />
          <button type="submit" disabled={loading || !query.trim()} className="mcp-search-btn">
            {loading ? '搜索中...' : '搜索'}
          </button>
        </div>
      </form>

      {error && <div className="mcp-error">{error}</div>}

      {Object.keys(filters).length > 0 && (
        <div className="mcp-filters">
          <strong>搜索条件：</strong>
          <div className="filter-tags">
            {Object.entries(filters).map(([key, value]) => (
              <span key={key} className="filter-tag">
                {key}: {value}
              </span>
            ))}
          </div>
        </div>
      )}

      {total > 0 && (
        <div className="mcp-results-info">
          找到 {total} 张图片
          {Object.keys(filters).length > 0 && (
            <button onClick={handleGoToImageList} className="goto-image-list-btn">
              在图片库中使用这些条件搜索
            </button>
          )}
        </div>
      )}

      {images.length > 0 ? (
        <>
          <div className="image-grid">
            {images.map((image) => (
              <ImageCard key={image.id} image={image} />
            ))}
          </div>

          {total > pageSize && (
            <div className="pagination">
              <button
                onClick={() => handlePageChange(page - 1)}
                disabled={page <= 1 || loading}
                className="page-btn"
              >
                上一页
              </button>
              <span className="page-info">
                第 {page} 页，共 {Math.ceil(total / pageSize)} 页
              </span>
              <button
                onClick={() => handlePageChange(page + 1)}
                disabled={page >= Math.ceil(total / pageSize) || loading}
                className="page-btn"
              >
                下一页
              </button>
            </div>
          )}
        </>
      ) : (
        !loading && query && (
          <div className="mcp-empty">未找到匹配的图片</div>
        )
      )}
    </div>
  )
}

export default MCPSearchPage

