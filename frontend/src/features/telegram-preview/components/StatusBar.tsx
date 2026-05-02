import { PREVIEW_STATUS_TIME } from './previewConstants'

export function StatusBar() {
  return (
    <div className="relative z-10 flex shrink-0 items-center justify-between px-[14px] pb-[1px] pt-[24px]">
      <span className="text-[13px] font-semibold leading-[16px] tracking-[-0.02em] text-black">
        {PREVIEW_STATUS_TIME}
      </span>
      <div className="flex items-center gap-[4px] text-black">
        {/* Cellular */}
        <svg width="14" height="10" viewBox="0 0 17 12" fill="none">
          <rect x="0" y="9" width="3" height="3" rx="0.5" fill="currentColor" />
          <rect x="4.5" y="6" width="3" height="6" rx="0.5" fill="currentColor" />
          <rect x="9" y="3" width="3" height="9" rx="0.5" fill="currentColor" />
          <rect x="13.5" y="0" width="3" height="12" rx="0.5" fill="currentColor" />
        </svg>
        {/* WiFi */}
        <svg width="13" height="10" viewBox="0 0 16 12" fill="none">
          <path d="M8 10.5a1.5 1.5 0 1 0 0 3 1.5 1.5 0 0 0 0-3Z" fill="currentColor" transform="translate(0,-1.5)" />
          <path d="M4.93 9.07a4.36 4.36 0 0 1 6.14 0" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
          <path d="M2.1 6.24a8.05 8.05 0 0 1 11.8 0" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
          <path d="M.34 3.34A11 11 0 0 1 15.66 3.34" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
        </svg>
        {/* Battery */}
        <svg width="20" height="10" viewBox="0 0 25 12" fill="none">
          <rect x="0.5" y="0.5" width="21" height="11" rx="2.5" stroke="currentColor" strokeOpacity="0.35" />
          <rect x="2" y="2" width="16" height="7" rx="1.5" fill="currentColor" />
          <path d="M23 4v4a2 2 0 0 0 0-4Z" fill="currentColor" fillOpacity="0.4" />
        </svg>
      </div>
    </div>
  )
}
