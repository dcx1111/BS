import { useEffect, useState, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { useSlideshowStore } from '../store/slideshowStore'
import './SlideshowPage.css'

const SlideshowPage = () => {
  const navigate = useNavigate()
  const items = useSlideshowStore((state) => state.items)
  const [currentIndex, setCurrentIndex] = useState(0)
  const [isPlaying, setIsPlaying] = useState(true)
  const [showControlsMobile, setShowControlsMobile] = useState(false)
  const controlsTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

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

  // 手机端按钮显示2秒后自动隐藏
  useEffect(() => {
    // 清理之前的定时器
    if (controlsTimerRef.current) {
      clearTimeout(controlsTimerRef.current)
      controlsTimerRef.current = null
    }

    if (showControlsMobile) {
      controlsTimerRef.current = setTimeout(() => {
        setShowControlsMobile(false)
        controlsTimerRef.current = null
      }, 2000)
    }

    return () => {
      if (controlsTimerRef.current) {
        clearTimeout(controlsTimerRef.current)
        controlsTimerRef.current = null
      }
    }
  }, [showControlsMobile])

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

  const handleImageClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    // 如果按钮已经显示，点击图片则隐藏；否则显示按钮
    setShowControlsMobile((prev) => {
      // 如果之前是显示状态，现在隐藏；否则显示
      const newValue = !prev
      // 如果隐藏，清除定时器
      if (!newValue && controlsTimerRef.current) {
        clearTimeout(controlsTimerRef.current)
        controlsTimerRef.current = null
      }
      return newValue
    })
  }

  const handlePageClick = (e: React.MouseEvent) => {
    // 如果点击的不是按钮本身，隐藏控制按钮
    const target = e.target as HTMLElement
    if (!target.closest('.control-btn')) {
      setShowControlsMobile(false)
    }
  }

  if (items.length === 0) {
    return null
  }

  const currentItem = items[currentIndex]
  const originalUrl = `${import.meta.env.VITE_API_BASE_URL ?? '/api/v1'}/images/${currentItem.imageId}/original`

  return (
    <div className="slideshow-page" onClick={handlePageClick}>
      <div className="slideshow-wrapper">
        <div className="slideshow-title-mobile">
          <h2>{currentItem.image.originalFilename}</h2>
          <p>
            {currentIndex + 1} / {items.length}
          </p>
        </div>
        <div className="slideshow-container">
          <img 
            src={originalUrl} 
            alt={currentItem.image.originalFilename} 
            className="slideshow-image"
            onClick={handleImageClick}
          />
          
          <div className="slideshow-overlay">
            <div className="slideshow-info">
              <h2>{currentItem.image.originalFilename}</h2>
              <p>
                {currentIndex + 1} / {items.length}
              </p>
            </div>
          </div>

        <div className={`slideshow-controls ${showControlsMobile ? 'show-mobile' : ''}`}>
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

