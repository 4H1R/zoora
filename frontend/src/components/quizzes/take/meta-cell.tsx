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
    <div className="bg-card flex flex-col gap-3 px-5 py-6 md:px-6 md:py-7">
      <span className="text-muted-foreground inline-flex items-center gap-2.5 text-xs font-medium">
        <span className="bg-foreground/5 ring-foreground/10 text-foreground/70 flex size-7 items-center justify-center rounded-lg ring-1 [&_svg]:size-3.5">
          {icon}
        </span>
        {label}
      </span>
      <span
        className={cn(
          "text-foreground ps-0.5 text-xl leading-none font-semibold md:text-2xl",
          mono && "font-mono tabular-nums",
        )}
      >
        {value}
      </span>
    </div>
  )
}
