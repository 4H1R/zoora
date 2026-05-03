import { create } from "zustand"
import { persist } from "zustand/middleware"

type Theme = "light" | "dark"

interface ThemeState {
  theme: Theme
  toggle: () => void
  set: (theme: Theme) => void
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set) => ({
      theme:
        typeof window !== "undefined" && window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light",
      toggle: () =>
        set((s) => {
          const next = s.theme === "light" ? "dark" : "light"
          return { theme: next }
        }),
      set: (theme) => {
        set({ theme })
      },
    }),
    { name: "theme" }
  )
)
