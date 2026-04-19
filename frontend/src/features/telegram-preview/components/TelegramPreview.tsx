import { cn } from '@/lib/utils'
import type { MessageContent } from '../types'
import { TelegramHeader } from './TelegramHeader'
import { MessageBubble } from './MessageBubble'
import { MediaMessage } from './MediaMessage'
import { StickerMessage } from './StickerMessage'
import { InlineKeyboard } from './InlineKeyboard'
import { PhoneFrame } from './PhoneFrame'
import { renderTextWithEmoji } from './renderEmoji'
import '../styles/telegram.css'

export interface TelegramPreviewProps {
  botName: string
  botAvatar?: string
  content: MessageContent
  className?: string
  showFrame?: boolean
  theme?: 'light' | 'dark'
}

export function TelegramPreview({
  botName,
  botAvatar,
  content,
  className,
  showFrame,
  theme = 'light',
}: TelegramPreviewProps) {
  const isDark = theme === 'dark'

  const chatContent = (
    <div className="flex flex-col h-full">
      <TelegramHeader botName={botName} botAvatar={botAvatar} theme={theme} />
      <div
        className={cn(
          'flex-1 overflow-y-auto p-3 space-y-1.5',
          isDark ? 'tg-chat-bg-dark' : 'tg-chat-bg'
        )}
      >
        {content.parts.map((part, i) => {
          const isLast = i === content.parts.length - 1
          const buttons = isLast ? content.buttons : undefined

          if (part.type === 'sticker') {
            return (
              <div key={i}>
                <StickerMessage mediaUrl={part.media_url} />
                {buttons && buttons.length > 0 && (
                  <InlineKeyboard buttons={buttons} theme={theme} />
                )}
              </div>
            )
          }

          if (
            part.type === 'photo' ||
            part.type === 'video' ||
            part.type === 'animation'
          ) {
            return (
              <div key={i}>
                <MediaMessage
                  type={part.type}
                  mediaUrl={part.media_url}
                  caption={part.text}
                  theme={theme}
                  showTail={i === 0}
                />
                {buttons && buttons.length > 0 && (
                  <InlineKeyboard buttons={buttons} theme={theme} />
                )}
              </div>
            )
          }

          if (part.type === 'document' || part.type === 'audio' || part.type === 'voice') {
            return (
              <div key={i}>
                <MessageBubble
                  theme={theme}
                  showTail={i === 0}
                >
                  <div className="tg-document">
                    <div className="tg-document-icon">
                      <svg width="18" height="22" viewBox="0 0 18 22" fill="none">
                        <path d="M2 0C0.9 0 0 0.9 0 2v18c0 1.1 0.9 2 2 2h14c1.1 0 2-0.9 2-2V6l-6-6H2z" fill="white"/>
                        <path d="M12 0v6h6" fill="white" fillOpacity="0.5"/>
                      </svg>
                    </div>
                    <div>
                      <div className="tg-document-name">
                        {part.media_url?.split('/').pop() || 'Document'}
                      </div>
                    </div>
                  </div>
                  {part.text && <div className="tg-bubble-caption">{renderTextWithEmoji(part.text)}</div>}
                </MessageBubble>
                {buttons && buttons.length > 0 && (
                  <InlineKeyboard buttons={buttons} theme={theme} />
                )}
              </div>
            )
          }

          // Default: text
          return (
            <div key={i}>
              <MessageBubble theme={theme} showTail={i === 0}>
                {part.text}
              </MessageBubble>
              {buttons && buttons.length > 0 && (
                <InlineKeyboard buttons={buttons} theme={theme} />
              )}
            </div>
          )
        })}

        {content.parts.length === 0 && (
          <div className="flex items-center justify-center h-32 text-sm text-gray-400 italic">
            Нет сообщений
          </div>
        )}
      </div>
    </div>
  )

  if (showFrame) {
    return (
      <PhoneFrame className={className}>
        <div className="flex h-[min(520px,65vh)] w-full flex-col sm:h-[520px]">{chatContent}</div>
      </PhoneFrame>
    )
  }

  return (
    <div
      className={cn(
        'w-full max-w-[360px] rounded-xl overflow-hidden border border-gray-200',
        className
      )}
    >
      <div className="flex h-[min(420px,55vh)] w-full flex-col sm:h-[420px]">{chatContent}</div>
    </div>
  )
}
