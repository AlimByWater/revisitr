import { type ReactNode, Fragment } from 'react'

const EMOJI_RE = /\{\{emoji:([^}]+)\}\}/g

export function renderTextWithEmoji(text: string): ReactNode[] {
  const parts: ReactNode[] = []
  let lastIndex = 0
  let match: RegExpExecArray | null

  EMOJI_RE.lastIndex = 0
  while ((match = EMOJI_RE.exec(text)) !== null) {
    if (match.index > lastIndex) {
      parts.push(<Fragment key={`t-${lastIndex}`}>{text.slice(lastIndex, match.index)}</Fragment>)
    }
    parts.push(
      <img
        key={`e-${match.index}`}
        src={match[1]}
        alt=""
        className="inline-block w-5 h-5 align-text-bottom rounded-sm object-cover"
      />
    )
    lastIndex = match.index + match[0].length
  }

  if (lastIndex < text.length) {
    parts.push(<Fragment key={`t-${lastIndex}`}>{text.slice(lastIndex)}</Fragment>)
  }

  return parts
}

export function renderChildren(children: ReactNode): ReactNode {
  if (typeof children === 'string') {
    return renderTextWithEmoji(children)
  }
  return children
}
