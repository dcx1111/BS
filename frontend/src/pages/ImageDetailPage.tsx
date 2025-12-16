import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { deleteImage, fetchImageDetail, uploadImage, addImageTag, updateImageTag, removeImageTag } from '../api/images'
import type { ImageMeta, Tag } from '../types'
import ImageEditor from '../components/ImageEditor'
import './ImageDetailPage.css'

const ImageDetailPage = () => {
  const { id } = useParams()
  const navigate = useNavigate()
  const [image, setImage] = useState<ImageMeta | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isEditing, setIsEditing] = useState(false)
  const [saving, setSaving] = useState(false)
  const [editingTagId, setEditingTagId] = useState<number | null>(null)
  const [editingTagName, setEditingTagName] = useState('')
  const [newTagName, setNewTagName] = useState('')
  const [tagMessage, setTagMessage] = useState<string | null>(null)

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

  const handleAddTag = async () => {
    if (!id || !newTagName.trim()) return
    try {
      await addImageTag(id, newTagName.trim())
      setNewTagName('')
      setTagMessage('标签添加成功')
      await loadDetail()
      setTimeout(() => setTagMessage(null), 2000)
    } catch (err: any) {
      setTagMessage(err.response?.data?.message ?? '添加失败')
      setTimeout(() => setTagMessage(null), 2000)
    }
  }

  const handleStartEditTag = (tag: Tag) => {
    setEditingTagId(tag.id)
    setEditingTagName(tag.name)
  }

  const handleSaveEditTag = async () => {
    if (!id || !editingTagId || !editingTagName.trim()) return
    try {
      await updateImageTag(id, editingTagId, editingTagName.trim())
      setEditingTagId(null)
      setEditingTagName('')
      setTagMessage('标签修改成功')
      await loadDetail()
      setTimeout(() => setTagMessage(null), 2000)
    } catch (err: any) {
      setTagMessage(err.response?.data?.message ?? '修改失败')
      setTimeout(() => setTagMessage(null), 2000)
    }
  }

  const handleCancelEditTag = () => {
    setEditingTagId(null)
    setEditingTagName('')
  }

  const handleDeleteTag = async (tagId: number) => {
    if (!id) return
    if (!confirm('确定删除该标签吗？')) return
    try {
      await removeImageTag(id, tagId.toString())
      setTagMessage('标签删除成功')
      await loadDetail()
      setTimeout(() => setTagMessage(null), 2000)
    } catch (err: any) {
      setTagMessage(err.response?.data?.message ?? '删除失败')
      setTimeout(() => setTagMessage(null), 2000)
    }
  }

  const handleEdit = () => {
    setIsEditing(true)
  }

  const handleCancelEdit = () => {
    setIsEditing(false)
  }

  const handleSaveEdit = async (editedImageBlob: Blob) => {
    if (!id || !image) return
    setSaving(true)
    try {
      // 创建新文件名，添加edited前缀
      const originalName = image.originalFilename
      const nameWithoutExt = originalName.substring(0, originalName.lastIndexOf('.')) || originalName
      const ext = originalName.substring(originalName.lastIndexOf('.')) || '.jpg'
      const newFileName = `${nameWithoutExt}_edited${ext}`
      
      const file = new File([editedImageBlob], newFileName, {
        type: 'image/jpeg',
      })
      
      // 继承原图的标签
      const tagNames = image.tags?.map(tag => tag.name) || []
      
      // 上传为新图片
      const newImage = await uploadImage(file, tagNames)
      setIsEditing(false)
      
      // 跳转到新图片的详情页
      navigate(`/images/${newImage.id}`)
    } catch (err: any) {
      alert(err.response?.data?.message ?? '保存失败')
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return <div className="detail-card">加载中...</div>
  }

  if (error || !image) {
    return <div className="detail-card">{error ?? '图片不存在'}</div>
  }

  const originalUrl = `${import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api/v1'}/images/${image.id}/original`

  if (isEditing) {
    return (
      <ImageEditor
        imageUrl={originalUrl}
        onSave={handleSaveEdit}
        onCancel={handleCancelEdit}
      />
    )
  }

  return (
    <div className="detail-card">
      <header>
        <div>
          <h2>{image.originalFilename}</h2>
          <p>{new Date(image.createdAt).toLocaleString()}</p>
        </div>
        <div className="header-actions">
          <button onClick={handleEdit} className="btn-edit">编辑</button>
          <button onClick={handleDelete} className="btn-delete">删除</button>
        </div>
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
          {tagMessage && <div className="tag-message">{tagMessage}</div>}
          
          <div className="tag-list">
            {image.tags?.map((tag) => (
              <div key={tag.id} className="tag-item">
                {editingTagId === tag.id ? (
                  <div className="tag-edit">
                    <input
                      type="text"
                      value={editingTagName}
                      onChange={(e) => setEditingTagName(e.target.value)}
                      className="tag-edit-input"
                      autoFocus
                    />
                    <button onClick={handleSaveEditTag} className="btn-save-small">保存</button>
                    <button onClick={handleCancelEditTag} className="btn-cancel-small">取消</button>
                  </div>
                ) : (
                  <div className="tag-display">
                    <span className="tag-pill" style={{ backgroundColor: tag.color ?? '#dbeafe' }}>
                      {tag.name}
                    </span>
                    <button onClick={() => handleStartEditTag(tag)} className="btn-edit-small">修改</button>
                    <button onClick={() => handleDeleteTag(tag.id)} className="btn-delete-small">删除</button>
                  </div>
                )}
              </div>
            ))}
            {(!image.tags || image.tags.length === 0) && <p className="no-tags">暂无标签</p>}
          </div>

          <div className="tag-add">
            <input
              type="text"
              placeholder="输入新标签名称"
              value={newTagName}
              onChange={(e) => setNewTagName(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && handleAddTag()}
              className="tag-add-input"
            />
            <button onClick={handleAddTag} className="btn-add-tag">添加</button>
          </div>
        </section>
      </div>
    </div>
  )
}

export default ImageDetailPage

