export type MessagePartType =
  | 'text'
  | 'photo'
  | 'video'
  | 'document'
  | 'animation'
  | 'sticker'
  | 'audio'
  | 'voice'

export interface MessagePart {
  type: MessagePartType
  text?: string
  media_url?: string
  media_id?: string
  parse_mode?: 'Markdown' | 'HTML' | ''
}

export interface InlineButton {
  text: string
  url?: string
  data?: string
  style?: 'danger' | 'success' | 'primary' | ''
  icon_custom_emoji_id?: string
  icon_image_url?: string // admin-only: preview of selected emoji
}

export interface MessageContent {
  parts: MessagePart[]
  buttons?: InlineButton[][]
}
