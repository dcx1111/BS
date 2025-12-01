import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { deleteImage, fetchImageDetail } from '../api/images'
import type { ImageMeta } from '../types'
import './ImageDetailPage.css'

const ImageDetailPage = () => {
  const { id } = useParams()
  const navigate = useNavigate()
  const [image, setImage] = useState<ImageMeta | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const loadDetail = async () => {
    if (!id) return
    setLoading(true)
    setError(null)
    try {
      const data = await fetchImageDetail(id)
      setImage(data)
    } catch (err: any) {
      setError(err.response?.data?.message ?? '加载失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadDetail()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  const handleDelete = async () => {
    if (!id) return
    if (!confirm('确定删除该图片吗？')) return
    await deleteImage(id)
    navigate('/images')
  }

  if (loading) {
    return <div className="detail-card">加载中...</div>
  }

  if (error || !image) {
    return <div className="detail-card">{error ?? '图片不存在'}</div>
  }

  const originalUrl = `${import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api/v1'}/images/${image.id}/original`

  return (
    <div className="detail-card">
      <header>
        <div>
          <h2>{image.originalFilename}</h2>
          <p>{new Date(image.createdAt).toLocaleString()}</p>
        </div>
        <button onClick={handleDelete}>删除</button>
      </header>

      <div className="detail-content">
        <img src={originalUrl} alt={image.originalFilename} />
        <section className="meta-panel">
          <h3>基本信息</h3>
          <ul>
            <li>分辨率：{image.width} x {image.height}</li>
            <li>文件大小：{(image.fileSize / 1024 / 1024).toFixed(2)} MB</li>
            {image.exif?.cameraModel && <li>相机：{image.exif.cameraModel}</li>}
            {image.exif?.takenAt && <li>拍摄时间：{new Date(image.exif.takenAt).toLocaleString()}</li>}
            {image.exif?.locationName && <li>地点：{image.exif.locationName}</li>}
          </ul>

          <h3>标签</h3>
          <div className="tag-list">
            {image.tags?.map((tag) => (
              <span key={tag.id} className="tag-pill" style={{ backgroundColor: tag.color ?? '#dbeafe' }}>
                {tag.name}
              </span>
            )) ?? '暂无标签'}
          </div>
        </section>
      </div>
    </div>
  )
}

export default ImageDetailPage

