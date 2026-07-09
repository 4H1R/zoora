import { createContext, useContext } from "react"

/**
 * Jump-to-message channel. Deeply-nested UI — reply previews (Phase 6), pinned
 * jumps, mention links — call this to scroll the thread to a message without
 * threading a prop through every layer. `<ChatThread>` supplies the real
 * implementation; the default no-op keeps consumers safe outside a thread.
 */
const JumpToMessageContext = createContext<(id: string) => void>(() => {})

export const JumpToMessageProvider = JumpToMessageContext.Provider

/** Access the thread's jump handler. Returns a no-op outside `<ChatThread>`. */
export function useJumpToMessage(): (id: string) => void {
  return useContext(JumpToMessageContext)
}
