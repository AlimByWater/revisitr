import { Node, mergeAttributes } from '@tiptap/core'

export const CustomEmoji = Node.create({
  name: 'customEmoji',
  group: 'inline',
  inline: true,
  atom: true,

  addAttributes() {
    return {
      src: { default: null },
    }
  },

  parseHTML() {
    return [{ tag: 'img[data-emoji-src]' }]
  },

  renderHTML({ HTMLAttributes }) {
    return [
      'img',
      mergeAttributes(HTMLAttributes, {
        'data-emoji-src': HTMLAttributes.src,
        src: HTMLAttributes.src,
        class:
          'inline-block w-5 h-5 align-text-bottom rounded-sm object-cover cursor-default',
        draggable: 'false',
      }),
    ]
  },
})
