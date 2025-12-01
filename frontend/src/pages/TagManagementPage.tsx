import { useEffect, useState } from 'react'
import { createTag, fetchTags } from '../api/tags'
import type { Tag } from '../types'
import './TagManagementPage.css'

const TagManagementPage = () => {
  const [tags, setTags] = useState<Tag[]>([])
  const [name, setName] = useState('')
  const [color, setColor] = useState('#38bdf8')
  const [message, setMessage] = useState<string | null>(null)

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
    await createTag({ name: name.trim(), color })
    setName('')
    setMessage('创建成功')
    loadTags()
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
          <span key={tag.id} className="tag-pill" style={{ backgroundColor: tag.color ?? '#c7d2fe' }}>
            {tag.name}
          </span>
        ))}
        {tags.length === 0 && <p>还没有标签，先创建一个吧。</p>}
      </div>
    </div>
  )
}

export default TagManagementPage

