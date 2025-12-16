import { useEffect, useState } from 'react'
import { createTag, fetchTags, updateTagColor } from '../api/tags'
import type { Tag } from '../types'
import './TagManagementPage.css'

const TagManagementPage = () => {
  const [tags, setTags] = useState<Tag[]>([])
  const [name, setName] = useState('')
  const [color, setColor] = useState('#38bdf8')
  const [message, setMessage] = useState<string | null>(null)
  const [editingTagId, setEditingTagId] = useState<number | null>(null)
  const [editingColor, setEditingColor] = useState('')

  const loadTags = async () => {
    const data = await fetchTags()
    setTags(data)
  }

  useEffect(() => {
    loadTags()
  }, [])

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault()
    if (!name.trim()) {
      setMessage('请输入标签名称')
      return
    }
    try {
      await createTag({ name: name.trim(), color })
      setName('')
      setColor('#38bdf8')
      setMessage('创建成功')
      loadTags()
    } catch (err: any) {
      setMessage(err.response?.data?.message ?? '创建失败')
    }
  }

  const handleStartEdit = (tag: Tag) => {
    setEditingTagId(tag.id)
    setEditingColor(tag.color || '#c7d2fe')
  }

  const handleSaveEdit = async (tagId: number) => {
    try {
      await updateTagColor(tagId, editingColor)
      setEditingTagId(null)
      setMessage('颜色更新成功')
      loadTags()
    } catch (err: any) {
      setMessage(err.response?.data?.message ?? '更新失败')
    }
  }

  const handleCancelEdit = () => {
    setEditingTagId(null)
    setEditingColor('')
  }

  return (
    <div className="tag-page">
      <form className="tag-form" onSubmit={handleSubmit}>
        <input
          type="text"
          placeholder="标签名"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <input type="color" value={color} onChange={(e) => setColor(e.target.value)} />
        <button type="submit">创建</button>
      </form>
      {message && <div className="tag-message">{message}</div>}
      <div className="tag-collection">
        {tags.map((tag) => (
          <div key={tag.id} className="tag-item">
            {editingTagId === tag.id ? (
              <div className="tag-edit">
                <span className="tag-pill" style={{ backgroundColor: editingColor || '#c7d2fe' }}>
                  {tag.name}
                </span>
                <input
                  type="color"
                  value={editingColor}
                  onChange={(e) => setEditingColor(e.target.value)}
                  className="color-input"
                />
                <button onClick={() => handleSaveEdit(tag.id)} className="btn-save">保存</button>
                <button onClick={handleCancelEdit} className="btn-cancel">取消</button>
              </div>
            ) : (
              <div className="tag-display">
                <span
                  className="tag-pill"
                  style={{ backgroundColor: tag.color || '#c7d2fe' }}
                >
                  {tag.name}
                </span>
                {!tag.color && <span className="no-color-badge">无色</span>}
                <button onClick={() => handleStartEdit(tag)} className="btn-edit">
                  编辑颜色
                </button>
              </div>
            )}
          </div>
        ))}
        {tags.length === 0 && <p>还没有标签，先创建一个吧。</p>}
      </div>
    </div>
  )
}

export default TagManagementPage

