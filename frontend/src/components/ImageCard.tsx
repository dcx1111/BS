import { Link } from 'react-router-dom'
import { format } from 'date-fns'
import type { ImageMeta } from '../types'
import { useSlideshowStore } from '../store/slideshowStore'
import './ImageCard.css'

interface Props {
  image: ImageMeta
}

const ImageCard = ({ image }: Props) => {
  const thumbnailUrl = `${import.meta.env.VITE_API_BASE_URL ?? '/api/v1'}/images/${image.id}/thumbnail`
  const addImage = useSlideshowStore((state) => state.addImage)
  const removeImage = useSlideshowStore((state) => state.removeImage)
  const items = useSlideshowStore((state) => state.items)
  const isInSlideshow = items.some((item) => item.imageId === image.id)

  const handleAddToSlideshow = (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (isInSlideshow) {
      removeImage(image.id)
    } else {
      addImage(image)
    }
  }

  return (
    <Link to={`/images/${image.id}`} className="image-card">
      <div className="image-card-image-wrapper">
        <img src={thumbnailUrl} alt={image.originalFilename} loading="lazy" />
        <button
          className={`slideshow-add-btn ${isInSlideshow ? 'added' : ''}`}
          onClick={handleAddToSlideshow}
          title={isInSlideshow ? '点击移出轮播组' : '添加到轮播组'}
        >
          {isInSlideshow ? '✓' : '+'}
        </button>
      </div>
      <div className="image-info">
        <h4 title={image.originalFilename}>{image.originalFilename}</h4>
        <p>{format(new Date(image.createdAt), 'yyyy-MM-dd HH:mm')}</p>
        <div className="tag-row">
          {image.tags?.slice(0, 3).map((tag) => (
            <span key={tag.id} className="tag-badge" style={{ backgroundColor: tag.color ?? '#e0e7ff' }}>
              {tag.name}
            </span>
          ))}
        </div>
      </div>
    </Link>
  )
}

export default ImageCard

