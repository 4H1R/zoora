import { Moon, Sun } from "lucide-react"
import { useTranslation } from "react-i18next"

import { cn } from "@/lib/utils"
import { useThemeStore } from "@/stores/theme"

export function ThemeToggle({ className }: { className?: string }) {
  const { t } = useTranslation()
  const { theme, toggle } = useThemeStore()
  const isDark = theme === "dark"

  return (
    <button
      type="button"
      role="switch"
      aria-checked={isDark}
      aria-label={isDark ? t("theme.light") : t("theme.dark")}
      onClick={toggle}
      className={cn(
        "border-border/60 relative inline-flex h-7 w-[52px] shrink-0 cursor-pointer items-center rounded-full border transition-colors duration-[var(--dur-slow)] ease-[var(--ease-out)]",
        isDark
          ? "border-[var(--neutral-700)] bg-[var(--neutral-800)]"
          : "border-[var(--neutral-200)] bg-[var(--neutral-100)]",
        className
      )}
    >
      <span
        className={cn(
          "pointer-events-none absolute flex size-5 items-center justify-center rounded-full shadow-sm transition-all duration-[var(--dur-slow)] ease-[var(--ease-out)]",
          isDark ? "start-[26px] bg-[var(--neutral-700)]" : "start-[3px] bg-white"
        )}
      >
        {isDark ? (
          <Moon className="size-3 text-[var(--neutral-300)]" />
        ) : (
          <Sun className="size-3 text-[var(--neutral-500)]" />
        )}
      </span>
    </button>
  )
}
