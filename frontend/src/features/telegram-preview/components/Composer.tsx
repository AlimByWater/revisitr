const pillStyle: React.CSSProperties = {
  background: 'rgba(255, 255, 255, 0.88)',
  boxShadow: '0 1px 6px rgba(0, 0, 0, 0.08)',
}

export function Composer() {
  return (
    <div className="flex shrink-0 items-end gap-[4px] px-[14px] pb-[20px] pt-[4px]">
      {/* Attach */}
      <div className="flex h-[31px] w-[31px] shrink-0 items-center justify-center rounded-full" style={pillStyle}>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
          <path d="M8.5 12.5 14 7a3.5 3.5 0 1 1 5 5l-8.5 8.5a5 5 0 1 1-7-7L13 4" stroke="#555" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </div>

      {/* Field */}
      <div className="flex min-h-[31px] flex-1 items-center rounded-[16px]" style={pillStyle}>
        <div className="flex flex-1 items-center justify-between px-[8px] py-[5px]">
          <span className="text-[11px] font-normal text-[#999]">Сообщение</span>
          <div className="ml-auto flex items-center gap-[4px]">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" className="text-[#999]">
              <circle cx="12" cy="12" r="9" stroke="currentColor" strokeWidth="1.5" />
              <path d="M8 14c.8 1 1.8 1.5 4 1.5s3.2-.5 4-1.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
              <circle cx="9" cy="10" r="1" fill="currentColor" />
              <circle cx="15" cy="10" r="1" fill="currentColor" />
            </svg>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" className="text-[#999]">
              <rect x="3" y="5" width="18" height="14" rx="3" stroke="currentColor" strokeWidth="1.5" />
              <circle cx="8" cy="10" r="0.8" fill="currentColor" />
              <circle cx="12" cy="10" r="0.8" fill="currentColor" />
              <circle cx="16" cy="10" r="0.8" fill="currentColor" />
              <circle cx="8" cy="14" r="0.8" fill="currentColor" />
              <circle cx="12" cy="14" r="0.8" fill="currentColor" />
              <circle cx="16" cy="14" r="0.8" fill="currentColor" />
            </svg>
          </div>
        </div>
      </div>

      {/* Mic */}
      <div className="flex h-[31px] w-[31px] shrink-0 items-center justify-center rounded-full" style={pillStyle}>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
          <path d="M12 4a2.5 2.5 0 0 0-2.5 2.5v5a2.5 2.5 0 0 0 5 0v-5A2.5 2.5 0 0 0 12 4Z" stroke="#555" strokeWidth="1.5" />
          <path d="M7.5 10.5v1a4.5 4.5 0 0 0 9 0v-1M12 16v3" stroke="#555" strokeWidth="1.5" strokeLinecap="round" />
        </svg>
      </div>
    </div>
  )
}
