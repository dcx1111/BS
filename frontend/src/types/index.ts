export interface User {
  id: number
  username: string
  email: string
}

export interface Tag {
  id: number
  name: string
  color?: string
}

export interface Thumbnail {
  imageId: number
  width: number
  height: number
}

export interface ImageMeta {
  id: number
  originalFilename: string
  storedFilename: string
  filePath: string
  mimeType: string
  fileSize: number
  width: number
  height: number
  createdAt: string
  tags?: Tag[]
  thumbnail?: Thumbnail
  exif?: {
    cameraMake?: string
    cameraModel?: string
    takenAt?: string
    locationName?: string
  }
}

export interface PaginatedResponse<T> {
  total: number
  page: number
  pageSize: number
  items: T[]
}

export interface AuthResponse {
  token: string
  user: User
}

