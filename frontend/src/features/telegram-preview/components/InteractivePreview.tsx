import { useCallback, useEffect, useRef } from 'react'
import type { ReactNode } from 'react'
import type { InlineButton, MessageContent, PreviewScreen } from '../types'
import { useScreenStack } from '../hooks/useScreenStack'
import { PhoneFrame } from './PhoneFrame'
import { ChatHeader } from './ChatHeader'
import { MessageBubble } from './MessageBubble'
import { MediaMessage } from './MediaMessage'
import { StickerMessage } from './StickerMessage'
import { InlineKeyboard } from './InlineKeyboard'
import { Composer } from './Composer'
import { renderTextWithEmoji } from './renderEmoji'
import '../styles/telegram-preview.css'

export interface InteractivePreviewProps {
  initialContent: MessageContent
  botName: string
  botAvatar?: string
  onButtonClick?: (button: InlineButton) => PreviewScreen | null
  className?: string
}

export function InteractivePreview({
  initialContent,
  botName,
  botAvatar,
  onButtonClick,
  className,
}: InteractivePreviewProps) {
  const { stack, push, replace, reset, canReset } = useScreenStack(initialContent)
  const scrollRef = useRef<HTMLDivElement>(null)

  // Scroll to bottom when stack changes
  useEffect(() => {
    const el = scrollRef.current
    if (el) {
      el.scrollTop = el.scrollHeight
    }
  }, [stack])

  const handleButtonClick = useCallback(
    (button: InlineButton) => {
      if (!onButtonClick) return
      const nextScreen = onButtonClick(button)
      if (nextScreen) {
        if (nextScreen.transition === 'replace') {
          replace(nextScreen)
          return
        }
        push(nextScreen)
      }
    },
    [onButtonClick, push, replace],
  )

  return (
    <div className={`relative ${className ?? ''}`}>
      {/* Reset button */}
      {canReset && (
        <button
          type="button"
          onClick={reset}
          className="mb-2 flex items-center gap-1 rounded-lg px-2.5 py-1 text-xs font-medium text-neutral-500 transition-colors hover:bg-neutral-100 hover:text-neutral-700"
        >
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <path
              d="M2.5 7a4.5 4.5 0 0 1 8.3-2.4M11.5 7a4.5 4.5 0 0 1-8.3 2.4"
              stroke="currentColor"
              strokeWidth="1.3"
              strokeLinecap="round"
            />
            <path d="M10.5 2v2.6h-2.6" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round" strokeLinejoin="round" />
            <path d="M3.5 12v-2.6h2.6" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
          Сбросить
        </button>
      )}

      <PhoneFrame className={className}>
        {/* Header + messages + composer inside PhoneFrame's gradient */}
        <div className="relative flex flex-1 flex-col overflow-hidden">
          {/* Header — floats on top of gradient */}
          <ChatHeader botName={botName} botAvatar={botAvatar} />

          {/* Messages scroll area */}
          <div
            ref={scrollRef}
            className="flex flex-1 flex-col gap-[4px] overflow-y-auto px-[10px] pb-[4px] pt-[8px]"
          >
            <div className="mt-auto" />
            {stack.map((screen, screenIndex) => (
              <ScreenMessages
                key={screenIndex}
                screen={screen}
                isLast={screenIndex === stack.length - 1}
                onButtonClick={onButtonClick ? handleButtonClick : undefined}
              />
            ))}

            {stack.length === 1 && stack[0].content.parts.length === 0 && (
              <div className="flex flex-1 items-center justify-center text-sm italic text-white/60">
                Нет сообщений
              </div>
            )}
          </div>

          {/* Composer — floats at bottom of gradient */}
          <Composer />
        </div>
      </PhoneFrame>
    </div>
  )
}

function ScreenMessages({
  screen,
  isLast,
  onButtonClick,
}: {
  screen: PreviewScreen
  isLast: boolean
  onButtonClick?: (button: InlineButton) => void
}) {
  const { content, outgoingText } = screen

  return (
    <>
      {/* Outgoing bubble (user tap) */}
      {outgoingText && (
        <MessageBubble direction="outgoing" showTail>
          {outgoingText}
        </MessageBubble>
      )}

      {/* Bot message parts */}
      {content.parts.map((part, i) => {
        const showButtons = isLast && i === content.parts.length - 1
        const buttons = showButtons ? content.buttons : undefined

        if (part.type === 'sticker') {
          return (
            <MessageGroup key={i} buttons={buttons} onButtonClick={onButtonClick} wide>
              <StickerMessage mediaUrl={part.media_url} />
            </MessageGroup>
          )
        }

        if (part.type === 'photo' || part.type === 'video' || part.type === 'animation') {
          return (
            <MessageGroup key={i} buttons={buttons} onButtonClick={onButtonClick}>
              <MediaMessage
                type={part.type}
                mediaUrl={part.media_url}
                caption={part.text}
                showTail={i === 0 && !outgoingText}
                containerClassName="pl-0"
                mediaClassName="w-full max-w-none"
              />
            </MessageGroup>
          )
        }

        if (part.type === 'document' || part.type === 'audio' || part.type === 'voice') {
          return (
            <MessageGroup key={i} buttons={buttons} onButtonClick={onButtonClick}>
              <MessageBubble
                showTail={i === 0 && !outgoingText}
                containerClassName="pl-0"
                bubbleClassName="w-full max-w-none"
              >
                <div className="flex items-center gap-[8px] py-[2px]">
                  <div className="flex h-[34px] w-[34px] shrink-0 items-center justify-center rounded-lg bg-gradient-to-b from-[#5db3ff] to-[#2f7be8]">
                    <svg width="14" height="18" viewBox="0 0 18 22" fill="none">
                      <path d="M2 0C.9 0 0 .9 0 2v18c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V6l-6-6H2z" fill="white" />
                      <path d="M12 0v6h6" fill="white" fillOpacity="0.5" />
                    </svg>
                  </div>
                  <div>
                    <div className="text-[12px] font-semibold text-[#1877f2]">
                      {part.media_url?.split('/').pop() || 'Document'}
                    </div>
                  </div>
                </div>
                {part.text && (
                  <div className="mt-[4px]">{renderTextWithEmoji(part.text)}</div>
                )}
              </MessageBubble>
            </MessageGroup>
          )
        }

        // Text message
        return (
          <MessageGroup key={i} buttons={buttons} onButtonClick={onButtonClick}>
            <MessageBubble
              showTail={i === 0 && !outgoingText}
              containerClassName="pl-0"
              bubbleClassName="w-full max-w-none"
            >
              {part.text}
            </MessageBubble>
          </MessageGroup>
        )
      })}
    </>
  )
}

function MessageGroup({
  children,
  buttons,
  onButtonClick,
  wide = false,
}: {
  children: ReactNode
  buttons?: InlineButton[][]
  onButtonClick?: (button: InlineButton) => void
  wide?: boolean
}) {
  const fullWidth = Boolean(buttons && buttons.length > 0)
  return (
    <div className={`flex justify-start pl-0 pr-[8px] ${fullWidth || wide ? 'w-full' : 'max-w-[312px]'}`}>
      <div className={`${fullWidth || wide ? 'w-full' : 'w-fit max-w-[88%]'}`}>
        {children}
        {buttons && buttons.length > 0 && (
          <InlineKeyboard buttons={buttons} onButtonClick={onButtonClick} />
        )}
      </div>
    </div>
  )
}
