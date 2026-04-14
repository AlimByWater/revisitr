import { Smile } from 'lucide-react'

interface StickerMessageProps {
  mediaUrl?: string
}

export function StickerMessage({ mediaUrl }: StickerMessageProps) {
  return (
    <div className="flex justify-start">
      <div className="tg-sticker">
        {mediaUrl ? (
          <img
            src={mediaUrl}
            alt="sticker"
            loading="lazy"
            decoding="async"
            onError={(e) => {
              e.currentTarget.style.display = 'none'
              e.currentTarget.nextElementSibling?.classList.remove('hidden')
            }}
          />
        ) : null}
        <div className={mediaUrl ? 'hidden' : 'w-full h-full flex items-center justify-center bg-gray-100 rounded-2xl'}>
          <Smile className="w-16 h-16 text-gray-300" />
        </div>
      </div>
    </div>
  )
}
