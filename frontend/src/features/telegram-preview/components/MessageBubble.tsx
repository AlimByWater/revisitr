import type { ReactNode } from 'react'
import { renderChildren } from './renderEmoji'
import { PREVIEW_MESSAGE_TIME } from './previewConstants'
import tailIncoming from '../assets/tail-incoming.svg'
import tailOutgoing from '../assets/tail-outgoing.svg'

interface MessageBubbleProps {
  children: ReactNode
  direction?: 'incoming' | 'outgoing'
  showTail?: boolean
  containerClassName?: string
  bubbleClassName?: string
}

export function MessageBubble({
  children,
  direction = 'incoming',
  showTail = false,
  containerClassName,
  bubbleClassName,
}: MessageBubbleProps) {
  const isOutgoing = direction === 'outgoing'
  const defaultContainerPadding = isOutgoing ? 'pr-[4px]' : 'pl-[4px]'

  return (
    <div
      className={`flex ${isOutgoing ? 'justify-end' : 'justify-start'} ${containerClassName ?? defaultContainerPadding}`}
    >
      <div
        className={`relative max-w-[312px] px-[8px] py-[4px] pb-[18px] ${bubbleClassName ?? ''}`}
        style={{
          background: isOutgoing ? '#EFFEDD' : '#FFFFFF',
          borderRadius: isOutgoing
            ? '13px 13px 4px 13px'
            : '13px 13px 13px 4px',
        }}
      >
        <div
          className="whitespace-pre-wrap text-[13px] font-normal leading-[15px] tracking-[-0.16px] text-black"
          style={{ wordBreak: 'break-word' }}
        >
          {renderChildren(children)}
        </div>
        <span className="absolute bottom-[6px] right-[8px] text-[10px] leading-none text-black/40">
          {PREVIEW_MESSAGE_TIME}
        </span>

        {/* Tail */}
        {showTail && isOutgoing && (
          <img
            src={tailOutgoing}
            alt=""
            className="tg-tail-outgoing"
            draggable={false}
          />
        )}
      </div>
    </div>
  )
}
