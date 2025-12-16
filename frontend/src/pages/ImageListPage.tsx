import { useEffect, useState } from 'react'
import { fetchImages } from '../api/images'
import type { ImageMeta } from '../types'
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
}

const ImageListPage = () => {
  const [images, setImages] = useState<ImageMeta[]>([])
  const [filters, setFilters] = useState(DEFAULT_FILTERS)
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const [showAdvanced, setShowAdvanced] = useState(false)

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
        // 将MB转换为字节
        size_min: filters.size_min_mb ? Number(filters.size_min_mb) * 1024 * 1024 : undefined,
        size_max: filters.size_max_mb ? Number(filters.size_max_mb) * 1024 * 1024 : undefined,
        tags: filters.tags,
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

  useEffect(() => {
    loadImages()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const handleChange = (field: keyof typeof DEFAULT_FILTERS, value: string) => {
    setFilters((prev) => ({ ...prev, [field]: value }))
  }

  const handleReset = () => {
    setFilters(DEFAULT_FILTERS)
  }

  return (
    <div className="image-list-page">
      <section className="filter-panel">
        <div className="filter-row">
          <div>
            <label>关键词</label>
            <input value={filters.keyword} onChange={(e) => handleChange('keyword', e.target.value)} placeholder="文件名" />
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
              <input type="number" min="0" value={filters.size_min_mb} onChange={(e) => handleChange('size_min_mb', e.target.value)} />
            </div>
            <div>
              <label>文件大小最大(MB)</label>
              <input type="number" min="0" value={filters.size_max_mb} onChange={(e) => handleChange('size_max_mb', e.target.value)} />
            </div>
            <div>
              <label>标签（逗号分隔，需同时匹配）</label>
              <input value={filters.tags} onChange={(e) => handleChange('tags', e.target.value)} placeholder="例如：旅行, 北京" />
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

