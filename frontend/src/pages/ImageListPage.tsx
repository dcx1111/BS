import { useEffect, useState } from 'react'
import { useLocation } from 'react-router-dom'
import { fetchImages } from '../api/images'
import type { ImageMeta } from '../types'
import { useImageListStore } from '../store/imageListStore'
import ImageCard from '../components/ImageCard'
import './ImageListPage.css'

const DEFAULT_FILTERS = {
  keyword: '',
  start_date: '',
  end_date: '',
  taken_start: '',
  taken_end: '',
  width_min: '',
  width_max: '',
  height_min: '',
  height_max: '',
  size_min_mb: '',
  size_max_mb: '',
  tags: '',
  keyword_mode: 'or', // 'and' 或 'or'，表示关键词和其他条件的关系
  tag_mode: 'or',     // 'and' 或 'or'，表示标签之间的关系
}

const FILTERS_STORAGE_KEY = 'image_list_filters'

// 从localStorage加载筛选状态
const loadFiltersFromStorage = (): typeof DEFAULT_FILTERS => {
  try {
    const stored = localStorage.getItem(FILTERS_STORAGE_KEY)
    if (stored) {
      return { ...DEFAULT_FILTERS, ...JSON.parse(stored) }
    }
  } catch (err) {
    console.error('Failed to load filters from storage:', err)
  }
  return DEFAULT_FILTERS
}

// 保存筛选状态到localStorage
const saveFiltersToStorage = (filters: typeof DEFAULT_FILTERS) => {
  try {
    localStorage.setItem(FILTERS_STORAGE_KEY, JSON.stringify(filters))
  } catch (err) {
    console.error('Failed to save filters to storage:', err)
  }
}

