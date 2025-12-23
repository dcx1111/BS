import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useSlideshowStore, type SlideshowItem } from '../store/slideshowStore'
import './SlideshowEditPage.css'

const SlideshowEditPage = () => {
  const navigate = useNavigate()
  const items = useSlideshowStore((state) => state.items)
  const updateOrder = useSlideshowStore((state) => state.updateOrder)
  const updateDuration = useSlideshowStore((state) => state.updateDuration)
  const removeImage = useSlideshowStore((state) => state.removeImage)
  const clear = useSlideshowStore((state) => state.clear)
  const [localItems, setLocalItems] = useState<SlideshowItem[]>(items)

  const handleDurationChange = (imageId: number, duration: number) => {
    const newItems = localItems.map((item) =>
      item.imageId === imageId ? { ...item, duration } : item
    )
    setLocalItems(newItems)
    updateDuration(imageId, duration)
  }

  const handleMoveUp = (index: number) => {
    if (index === 0) return
    const newItems = [...localItems]
    ;[newItems[index - 1], newItems[index]] = [newItems[index], newItems[index - 1]]
    setLocalItems(newItems)
    updateOrder(newItems)
  }

  const handleMoveDown = (index: number) => {
    if (index === localItems.length - 1) return
    const newItems = [...localItems]
    ;[newItems[index], newItems[index + 1]] = [newItems[index + 1], newItems[index]]
    setLocalItems(newItems)
    updateOrder(newItems)
  }

  const handleRemove = (imageId: number) => {
    const newItems = localItems.filter((item) => item.imageId !== imageId)
    setLocalItems(newItems)
    removeImage(imageId)
  }

  const handleSave = () => {
    navigate('/slideshow')
  }

  const handleClear = () => {
    if (confirm('确定清空所有图片吗？')) {
      clear()
      setLocalItems([])
      navigate('/images')
    }
  }

  if (localItems.length === 0) {
    return (
      <div className="slideshow-edit-page">
        <div className="edit-header">
          <h2>编辑轮播组</h2>
          <button onClick={() => navigate('/images')} className="back-btn">
            返回
          </button>
        </div>
        <div className="empty-state">
          <p>轮播组为空，去添加一些图片吧</p>
          <button onClick={() => navigate('/images')} className="goto-images-btn">
            前往图片列表
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="slideshow-edit-page">
      <div className="edit-header">
        <h2>编辑轮播组</h2>
        <div className="header-actions">
          <button onClick={handleSave} className="save-btn">
            保存并播放
          </button>
          <button onClick={handleClear} className="clear-btn">
            清空
          </button>
          <button onClick={() => navigate('/images')} className="back-btn">
            返回
          </button>
        </div>
      </div>

      <div className="edit-content">
        <div className="items-list">
          {localItems.map((item, index) => {
            const thumbnailUrl = `${import.meta.env.VITE_API_BASE_URL ?? '/api/v1'}/images/${item.imageId}/thumbnail`
            return (
              <div key={item.imageId} className="edit-item">
                <div className="item-thumbnail">
                  <img src={thumbnailUrl} alt={item.image.originalFilename} />
                  <span className="item-index">{index + 1}</span>
                </div>
                <div className="item-info">
                  <h4>{item.image.originalFilename}</h4>
                  <div className="item-controls">
                    <div className="duration-control">
                      <label>显示时长（秒）</label>
                      <input
                        type="number"
                        min="1"
                        max="60"
                        value={item.duration}
                        onChange={(e) =>
                          handleDurationChange(item.imageId, parseInt(e.target.value) || 1)
                        }
                      />
                    </div>
                    <div className="order-controls">
                      <button
                        onClick={() => handleMoveUp(index)}
                        disabled={index === 0}
                        className="move-btn"
                      >
                        ↑
                      </button>
                      <button
                        onClick={() => handleMoveDown(index)}
                        disabled={index === localItems.length - 1}
                        className="move-btn"
                      >
                        ↓
                      </button>
                      <button onClick={() => handleRemove(item.imageId)} className="remove-btn">
                        删除
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}

export default SlideshowEditPage

