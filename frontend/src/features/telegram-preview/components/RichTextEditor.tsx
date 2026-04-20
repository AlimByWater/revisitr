import { useEditor, EditorContent } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import Placeholder from '@tiptap/extension-placeholder'
import CharacterCount from '@tiptap/extension-character-count'
import { useEffect, useCallback, useImperativeHandle, forwardRef } from 'react'
import { cn } from '@/lib/utils'
import { CustomEmoji } from './EmojiNode'

const EMOJI_RE = /\{\{emoji:([^}]+)\}\}/g

/** Convert plain text with {{emoji:URL}} markers to Tiptap-compatible HTML. */
function deserialize(text: string): string {
  if (!text) return '<p></p>'
  const html = text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(EMOJI_RE, (_match, src) => `<img data-emoji-src="${src}" src="${src}">`)
    .replace(/\n/g, '<br>')
  return `<p>${html}</p>`
}

/** Convert Tiptap editor JSON back to plain text with {{emoji:URL}} markers. */
function serialize(editor: ReturnType<typeof useEditor>): string {
  if (!editor) return ''
  const doc = editor.getJSON()
  return serializeNode(doc).trim()
}

function serializeNode(node: Record<string, unknown>): string {
  if (node.type === 'customEmoji') {
    const attrs = node.attrs as { src?: string } | undefined
    return attrs?.src ? `{{emoji:${attrs.src}}}` : ''
  }
  if (node.type === 'text') {
    return (node.text as string) || ''
  }
  if (node.type === 'hardBreak') {
    return '\n'
  }

  const children = (node.content as Record<string, unknown>[] | undefined) || []
  const inner = children.map(serializeNode).join('')

  if (node.type === 'paragraph') {
    return inner + '\n'
  }
  return inner
}

export interface RichTextEditorHandle {
  insertEmoji: (src: string) => void
}

interface RichTextEditorProps {
  value: string
  onChange: (text: string) => void
  placeholder?: string
  maxLength?: number
  className?: string
}

export const RichTextEditor = forwardRef<RichTextEditorHandle, RichTextEditorProps>(
  function RichTextEditor({ value, onChange, placeholder, maxLength = 4096, className }, ref) {
    const editor = useEditor({
      extensions: [
        StarterKit.configure({
          // Disable formatting — plain text only with inline emoji
          bold: false,
          italic: false,
          strike: false,
          code: false,
          codeBlock: false,
          blockquote: false,
          heading: false,
          bulletList: false,
          orderedList: false,
          listItem: false,
          horizontalRule: false,
        }),
        CustomEmoji,
        Placeholder.configure({ placeholder: placeholder || 'Текст сообщения...' }),
        CharacterCount.configure({ limit: maxLength }),
      ],
      content: deserialize(value),
      onUpdate({ editor: ed }) {
        onChange(serialize(ed))
      },
      editorProps: {
        attributes: {
          class: 'outline-none min-h-[60px] whitespace-pre-wrap',
        },
      },
    })

    // Sync external value changes (e.g., on load)
    useEffect(() => {
      if (!editor || editor.isDestroyed) return
      const current = serialize(editor)
      if (current.trim() !== value.trim()) {
        editor.commands.setContent(deserialize(value), { emitUpdate: false })
      }
    }, [value]) // eslint-disable-line react-hooks/exhaustive-deps

    const insertEmoji = useCallback(
      (src: string) => {
        if (!editor) return
        editor.chain().focus().insertContent({ type: 'customEmoji', attrs: { src } }).run()
      },
      [editor],
    )

    useImperativeHandle(ref, () => ({ insertEmoji }), [insertEmoji])

    const charCount = editor?.storage.characterCount?.characters() ?? 0

    return (
      <div className={cn('space-y-2', className)}>
        <div
          className={cn(
            'w-full px-3 py-2 border border-neutral-200 rounded-lg text-sm',
            'focus-within:ring-2 focus-within:ring-accent/20 focus-within:border-accent',
            'transition-colors',
          )}
        >
          <EditorContent editor={editor} />
        </div>
        <div className="flex items-center justify-between text-xs text-neutral-400">
          <span>*жирный*, _курсив_, `код`</span>
          <span>{charCount}/{maxLength}</span>
        </div>
      </div>
    )
  },
)
