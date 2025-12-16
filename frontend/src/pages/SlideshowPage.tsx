import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useSlideshowStore } from '../store/slideshowStore'
import './SlideshowPage.css'

const SlideshowPage = () => {
  const navigate = useNavigate()
  const items = useSlideshowStore((state) => state.items)
  const [currentIndex, setCurrentIndex] = useState(0)
  const [isPlaying, setIsPlaying] = useState(true)

  useEffect(() => {
    if (items.length === 0) {
      navigate('/images')
      return
    }
    // 如果当前索引超出范围，重置为0
    if (currentIndex >= items.length) {
      setCurrentIndex(0)
    }
  }, [items.length, navigate, currentIndex])

  useEffect(() => {
    if (!isPlaying || items.length === 0) return

    const currentItem = items[currentIndex]
    if (!currentItem) return

    const timer = setTimeout(() => {
      setCurrentIndex((prev) => (prev + 1) % items.length)
    }, currentItem.duration * 1000)

    return () => clearTimeout(timer)
  }, [currentIndex, items, isPlaying])

  const handlePrevious = () => {
    setCurrentIndex((prev) => (prev - 1 + items.length) % items.length)
  }

  const handleNext = () => {
    setCurrentIndex((prev) => (prev + 1) % items.length)
  }

  const handleTogglePlay = () => {
    setIsPlaying((prev) => !prev)
  }

  const handleGoToEdit = () => {
    navigate('/slideshow/edit')
  }

  if (items.length === 0) {
    return null
  }

  const currentItem = items[currentIndex]
  const originalUrl = `${import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api/v1'}/images/${currentItem.imageId}/original`

  return (
    <div className="slideshow-page">
      <div className="slideshow-container">
        <img src={originalUrl} alt={currentItem.image.originalFilename} className="slideshow-image" />
        
        <div className="slideshow-overlay">
          <div className="slideshow-info">
            <h2>{currentItem.image.originalFilename}</h2>
            <p>
              {currentIndex + 1} / {items.length}
            </p>
          </div>
        </div>

        <div className="slideshow-controls">
          <button onClick={handlePrevious} className="control-btn">‹</button>
          <button onClick={handleTogglePlay} className="control-btn play-btn">
            {isPlaying ? '⏸' : '▶'}
          </button>
          <button onClick={handleNext} className="control-btn">›</button>
        </div>

        <div className="slideshow-progress">
          {items.map((_, index) => (
            <div
              key={index}
              className={`progress-dot ${index === currentIndex ? 'active' : ''}`}
              onClick={() => setCurrentIndex(index)}
            />
          ))}
        </div>
      </div>

      <div className="slideshow-actions">
        <button onClick={handleGoToEdit} className="edit-btn">
          编辑轮播组
        </button>
        <button onClick={() => navigate('/images')} className="back-btn">
          返回图片列表
        </button>
      </div>
    </div>
  )
}

export default SlideshowPage

