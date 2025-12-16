import { useState } from 'react'
import { uploadImage } from '../api/images'
import './UploadPage.css'

const UploadPage = () => {
  const [file, setFile] = useState<File | null>(null)
  const [tags, setTags] = useState('')
  const [message, setMessage] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault()
    if (!file) {
      setMessage('请选择文件')
      return
    }
    setLoading(true)
    setMessage(null)
    try {
      const payloadTags = tags
        .split(',')
        .map((tag) => tag.trim())
        .filter(Boolean)
      await uploadImage(file, payloadTags)
      setMessage('上传成功')
      setFile(null)
      setTags('')
    } catch (err: any) {
      setMessage(err.response?.data?.message ?? '上传失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form className="upload-card" onSubmit={handleSubmit}>
      <h2>上传图片</h2>
      {message && <div className="upload-message">{message}</div>}

      <label className="file-drop">
        <input
          type="file"
          accept="image/*"
          onChange={(e) => setFile(e.target.files?.[0] ?? null)}
        />
        {file ? (
          <span>{file.name}</span>
        ) : (
          <span>点击或拖拽图片至此（最大10MB，支持 JPEG, PNG, GIF, BMP, TIFF, WebP）</span>
        )}
      </label>

      <label>
        自定义标签（用逗号分隔）
        <input
          type="text"
          value={tags}
          onChange={(e) => setTags(e.target.value)}
          placeholder="例：旅行, 北京, 朋友"
        />
      </label>

      <button type="submit" disabled={loading}>
        {loading ? '上传中...' : '上传'}
      </button>
    </form>
  )
}

export default UploadPage

