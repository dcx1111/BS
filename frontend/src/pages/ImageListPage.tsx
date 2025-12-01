import { useEffect, useState } from 'react'
import { fetchImages } from '../api/images'
import type { ImageMeta } from '../types'
import ImageCard from '../components/ImageCard'
import './ImageListPage.css'

const DEFAULT_FILTERS = {
  keyword: '',
  start_date: '',
  end_date: '',
}

const ImageListPage = () => {
  const [images, setImages] = useState<ImageMeta[]>([])
  const [filters, setFilters] = useState(DEFAULT_FILTERS)
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)

  const loadImages = async () => {
    setLoading(true)
    try {
      const data = await fetchImages({ ...filters, page: 1, pageSize: 40 })
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

  return (
    <div className="image-list-page">
      <section className="filter-panel">
        <div>
          <label>关键词</label>
          <input value={filters.keyword} onChange={(e) => handleChange('keyword', e.target.value)} placeholder="文件名" />
        </div>
        <div>
          <label>开始日期</label>
          <input type="date" value={filters.start_date} onChange={(e) => handleChange('start_date', e.target.value)} />
        </div>
        <div>
          <label>结束日期</label>
          <input type="date" value={filters.end_date} onChange={(e) => handleChange('end_date', e.target.value)} />
        </div>
        <button onClick={loadImages} disabled={loading}>
          {loading ? '查询中...' : '查询'}
        </button>
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

