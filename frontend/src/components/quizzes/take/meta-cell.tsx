import type { ReactNode } from "react"

import { cn } from "@/lib/utils"

interface MetaCellProps {
  icon: ReactNode
  label: string
  value: string
  mono?: boolean
}

export function MetaCell({ icon, label, value, mono = false }: MetaCellProps) {
  return (
    <div className="flex flex-col gap-2 border-b border-dashed py-5 pe-4 ps-4 md:border-b-0 md:border-s md:py-0 md:first:border-s-0 md:first:ps-0">
      <span className="text-muted-foreground inline-flex items-center gap-2 font-mono text-xs tracking-[0.25em] uppercase">
        {icon}
        {label}
      </span>
      <span
        className={cn(
          "text-foreground text-base leading-tight font-medium md:text-lg",
          mono && "font-mono tabular-nums",
        )}
      >
        {value}
      </span>
    </div>
  )
}
