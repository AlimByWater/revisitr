import { useState } from "react";
import { Smile } from "lucide-react";

interface StickerMessageProps {
  mediaUrl?: string;
}

export function StickerMessage({ mediaUrl }: StickerMessageProps) {
  const [hasError, setHasError] = useState(false);

  return (
    <div className="flex justify-start">
      <div className="h-[150px] w-[150px]">
        {mediaUrl && !hasError ? (
          <img
            src={mediaUrl}
            alt="sticker"
            loading="lazy"
            decoding="async"
            className="h-full w-full object-contain"
            onError={() => setHasError(true)}
          />
        ) : (
          <div className="flex h-full w-full items-center justify-center rounded-2xl bg-gray-100">
            <Smile className="h-16 w-16 text-gray-300" />
          </div>
        )}
      </div>
    </div>
  );
}
