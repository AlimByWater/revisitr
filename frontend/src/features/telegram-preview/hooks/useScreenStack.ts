import { useState, useCallback, useMemo } from 'react'
import type { MessageContent, PreviewScreen } from '../types'

export function useScreenStack(initialContent: MessageContent) {
  const [stack, setStack] = useState<PreviewScreen[]>([{ content: initialContent }])

  const push = useCallback(
    (screen: PreviewScreen) => setStack((prev) => [...prev, screen]),
    [],
  )

  const replace = useCallback(
    (screen: PreviewScreen) =>
      setStack((prev) => (prev.length > 0 ? [...prev.slice(0, -1), screen] : [screen])),
    [],
  )

  const reset = useCallback(
    () => setStack([{ content: initialContent }]),
    [initialContent],
  )

  const currentScreen = useMemo(() => stack[stack.length - 1], [stack])
  const canReset = stack.length > 1

  return { stack, currentScreen, push, replace, reset, canReset }
}
