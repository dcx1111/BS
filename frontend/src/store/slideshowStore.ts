import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { ImageMeta } from '../types'

export interface SlideshowItem {
  imageId: number
  image: ImageMeta
  duration: number // 秒
}

interface SlideshowState {
  items: SlideshowItem[]
  addImage: (image: ImageMeta) => void
  removeImage: (imageId: number) => void
  updateOrder: (items: SlideshowItem[]) => void
  updateDuration: (imageId: number, duration: number) => void
  clear: () => void
}

export const useSlideshowStore = create<SlideshowState>()(
  persist(
    (set) => ({
      items: [],
      addImage: (image) =>
        set((state) => {
          // 检查是否已存在
          if (state.items.some((item) => item.imageId === image.id)) {
            return state
          }
          return {
            items: [
              ...state.items,
              {
                imageId: image.id,
                image,
                duration: 5, // 默认5秒
              },
            ],
          }
        }),
      removeImage: (imageId) =>
        set((state) => ({
          items: state.items.filter((item) => item.imageId !== imageId),
        })),
      updateOrder: (items) => set({ items }),
      updateDuration: (imageId, duration) =>
        set((state) => ({
          items: state.items.map((item) =>
            item.imageId === imageId ? { ...item, duration } : item
          ),
        })),
      clear: () => set({ items: [] }),
    }),
    {
      name: 'image-manager-slideshow',
    },
  ),
)

