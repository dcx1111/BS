import { useState, useRef, useEffect } from 'react'
import './ImageEditor.css'

interface ImageEditorProps {
  imageUrl: string
  onSave: (editedImageBlob: Blob) => void
  onCancel: () => void
}

interface CropArea {
  x: number
  y: number
  width: number
  height: number
}

interface ImageFilters {
  brightness: number
  contrast: number
  saturation: number
  hue: number
}

const ImageEditor = ({ imageUrl, onSave, onCancel }: ImageEditorProps) => {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const imageRef = useRef<HTMLImageElement | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const [imageLoaded, setImageLoaded] = useState(false)
  const [mode, setMode] = useState<'crop' | 'adjust'>('crop')
  const [cropArea, setCropArea] = useState<CropArea | null>(null)
  const [isDragging, setIsDragging] = useState(false)
  const [isResizing, setIsResizing] = useState(false)
  const [resizeDirection, setResizeDirection] = useState<string>('')
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 })
  const [filters, setFilters] = useState<ImageFilters>({
    brightness: 0,
    contrast: 0,
    saturation: 0,
    hue: 0,
  })

  const [imageSize, setImageSize] = useState({ width: 0, height: 0 })
  const [displaySize, setDisplaySize] = useState({ width: 0, height: 0 })

  useEffect(() => {
    const img = new Image()
    img.crossOrigin = 'anonymous'
    img.onload = () => {
      imageRef.current = img
      setImageSize({ width: img.width, height: img.height })
      
      // 计算显示尺寸
      if (containerRef.current) {
        const containerWidth = containerRef.current.clientWidth - 40
        const scale = Math.min(containerWidth / img.width, 600 / img.height, 1)
        setDisplaySize({
          width: img.width * scale,
          height: img.height * scale,
        })
        // 初始化裁剪区域为整个图片
        setCropArea({
          x: 0,
          y: 0,
          width: img.width * scale,
          height: img.height * scale,
        })
      }
      setImageLoaded(true)
      drawImage()
    }
    img.src = imageUrl
  }, [imageUrl])

  useEffect(() => {
    if (imageLoaded) {
      drawImage()
    }
  }, [imageLoaded, filters, cropArea, mode])

  const drawImage = () => {
    const canvas = canvasRef.current
    if (!canvas || !imageRef.current) return

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    canvas.width = displaySize.width
    canvas.height = displaySize.height

    ctx.clearRect(0, 0, canvas.width, canvas.height)

    // 应用滤镜（始终应用，不管在哪个模式）
    const filterString = [
      `brightness(${100 + filters.brightness}%)`,
      `contrast(${100 + filters.contrast}%)`,
      `saturate(${100 + filters.saturation}%)`,
      `hue-rotate(${filters.hue}deg)`,
    ].join(' ')

    ctx.filter = filterString
    ctx.drawImage(imageRef.current, 0, 0, displaySize.width, displaySize.height)

    // 绘制裁剪框和遮罩（始终显示，不管在哪个模式）
    if (cropArea) {
      // 用半透明遮罩显示裁剪区域外的部分（两种模式都显示）
      // 使用路径排除裁剪区域，只在区域外绘制遮罩
      ctx.save()
      ctx.fillStyle = 'rgba(0, 0, 0, 0.5)'
      ctx.beginPath()
      // 绘制整个画布
      ctx.rect(0, 0, canvas.width, canvas.height)
      // 从路径中排除裁剪区域（使用逆时针方向）
      ctx.rect(cropArea.x, cropArea.y, cropArea.width, cropArea.height)
      ctx.fill('evenodd')
      ctx.restore()

      // 绘制裁剪框边框
      ctx.strokeStyle = '#3b82f6'
      ctx.lineWidth = mode === 'crop' ? 3 : 2
      ctx.setLineDash(mode === 'crop' ? [] : [5, 5])
      ctx.strokeRect(cropArea.x, cropArea.y, cropArea.width, cropArea.height)
    }
  }

  const getResizeDirection = (x: number, y: number): string => {
    if (!cropArea) return ''
    const threshold = 10
    let direction = ''

    // 检查是否在边缘
    if (Math.abs(x - cropArea.x) < threshold) direction += 'w'
    if (Math.abs(x - (cropArea.x + cropArea.width)) < threshold) direction += 'e'
    if (Math.abs(y - cropArea.y) < threshold) direction += 'n'
    if (Math.abs(y - (cropArea.y + cropArea.height)) < threshold) direction += 's'

    return direction
  }

  const handleMouseDown = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!cropArea) return

    // 在adjust模式下，不允许拖拽裁剪框
    if (mode === 'adjust') return

    const rect = canvasRef.current?.getBoundingClientRect()
    if (!rect) return

    const x = e.clientX - rect.left
    const y = e.clientY - rect.top

    // 检查是否在调整大小的边缘
    const resizeDir = getResizeDirection(x, y)
    if (resizeDir) {
      setIsResizing(true)
      setResizeDirection(resizeDir)
      setDragStart({ x, y })
      return
    }

    // 检查是否点击在裁剪框内
    const isInside =
      x >= cropArea.x &&
      x <= cropArea.x + cropArea.width &&
      y >= cropArea.y &&
      y <= cropArea.y + cropArea.height

    if (isInside) {
      setIsDragging(true)
      setDragStart({ x, y })
    }
  }

  const handleMouseMove = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!cropArea) return

    const rect = canvasRef.current?.getBoundingClientRect()
    if (!rect) return

    const x = e.clientX - rect.left
    const y = e.clientY - rect.top

    // 在adjust模式下，不允许拖拽裁剪框，只显示默认光标
    if (mode === 'adjust') {
      if (canvasRef.current) {
        canvasRef.current.style.cursor = 'default'
      }
      return
    }

    // 在crop模式下，允许拖拽和调整大小
    if (isResizing && resizeDirection) {
      const dx = x - dragStart.x
      const dy = y - dragStart.y

      let newCrop = { ...cropArea }

      if (resizeDirection.includes('n')) {
        const newY = Math.max(0, cropArea.y + dy)
        const newHeight = cropArea.height - (newY - cropArea.y)
        if (newHeight >= 20) {
          newCrop.y = newY
          newCrop.height = newHeight
        }
      }
      if (resizeDirection.includes('s')) {
        newCrop.height = Math.max(20, Math.min(cropArea.height + dy, displaySize.height - cropArea.y))
      }
      if (resizeDirection.includes('w')) {
        const newX = Math.max(0, cropArea.x + dx)
        const newWidth = cropArea.width - (newX - cropArea.x)
        if (newWidth >= 20) {
          newCrop.x = newX
          newCrop.width = newWidth
        }
      }
      if (resizeDirection.includes('e')) {
        newCrop.width = Math.max(20, Math.min(cropArea.width + dx, displaySize.width - cropArea.x))
      }

      setCropArea(newCrop)
      setDragStart({ x, y })
    } else if (isDragging) {
      const dx = x - dragStart.x
      const dy = y - dragStart.y

      setCropArea({
        x: Math.max(0, Math.min(cropArea.x + dx, displaySize.width - cropArea.width)),
        y: Math.max(0, Math.min(cropArea.y + dy, displaySize.height - cropArea.height)),
        width: cropArea.width,
        height: cropArea.height,
      })
      setDragStart({ x, y })
    } else {
      // 更新鼠标样式（仅在crop模式下）
      const resizeDir = getResizeDirection(x, y)
      if (resizeDir) {
        const cursorMap: Record<string, string> = {
          'n': 'n-resize',
          's': 's-resize',
          'w': 'w-resize',
          'e': 'e-resize',
          'nw': 'nw-resize',
          'ne': 'ne-resize',
          'sw': 'sw-resize',
          'se': 'se-resize',
        }
        if (canvasRef.current) {
          canvasRef.current.style.cursor = cursorMap[resizeDir] || 'default'
        }
      } else {
        const isInside =
          x >= cropArea.x &&
          x <= cropArea.x + cropArea.width &&
          y >= cropArea.y &&
          y <= cropArea.y + cropArea.height
        if (canvasRef.current) {
          canvasRef.current.style.cursor = isInside ? 'move' : 'default'
        }
      }
    }
  }

  const handleMouseUp = () => {
    setIsDragging(false)
    setIsResizing(false)
    setResizeDirection('')
    if (canvasRef.current) {
      canvasRef.current.style.cursor = 'default'
    }
  }

  const handleSave = async () => {
    if (!imageRef.current) return

    const canvas = document.createElement('canvas')
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    // 计算实际裁剪区域（相对于原始图片）
    const scaleX = imageSize.width / displaySize.width
    const scaleY = imageSize.height / displaySize.height

    // 始终应用裁剪（如果有设置的话）
    let sourceX = 0
    let sourceY = 0
    let sourceWidth = imageSize.width
    let sourceHeight = imageSize.height

    if (cropArea) {
      sourceX = Math.round(cropArea.x * scaleX)
      sourceY = Math.round(cropArea.y * scaleY)
      sourceWidth = Math.round(cropArea.width * scaleX)
      sourceHeight = Math.round(cropArea.height * scaleY)
    }

    canvas.width = sourceWidth
    canvas.height = sourceHeight

    // 始终应用滤镜
    const filterString = [
      `brightness(${100 + filters.brightness}%)`,
      `contrast(${100 + filters.contrast}%)`,
      `saturate(${100 + filters.saturation}%)`,
      `hue-rotate(${filters.hue}deg)`,
    ].join(' ')

    ctx.filter = filterString
    ctx.drawImage(
      imageRef.current,
      sourceX,
      sourceY,
      sourceWidth,
      sourceHeight,
      0,
      0,
      sourceWidth,
      sourceHeight
    )

    canvas.toBlob(
      (blob) => {
        if (blob) {
          onSave(blob)
        }
      },
      'image/jpeg',
      0.95
    )
  }

  const handleReset = () => {
    setFilters({
      brightness: 0,
      contrast: 0,
      saturation: 0,
      hue: 0,
    })
    if (imageRef.current && containerRef.current) {
      const containerWidth = containerRef.current.clientWidth - 40
      const scale = Math.min(containerWidth / imageRef.current.width, 600 / imageRef.current.height, 1)
      setCropArea({
        x: 0,
        y: 0,
        width: imageRef.current.width * scale,
        height: imageRef.current.height * scale,
      })
    }
  }

  return (
    <div className="image-editor">
      <div className="editor-header">
        <h3>图片编辑</h3>
        <div className="editor-modes">
          <button
            className={mode === 'crop' ? 'active' : ''}
            onClick={() => setMode('crop')}
          >
            裁剪
          </button>
          <button
            className={mode === 'adjust' ? 'active' : ''}
            onClick={() => setMode('adjust')}
          >
            色调调整
          </button>
        </div>
      </div>

      <div className="editor-content">
        <div className="canvas-container" ref={containerRef}>
          {imageLoaded ? (
            <canvas
              ref={canvasRef}
              onMouseDown={handleMouseDown}
              onMouseMove={handleMouseMove}
              onMouseUp={handleMouseUp}
              onMouseLeave={handleMouseUp}
            />
          ) : (
            <div className="loading">加载中...</div>
          )}
        </div>

        <div className="editor-controls">
          {mode === 'adjust' && (
            <div className="filter-controls">
              <div className="control-item">
                <label>亮度</label>
                <input
                  type="range"
                  min="-100"
                  max="100"
                  value={filters.brightness}
                  onChange={(e) =>
                    setFilters({ ...filters, brightness: Number(e.target.value) })
                  }
                />
                <span>{filters.brightness > 0 ? '+' : ''}{filters.brightness}</span>
              </div>

              <div className="control-item">
                <label>对比度</label>
                <input
                  type="range"
                  min="-100"
                  max="100"
                  value={filters.contrast}
                  onChange={(e) =>
                    setFilters({ ...filters, contrast: Number(e.target.value) })
                  }
                />
                <span>{filters.contrast > 0 ? '+' : ''}{filters.contrast}</span>
              </div>

              <div className="control-item">
                <label>饱和度</label>
                <input
                  type="range"
                  min="-100"
                  max="100"
                  value={filters.saturation}
                  onChange={(e) =>
                    setFilters({ ...filters, saturation: Number(e.target.value) })
                  }
                />
                <span>{filters.saturation > 0 ? '+' : ''}{filters.saturation}</span>
              </div>

              <div className="control-item">
                <label>色相</label>
                <input
                  type="range"
                  min="-180"
                  max="180"
                  value={filters.hue}
                  onChange={(e) =>
                    setFilters({ ...filters, hue: Number(e.target.value) })
                  }
                />
                <span>{filters.hue > 0 ? '+' : ''}{filters.hue}°</span>
              </div>
            </div>
          )}

          {mode === 'crop' && (
            <div className="crop-hint">
              <p>拖拽裁剪框来移动位置，拖拽边缘来调整大小</p>
            </div>
          )}
          {mode === 'adjust' && (
            <div className="crop-hint">
              <p>调整色调参数，裁剪区域内的部分会保留</p>
            </div>
          )}

          <div className="editor-actions">
            <button onClick={handleReset} className="btn-reset">
              重置
            </button>
            <button onClick={onCancel} className="btn-cancel">
              取消
            </button>
            <button onClick={handleSave} className="btn-save">
              保存
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

export default ImageEditor

