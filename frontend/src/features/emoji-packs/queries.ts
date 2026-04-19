import { useApiQuery, useApiMutation } from '@/lib/swr'
import { emojiPacksApi } from './api'
import type {
  CreateEmojiPackRequest,
  UpdateEmojiPackRequest,
  CreateEmojiItemRequest,
  UpdateEmojiItemRequest,
} from './types'

export function useEmojiPacksQuery() {
  return useApiQuery('emoji-packs', emojiPacksApi.list)
}

export function useEmojiPackQuery(id: number) {
  return useApiQuery(id ? `emoji-packs-${id}` : null, () => emojiPacksApi.getById(id))
}

export function useCreateEmojiPackMutation() {
  return useApiMutation(
    'emoji-packs/create',
    (data: CreateEmojiPackRequest) => emojiPacksApi.create(data),
    ['emoji-packs'],
  )
}

export function useUpdateEmojiPackMutation() {
  return useApiMutation(
    'emoji-packs/update',
    ({ id, data }: { id: number; data: UpdateEmojiPackRequest }) => emojiPacksApi.update(id, data),
    ['emoji-packs'],
  )
}

export function useDeleteEmojiPackMutation() {
  return useApiMutation(
    'emoji-packs/delete',
    (id: number) => emojiPacksApi.remove(id),
    ['emoji-packs'],
  )
}

export function useAddEmojiItemMutation(packId: number) {
  return useApiMutation(
    `emoji-packs/${packId}/add-item`,
    (data: CreateEmojiItemRequest) => emojiPacksApi.addItem(packId, data),
    [`emoji-packs-${packId}`, 'emoji-packs'],
  )
}

export function useUpdateEmojiItemMutation() {
  return useApiMutation(
    'emoji-packs/update-item',
    ({ itemId, data }: { itemId: number; data: UpdateEmojiItemRequest }) =>
      emojiPacksApi.updateItem(itemId, data),
    ['emoji-packs'],
  )
}

export function useDeleteEmojiItemMutation() {
  return useApiMutation(
    'emoji-packs/delete-item',
    (itemId: number) => emojiPacksApi.deleteItem(itemId),
    ['emoji-packs'],
  )
}

export function useReorderEmojiItemsMutation(packId: number) {
  return useApiMutation(
    `emoji-packs/${packId}/reorder`,
    (itemIds: number[]) => emojiPacksApi.reorderItems(packId, itemIds),
    [`emoji-packs-${packId}`, 'emoji-packs'],
  )
}
