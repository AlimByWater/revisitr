import { useState, useCallback, useRef, useLayoutEffect, forwardRef, type TextareaHTMLAttributes } from "react";
import { EmojiPicker } from "@/features/emoji-packs";
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from "@dnd-kit/core";
import {
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
  useSortable,
  arrayMove,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import {
  Plus,
  Trash2,
  GripVertical,
  Type,
  ImageIcon,
  Video,
  FileText,
  Smile,
  Music,
  Link,
  ChevronDown,
  Upload,
} from "lucide-react";
import { cn } from "@/lib/utils";
import type {
  MessageContent,
  MessagePart,
  MessagePartType,
  InlineButton,
} from "../types";

interface MessageContentEditorProps {
  value: MessageContent;
  onChange: (content: MessageContent) => void;
  onUpload: (file: File) => Promise<string>;
  maxParts?: number;
}

const PART_TYPE_OPTIONS: {
  type: MessagePartType;
  icon: typeof Type;
  label: string;
}[] = [
  { type: "text", icon: Type, label: "Текст" },
  { type: "photo", icon: ImageIcon, label: "Фото" },
  { type: "video", icon: Video, label: "Видео" },
  { type: "document", icon: FileText, label: "Документ" },
  { type: "sticker", icon: Smile, label: "Стикер" },
  { type: "audio", icon: Music, label: "Аудио" },
];

function createEmptyPart(type: MessagePartType): MessagePart {
  if (type === "text") {
    return { type, text: "", parse_mode: "Markdown" };
  }
  return { type, media_url: "", text: "", parse_mode: "Markdown" };
}

function convertPartType(
  part: MessagePart,
  nextType: MessagePartType,
): MessagePart {
  if (part.type === nextType) return part;

  if (nextType === "text") {
    return {
      type: "text",
      text: part.text || "",
      parse_mode: part.parse_mode || "Markdown",
    };
  }

  return {
    type: nextType,
    text: part.text || "",
    media_url: part.media_url || "",
    media_id: part.media_id || "",
    parse_mode: part.parse_mode || "Markdown",
  };
}

export function MessageContentEditor({
  value,
  onChange,
  onUpload,
  maxParts = 5,
}: MessageContentEditorProps) {
  const [showTypeMenu, setShowTypeMenu] = useState(false);

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    }),
  );

  const updatePart = useCallback(
    (index: number, updates: Partial<MessagePart>) => {
      const parts = [...value.parts];
      parts[index] = { ...parts[index], ...updates };
      onChange({ ...value, parts });
    },
    [value, onChange],
  );

  const removePart = useCallback(
    (index: number) => {
      const parts = value.parts.filter((_, i) => i !== index);
      onChange({ ...value, parts });
    },
    [value, onChange],
  );

  const addPart = useCallback(
    (type: MessagePartType) => {
      if (value.parts.length >= maxParts) return;
      onChange({ ...value, parts: [...value.parts, createEmptyPart(type)] });
      setShowTypeMenu(false);
    },
    [value, onChange, maxParts],
  );

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      const { active, over } = event;
      if (!over || active.id === over.id) return;
      const oldIndex = value.parts.findIndex(
        (_, i) => `part-${i}` === active.id,
      );
      const newIndex = value.parts.findIndex((_, i) => `part-${i}` === over.id);
      if (oldIndex === -1 || newIndex === -1) return;
      onChange({ ...value, parts: arrayMove(value.parts, oldIndex, newIndex) });
    },
    [value, onChange],
  );

  // Buttons management
  const updateButtons = useCallback(
    (buttons: InlineButton[][]) => {
      onChange({ ...value, buttons });
    },
    [value, onChange],
  );

  const addButtonRow = useCallback(() => {
    const buttons = [...(value.buttons || []), [{ text: "", url: "" }]];
    updateButtons(buttons);
  }, [value, updateButtons]);

  const partIds = value.parts.map((_, i) => `part-${i}`);

  return (
    <div className="space-y-3">
      {/* Parts list */}
      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragEnd={handleDragEnd}
      >
        <SortableContext items={partIds} strategy={verticalListSortingStrategy}>
          {value.parts.map((part, i) => (
            <SortablePartEditor
              key={`part-${i}`}
              id={`part-${i}`}
              part={part}
              index={i}
              onUpdate={updatePart}
              onRemove={removePart}
              onUpload={onUpload}
              canRemove={value.parts.length > 1}
            />
          ))}
        </SortableContext>
      </DndContext>

      {/* Add part button */}
      {value.parts.length < maxParts && (
        <div className="relative">
          <button
            type="button"
            onClick={() => setShowTypeMenu(!showTypeMenu)}
            className="flex min-h-11 items-center gap-2 w-full px-3 py-2 border border-dashed border-neutral-300 rounded-lg text-sm text-neutral-500 hover:border-neutral-400 hover:text-neutral-600 transition-colors"
          >
            <Plus className="w-4 h-4" />
            Добавить блок
            <ChevronDown
              className={cn(
                "w-3 h-3 ml-auto transition-transform",
                showTypeMenu && "rotate-180",
              )}
            />
          </button>
          {showTypeMenu && (
            <div className="absolute top-full left-0 right-0 mt-1 bg-white border border-neutral-200 rounded-lg shadow-lg z-10 py-1">
              {PART_TYPE_OPTIONS.map((opt) => (
                <button
                  key={opt.type}
                  type="button"
                  onClick={() => addPart(opt.type)}
                  className="flex items-center gap-2 w-full px-3 py-2 text-sm text-neutral-700 hover:bg-neutral-50"
                >
                  <opt.icon className="w-4 h-4" />
                  {opt.label}
                </button>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Inline buttons editor */}
      <div className="border-t pt-3">
        <div className="flex items-center justify-between mb-2">
          <span className="text-sm font-medium text-neutral-700">Кнопки</span>
          <button
            type="button"
            onClick={addButtonRow}
            className="inline-flex min-h-11 items-center gap-1 px-2 text-sm text-accent hover:text-accent-hover"
          >
            <Plus className="w-3 h-3" />
            Добавить ряд
          </button>
        </div>
        {(value.buttons || []).map((row, ri) => (
          <ButtonRowEditor
            key={ri}
            row={row}
            rowIndex={ri}
            onChange={(newRow) => {
              const buttons = [...(value.buttons || [])];
              buttons[ri] = newRow;
              updateButtons(buttons);
            }}
            onRemove={() => {
              const buttons = (value.buttons || []).filter((_, i) => i !== ri);
              updateButtons(buttons);
            }}
          />
        ))}
      </div>
    </div>
  );
}

// ── Sortable Part Editor ──────────────────────────────────────────────────────

function SortablePartEditor({
  id,
  part,
  index,
  onUpdate,
  onRemove,
  onUpload,
  canRemove,
}: {
  id: string;
  part: MessagePart;
  index: number;
  onUpdate: (i: number, updates: Partial<MessagePart>) => void;
  onRemove: (i: number) => void;
  onUpload: (file: File) => Promise<string>;
  canRemove: boolean;
}) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id });
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      className="bg-white border border-neutral-200 rounded-lg p-3"
    >
      <div className="flex items-center gap-2 mb-2">
        <button
          type="button"
          className="inline-flex min-h-11 min-w-11 items-center justify-center rounded-lg cursor-grab text-neutral-400 hover:text-neutral-600 hover:bg-neutral-50"
          {...attributes}
          {...listeners}
          aria-label={`Перетащить блок ${index + 1}`}
        >
          <GripVertical className="w-4 h-4" />
        </button>
        <select
          value={part.type}
          onChange={(e) =>
            onUpdate(
              index,
              convertPartType(part, e.target.value as MessagePartType),
            )
          }
          className="text-xs font-medium uppercase text-neutral-500 bg-transparent border border-neutral-200 rounded px-2 py-1 focus:outline-none focus:ring-1 focus:ring-accent/30"
          aria-label={`Тип блока ${index + 1}`}
        >
          {PART_TYPE_OPTIONS.map((option) => (
            <option key={option.type} value={option.type}>
              {option.label}
            </option>
          ))}
        </select>
        {canRemove && (
          <button
            type="button"
            onClick={() => onRemove(index)}
            className="ml-auto inline-flex min-h-11 min-w-11 items-center justify-center rounded-lg text-neutral-400 hover:text-red-500 hover:bg-red-50"
            aria-label={`Удалить блок ${index + 1}`}
          >
            <Trash2 className="w-3.5 h-3.5" />
          </button>
        )}
      </div>

      {part.type === "text" ? (
        <TextPartEditor part={part} onChange={(u) => onUpdate(index, u)} />
      ) : part.type === "sticker" ? (
        <MediaUploadField
          value={part.media_url || ""}
          onChange={(url) => onUpdate(index, { media_url: url })}
          onUpload={onUpload}
          accept=".webp,.png,.tgs"
          label="Стикер (.webp)"
        />
      ) : (
        <MediaPartEditor
          part={part}
          onChange={(u) => onUpdate(index, u)}
          onUpload={onUpload}
        />
      )}
    </div>
  );
}

// ── Text Part Editor ──────────────────────────────────────────────────────────

function TextPartEditor({
  part,
  onChange,
}: {
  part: MessagePart;
  onChange: (u: Partial<MessagePart>) => void;
}) {
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const handleEmojiInsert = (item: { name: string; image_url: string }) => {
    const textarea = textareaRef.current;
    if (!textarea) {
      onChange({ text: (part.text || "") + item.name });
      return;
    }
    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const text = part.text || "";
    const newText = text.slice(0, start) + item.name + text.slice(end);
    onChange({ text: newText });
    requestAnimationFrame(() => {
      const pos = start + item.name.length;
      textarea.setSelectionRange(pos, pos);
      textarea.focus();
    });
  };

  return (
    <div className="space-y-2">
      <div className="relative">
        <AutoResizeTextarea
          ref={textareaRef}
          value={part.text || ""}
          onChange={(e) => onChange({ text: e.target.value })}
          placeholder="Текст сообщения..."
          rows={3}
          maxLength={4096}
          className="w-full overflow-hidden px-3 py-2 pr-10 border border-neutral-200 rounded-lg text-sm resize-none focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent"
        />
        <div className="absolute top-2 right-2">
          <EmojiPicker onSelect={handleEmojiInsert} triggerClassName="w-6 h-6 border-0 bg-transparent hover:bg-neutral-100" />
        </div>
      </div>
      <div className="flex items-center justify-between text-xs text-neutral-400">
        <span>*жирный*, _курсив_, `код`</span>
        <span>{(part.text || "").length}/4096</span>
      </div>
    </div>
  );
}

// ── Media Part Editor ─────────────────────────────────────────────────────────

function MediaPartEditor({
  part,
  onChange,
  onUpload,
}: {
  part: MessagePart;
  onChange: (u: Partial<MessagePart>) => void;
  onUpload: (file: File) => Promise<string>;
}) {
  const accept =
    part.type === "photo"
      ? "image/*"
      : part.type === "video" || part.type === "animation"
        ? "video/*,image/gif"
        : part.type === "audio"
          ? "audio/*"
          : "*/*";

  return (
    <div className="space-y-2">
      <MediaUploadField
        value={part.media_url || ""}
        onChange={(url) => onChange({ media_url: url })}
        onUpload={onUpload}
        accept={accept}
        label="Загрузить файл"
      />
      {part.type !== "sticker" && (
        <AutoResizeTextarea
          value={part.text || ""}
          onChange={(e) => onChange({ text: e.target.value })}
          placeholder="Подпись (опционально)..."
          rows={2}
          maxLength={1024}
          className="w-full overflow-hidden px-3 py-2 border border-neutral-200 rounded-lg text-sm resize-none focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent"
        />
      )}
    </div>
  );
}

const AutoResizeTextarea = forwardRef<HTMLTextAreaElement, TextareaHTMLAttributes<HTMLTextAreaElement>>(
  function AutoResizeTextarea(props, ref) {
    const internalRef = useRef<HTMLTextAreaElement>(null);
    const resolvedRef = (ref as React.RefObject<HTMLTextAreaElement>) || internalRef;

    useLayoutEffect(() => {
      const textarea = resolvedRef.current;
      if (!textarea) return;

      textarea.style.height = "0px";
      textarea.style.height = `${textarea.scrollHeight}px`;
    }, [props.value]);

    return <textarea ref={resolvedRef} {...props} />;
  },
);

// ── Media Upload Field ────────────────────────────────────────────────────────

function MediaUploadField({
  value,
  onChange,
  onUpload,
  accept,
  label,
}: {
  value: string;
  onChange: (url: string) => void;
  onUpload: (file: File) => Promise<string>;
  accept: string;
  label: string;
}) {
  const [uploading, setUploading] = useState(false);
  const fileRef = useRef<HTMLInputElement>(null);

  const handleUpload = useCallback(
    async (file: File) => {
      setUploading(true);
      try {
        const url = await onUpload(file);
        onChange(url);
      } catch {
        // Upload failed silently
      } finally {
        setUploading(false);
      }
    },
    [onChange, onUpload],
  );

  return (
    <div>
      <input
        ref={fileRef}
        type="file"
        accept={accept}
        className="hidden"
        onChange={(e) => {
          const file = e.target.files?.[0];
          if (file) handleUpload(file);
        }}
      />
      {value ? (
        <div className="flex items-center gap-2">
          {value.match(/\.(jpg|jpeg|png|gif|webp)$/i) && (
            <img
              src={value}
              alt=""
              loading="lazy"
              decoding="async"
              className="w-12 h-12 rounded object-cover"
            />
          )}
          <span className="text-sm text-neutral-600 truncate flex-1">
            {value.split("/").pop()}
          </span>
          <button
            type="button"
            onClick={() => fileRef.current?.click()}
            className="text-xs text-accent hover:text-accent-hover"
          >
            Заменить
          </button>
          <button
            type="button"
            onClick={() => onChange("")}
            className="text-xs text-red-500 hover:text-red-600"
          >
            Удалить
          </button>
        </div>
      ) : (
        <button
          type="button"
          onClick={() => fileRef.current?.click()}
          disabled={uploading}
          className="flex items-center gap-2 w-full px-3 py-3 border border-dashed border-neutral-300 rounded-lg text-sm text-neutral-500 hover:border-accent hover:text-accent transition-colors disabled:opacity-50"
        >
          <Upload className="w-4 h-4" />
          {uploading ? "Загрузка..." : label}
        </button>
      )}
    </div>
  );
}

// ── Button Row Editor ─────────────────────────────────────────────────────────

function ButtonRowEditor({
  row,
  rowIndex,
  onChange,
  onRemove,
}: {
  row: InlineButton[];
  rowIndex: number;
  onChange: (row: InlineButton[]) => void;
  onRemove: () => void;
}) {
  return (
    <div className="flex items-start gap-2 mb-2">
      <div className="flex-1 space-y-1">
        {row.map((btn, bi) => (
          <div key={bi} className="flex items-center gap-1">
            <input
              type="text"
              value={btn.text}
              onChange={(e) => {
                const newRow = [...row];
                newRow[bi] = { ...btn, text: e.target.value };
                onChange(newRow);
              }}
              placeholder="Текст кнопки"
              className="flex-1 px-2 py-1.5 border border-neutral-200 rounded text-sm focus:outline-none focus:ring-1 focus:ring-accent/30"
            />
            <Link className="w-3.5 h-3.5 text-neutral-400 flex-shrink-0" />
            <input
              type="text"
              value={btn.url || ""}
              onChange={(e) => {
                const newRow = [...row];
                newRow[bi] = { ...btn, url: e.target.value };
                onChange(newRow);
              }}
              placeholder="URL"
              className="flex-1 px-2 py-1.5 border border-neutral-200 rounded text-sm focus:outline-none focus:ring-1 focus:ring-accent/30"
            />
            {row.length > 1 && (
              <button
                type="button"
                onClick={() => onChange(row.filter((_, i) => i !== bi))}
                className="text-neutral-400 hover:text-red-500"
              >
                <Trash2 className="w-3 h-3" />
              </button>
            )}
          </div>
        ))}
        {row.length < 8 && (
          <button
            type="button"
            onClick={() => onChange([...row, { text: "", url: "" }])}
            className="text-xs text-accent hover:text-accent-hover"
          >
            + кнопка в ряд
          </button>
        )}
      </div>
      <button
        type="button"
        onClick={onRemove}
        className="text-neutral-400 hover:text-red-500 mt-1.5"
      >
        <Trash2 className="w-3.5 h-3.5" />
      </button>
    </div>
  );
}
