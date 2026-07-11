import { create } from "zustand"

interface ImageLightboxState {
  /** Source of the image to show full-size, or null when closed. */
  src: string | null
  /** Alt / accessible label for the shown image. */
  alt: string
  open: (src: string, alt?: string) => void
  close: () => void
}

/**
 * Single global image lightbox. Rendered once at the conversations route root
 * (mirroring the profile card) so the full-size Dialog never lives inside a
 * message's context-menu subtree — a nested base-ui floating node there
 * entangles with the closed menu and reopens it on dismiss.
 */
export const useImageLightbox = create<ImageLightboxState>((set) => ({
  src: null,
  alt: "",
  open: (src, alt = "") => set({ src, alt }),
  close: () => set({ src: null, alt: "" }),
}))
