import { create } from "zustand"

interface ChatUiState {
  replyTo: string | null
  editingMessageId: string | null
  scrollToMessageId: string | null
  setReplyTo: (id: string | null) => void
  setEditing: (id: string | null) => void
  requestScrollTo: (id: string | null) => void
}

export const useChatUi = create<ChatUiState>((set) => ({
  replyTo: null,
  editingMessageId: null,
  scrollToMessageId: null,
  setReplyTo: (replyTo) => set({ replyTo }),
  setEditing: (editingMessageId) => set({ editingMessageId }),
  requestScrollTo: (scrollToMessageId) => set({ scrollToMessageId }),
}))
