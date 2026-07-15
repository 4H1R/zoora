import { useId } from "react"
import { useTranslation } from "react-i18next"

import { cn } from "@/lib/utils"

interface LogoMarkProps {
  className?: string
}

/** Zoora brand mark — a green tile with a stylized "Z". Font-independent, safe at any size. */
export function LogoMark({ className }: LogoMarkProps) {
  const gradientId = useId()
  return (
    <svg
      viewBox="0 0 48 48"
      fill="none"
      role="img"
      aria-hidden="true"
      className={cn("h-[1.15em] w-[1.15em] shrink-0", className)}
    >
      <defs>
        <linearGradient id={gradientId} x1="6" y1="4" x2="42" y2="46" gradientUnits="userSpaceOnUse">
          <stop offset="0" stopColor="#2fbd68" />
          <stop offset="0.55" stopColor="#16a34a" />
          <stop offset="1" stopColor="#15803d" />
        </linearGradient>
      </defs>
      <rect width="48" height="48" rx="13" fill={`url(#${gradientId})`} />
      <path d="M14 16 H34 L14 32 H34" stroke="#ffffff" strokeWidth="5.4" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

interface LogoProps {
  className?: string
  /** "full" (mark + wordmark) or "mark" (icon only). Defaults to "full". */
  variant?: "full" | "mark"
}

export function Logo({ className, variant = "full" }: LogoProps) {
  const { t } = useTranslation()
  if (variant === "mark") {
    return <LogoMark className={className} />
  }
  return (
    <span className={cn("inline-flex items-center gap-2 text-lg leading-none font-bold", className)}>
      <LogoMark />
      <span className="tracking-tight">{t("common.brandName")}</span>
    </span>
  )
}
