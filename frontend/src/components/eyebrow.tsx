import type { ComponentProps } from "react"

import { cn } from "@/lib/utils"

export function Eyebrow({ className, ...props }: ComponentProps<"span">) {
  return (
    <span
      className={cn(
        "text-muted-foreground font-mono text-xs tracking-[0.3em] uppercase",
        // Persian has no caps and breaks under tracking — signal "label" with
        // weight + a hair more size on Vazirmatn instead.
        "rtl:font-sans rtl:text-[0.8rem] rtl:font-semibold rtl:normal-case",
        className
      )}
      {...props}
    />
  )
}
