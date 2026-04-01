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
}

export interface MessageContent {
  parts: MessagePart[]
  buttons?: InlineButton[][]
}