const ImageListPage = () => {
  const location = useLocation()
  const [images, setImages] = useState<ImageMeta[]>([])
  const [filters, setFilters] = useState(loadFiltersFromStorage)
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const [showAdvanced, setShowAdvanced] = useState(false)
  const hasNewImages = useImageListStore((state) => state.hasNewImages)
  const setHasNewImages = useImageListStore((state) => state.setHasNewImages)
  const [lastLocationKey, setLastLocationKey] = useState<string | null>(null)

  const loadImages = async () => {
    setLoading(true)
    try {
      const params: Record<string, string | number | undefined> = {
        keyword: filters.keyword,
        start_date: filters.start_date,
        end_date: filters.end_date,
        taken_start: filters.taken_start,
        taken_end: filters.taken_end,
        width_min: filters.width_min,
        width_max: filters.width_max,
        height_min: filters.height_min,
        height_max: filters.height_max,
        // 直接传递MB值（可以是小数），后端会处理转换为字节
        size_min: filters.size_min_mb || undefined,
        size_max: filters.size_max_mb || undefined,
        tags: filters.tags,
        keyword_mode: filters.keyword_mode,
        tag_mode: filters.tag_mode,
        page: 1,
        pageSize: 40,
      }
      const data = await fetchImages(params)
      setImages(data.items)
      setTotal(data.total)
    } finally {
      setLoading(false)
    }
  }

  // 当路由切换到图片列表页面时，检查是否有新图片，如果有则重新加载
  useEffect(() => {
    // 检测路由切换：如果 location.key 发生变化，说明用户切换到了这个页面
    if (lastLocationKey !== location.key) {
      setLastLocationKey(location.key)
      
      // 如果有新图片标记，则重新加载图片，并清除标记
      const shouldRefresh = hasNewImages
      if (shouldRefresh) {
        setHasNewImages(false)
      }
      // 无论是否有新图片，路由切换时都需要加载数据（首次访问或手动切换）
      loadImages()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [location.key])

  const handleChange = (field: keyof typeof DEFAULT_FILTERS, value: string) => {
    setFilters((prev) => {
      const newFilters = { ...prev, [field]: value }
      saveFiltersToStorage(newFilters)
      return newFilters
    })
  }

  const handleReset = () => {
    setFilters(DEFAULT_FILTERS)
    localStorage.removeItem(FILTERS_STORAGE_KEY)
  }

  // 当筛选条件改变时，保存到localStorage
  useEffect(() => {
    saveFiltersToStorage(filters)
  }, [filters])

  return (
    <div className="image-list-page">
      <section className="filter-panel">
        <div className="filter-row">
          <div className="keyword-input-group">
            <label>关键词</label>
            <div className="input-with-mode">
              <input value={filters.keyword} onChange={(e) => handleChange('keyword', e.target.value)} placeholder="文件名" />
              <button
                type="button"
                className={`mode-toggle ${filters.keyword_mode === 'and' ? 'active' : ''}`}
                onClick={() => handleChange('keyword_mode', filters.keyword_mode === 'and' ? 'or' : 'and')}
                title={filters.keyword_mode === 'and' ? 'AND模式：关键词和其他条件都要满足' : 'OR模式：关键词或其他条件满足即可'}
              >
                {filters.keyword_mode.toUpperCase()}
              </button>
            </div>
          </div>
          <div>
            <label>创建时间起</label>
            <input type="datetime-local" value={filters.start_date} onChange={(e) => handleChange('start_date', e.target.value)} />
          </div>
          <div>
            <label>创建时间止</label>
            <input type="datetime-local" value={filters.end_date} onChange={(e) => handleChange('end_date', e.target.value)} />
          </div>
          <div className="filter-actions">
            <button onClick={loadImages} disabled={loading}>
              {loading ? '查询中...' : '查询'}
            </button>
            <button className="secondary-btn" onClick={handleReset} disabled={loading}>
              重置
            </button>
            <button className="secondary-btn" onClick={() => setShowAdvanced((prev) => !prev)}>
              {showAdvanced ? '收起筛选' : '展开筛选'}
            </button>
          </div>
        </div>

        {showAdvanced && (
          <div className="advanced-grid">
            <div>
              <label>拍摄时间起</label>
              <input type="datetime-local" value={filters.taken_start} onChange={(e) => handleChange('taken_start', e.target.value)} />
            </div>
            <div>
              <label>拍摄时间止</label>
              <input type="datetime-local" value={filters.taken_end} onChange={(e) => handleChange('taken_end', e.target.value)} />
            </div>
            <div>
              <label>宽度最小(px)</label>
              <input type="number" min="0" value={filters.width_min} onChange={(e) => handleChange('width_min', e.target.value)} />
            </div>
            <div>
              <label>宽度最大(px)</label>
              <input type="number" min="0" value={filters.width_max} onChange={(e) => handleChange('width_max', e.target.value)} />
            </div>
            <div>
              <label>高度最小(px)</label>
              <input type="number" min="0" value={filters.height_min} onChange={(e) => handleChange('height_min', e.target.value)} />
            </div>
            <div>
              <label>高度最大(px)</label>
              <input type="number" min="0" value={filters.height_max} onChange={(e) => handleChange('height_max', e.target.value)} />
            </div>
            <div>
              <label>文件大小最小(MB)</label>
              <input type="number" min="0" step="0.1" value={filters.size_min_mb} onChange={(e) => handleChange('size_min_mb', e.target.value)} />
            </div>
            <div>
              <label>文件大小最大(MB)</label>
              <input type="number" min="0" step="0.1" value={filters.size_max_mb} onChange={(e) => handleChange('size_max_mb', e.target.value)} />
            </div>
            <div className="tag-input-group">
              <label>标签（逗号分隔）</label>
              <div className="input-with-mode">
                <input value={filters.tags} onChange={(e) => handleChange('tags', e.target.value)} placeholder="例如：旅行, 北京 或 旅行，北京" />
                <button
                  type="button"
                  className={`mode-toggle ${filters.tag_mode === 'and' ? 'active' : ''}`}
                  onClick={() => handleChange('tag_mode', filters.tag_mode === 'and' ? 'or' : 'and')}
                  title={filters.tag_mode === 'and' ? 'AND模式：所有标签都要匹配' : 'OR模式：任意一个标签匹配即可'}
                >
                  {filters.tag_mode.toUpperCase()}
                </button>
              </div>
            </div>
          </div>
        )}
      </section>

      <p className="total-tip">共 {total} 张图片</p>

      <section className="image-grid">
        {images.map((image) => (
          <ImageCard key={image.id} image={image} />
        ))}
        {!loading && images.length === 0 && <div className="empty-state">暂无图片，去上传一张吧。</div>}
      </section>
    </div>
  )
}

export default ImageListPage

