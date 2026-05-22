import type { ComponentProps } from "react"

import { cn } from "@/lib/utils"

export function Eyebrow({ className, ...props }: ComponentProps<"span">) {
  return (
    <span
      className={cn("text-muted-foreground font-mono text-xs tracking-[0.3em] uppercase", className)}
      {...props}
    />
  )
}
