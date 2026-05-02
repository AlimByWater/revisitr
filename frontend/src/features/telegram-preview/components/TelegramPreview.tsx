import type { MessageContent } from '../types'
import { InteractivePreview } from './InteractivePreview'

export interface TelegramPreviewProps {
  botName: string
  botAvatar?: string
  content: MessageContent
  className?: string
  /** @deprecated Frame is always shown now */
  showFrame?: boolean
}

export function TelegramPreview({
  botName,
  botAvatar,
  content,
  className,
}: TelegramPreviewProps) {
  return (
    <InteractivePreview
      initialContent={content}
      botName={botName}
      botAvatar={botAvatar}
      className={className}
    />
  )
}
