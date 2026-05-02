interface ChatHeaderProps {
  botName: string
  botAvatar?: string
}

export function ChatHeader({ botName, botAvatar }: ChatHeaderProps) {
  return (
    <div className="relative z-10 flex shrink-0 items-center justify-between gap-[6px] px-[12px] pb-[4px] pt-[1px]">
      {/* Back button */}
      <GlassPill className="flex h-[33px] w-[33px] shrink-0 items-center justify-center">
        <svg width="8" height="13" viewBox="0 0 10 18" fill="none">
          <path
            d="M9 1L1.5 9 9 17"
            stroke="#1a1a1a"
            strokeWidth="2.2"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      </GlassPill>

      {/* Title */}
      <GlassPill className="flex h-[33px] min-w-0 flex-1 items-center justify-center px-[12px]">
        <div className="flex flex-col items-center">
          <span className="max-w-[140px] truncate text-[11px] font-semibold leading-[13px] tracking-[-0.17px] text-[#333]">
            {botName}
          </span>
          <span className="text-[9px] font-medium leading-[10px] text-[#999]">
            бот
          </span>
        </div>
      </GlassPill>

      {/* Avatar */}
      <GlassPill className="flex h-[33px] w-[33px] shrink-0 items-center justify-center overflow-hidden p-0">
        {botAvatar ? (
          <img
            src={botAvatar}
            alt={botName}
            className="h-full w-full rounded-full object-cover"
          />
        ) : (
          <div className="flex h-full w-full items-center justify-center rounded-full bg-gradient-to-br from-[#6cb6ff] via-[#4c8fe8] to-[#2e72db] text-[12px] font-semibold text-white">
            {botName.charAt(0).toUpperCase()}
          </div>
        )}
      </GlassPill>
    </div>
  )
}

function GlassPill({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div
      className={`relative overflow-hidden rounded-[296px] ${className ?? ''}`}
      style={{
        background: 'rgba(255, 255, 255, 0.92)',
        boxShadow: '0 1px 8px rgba(0, 0, 0, 0.1)',
      }}
    >
      {children}
    </div>
  )
}
