import { create } from 'zustand'

interface ImageListState {
  hasNewImages: boolean
  setHasNewImages: (value: boolean) => void
}

export const useImageListStore = create<ImageListState>((set) => ({
  hasNewImages: false,
  setHasNewImages: (value) => set({ hasNewImages: value }),
}))

