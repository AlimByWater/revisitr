import { api } from '@/lib/api'
import type {
  EmojiPack,
  EmojiItem,
  CreateEmojiPackRequest,
  UpdateEmojiPackRequest,
  CreateEmojiItemRequest,
  UpdateEmojiItemRequest,
} from './types'

export const emojiPacksApi = {
  list: async (): Promise<EmojiPack[]> => {
    const response = await api.get<EmojiPack[]>('/emoji-packs')
    return response.data
  },

  getById: async (id: number): Promise<EmojiPack> => {
    const response = await api.get<EmojiPack>(`/emoji-packs/${id}`)
    return response.data
  },

  create: async (data: CreateEmojiPackRequest): Promise<EmojiPack> => {
    const response = await api.post<EmojiPack>('/emoji-packs', data)
    return response.data
  },

  update: async (id: number, data: UpdateEmojiPackRequest): Promise<void> => {
    await api.patch(`/emoji-packs/${id}`, data)
  },

  remove: async (id: number): Promise<void> => {
    await api.delete(`/emoji-packs/${id}`)
  },

  addItem: async (packId: number, data: CreateEmojiItemRequest): Promise<EmojiItem> => {
    const response = await api.post<EmojiItem>(`/emoji-packs/${packId}/items`, data)
    return response.data
  },

  updateItem: async (itemId: number, data: UpdateEmojiItemRequest): Promise<void> => {
    await api.patch(`/emoji-packs/items/${itemId}`, data)
  },

  deleteItem: async (itemId: number): Promise<void> => {
    await api.delete(`/emoji-packs/items/${itemId}`)
  },

  reorderItems: async (packId: number, itemIds: number[]): Promise<void> => {
    await api.put(`/emoji-packs/${packId}/items/reorder`, { item_ids: itemIds })
  },
}
