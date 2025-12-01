import { Link } from 'react-router-dom'
import { format } from 'date-fns'
import type { ImageMeta } from '../types'
import './ImageCard.css'

interface Props {
  image: ImageMeta
}

const ImageCard = ({ image }: Props) => {
  const thumbnailUrl = `${import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api/v1'}/images/${image.id}/thumbnail`

  return (
    <Link to={`/images/${image.id}`} className="image-card">
      <img src={thumbnailUrl} alt={image.originalFilename} loading="lazy" />
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

