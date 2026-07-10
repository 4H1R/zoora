import { create } from "zustand"

/** What identifies the person whose card to show. */
export interface ProfileCardTarget {
  /** Known user id (avatar / member / search click). Absent for mention clicks. */
  userId?: string
  /** Username to resolve when `userId` is absent, and to display. */
  username?: string
  /** Prefetched display name, when the trigger already has it. */
  name?: string
}

interface ProfileCardState {
  target: ProfileCardTarget | null
  open: (target: ProfileCardTarget) => void
  close: () => void
}

/** Single global entry point for opening the profile card from anywhere. */
export const useProfileCard = create<ProfileCardState>((set) => ({
  target: null,
  open: (target) => set({ target }),
  close: () => set({ target: null }),
}))
