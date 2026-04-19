export interface EmojiPack {
  id: number
  org_id: number
  name: string
  sort_order: number
  created_at: string
  updated_at: string
  items?: EmojiItem[]
}

export interface EmojiItem {
  id: number
  pack_id: number
  name: string
  image_url: string
  sort_order: number
  tg_sticker_set?: string
  tg_custom_emoji_id?: string
  created_at: string
}

export interface CreateEmojiPackRequest {
  name: string
}

export interface UpdateEmojiPackRequest {
  name?: string
  sort_order?: number
}

export interface CreateEmojiItemRequest {
  name: string
  image_url: string
}

export interface UpdateEmojiItemRequest {
  name?: string
  sort_order?: number
}

export interface ReorderEmojiItemsRequest {
  item_ids: number[]
}
