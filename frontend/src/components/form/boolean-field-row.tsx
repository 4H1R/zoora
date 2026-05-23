import type { ReactNode } from "react"

import { Checkbox } from "@/components/ui/checkbox"
import { cn } from "@/lib/utils"

interface BooleanFieldRowProps {
  label: ReactNode
  hint?: ReactNode
  checked: boolean
  onCheckedChange: (checked: boolean) => void
  disabled?: boolean
  className?: string
}

export function BooleanFieldRow({
  label,
  hint,
  checked,
  onCheckedChange,
  disabled,
  className,
}: BooleanFieldRowProps) {
  return (
    <label
      className={cn(
        "border-foreground/10 hover:bg-accent/40 group/bool-row flex cursor-pointer items-start gap-3 rounded-md border border-dashed p-3 transition-colors",
        "data-[checked=true]:border-primary/40 data-[checked=true]:bg-primary/5",
        disabled && "cursor-not-allowed opacity-60",
        className
      )}
      data-checked={checked}
    >
      <Checkbox
        checked={checked}
        onCheckedChange={(c) => onCheckedChange(!!c)}
        disabled={disabled}
        className="mt-0.5 shrink-0"
      />
      <div className="flex min-w-0 flex-1 flex-col gap-0.5">
        <span className="text-sm leading-snug font-medium">{label}</span>
        {hint ? (
          <span className="text-muted-foreground text-xs leading-relaxed">{hint}</span>
        ) : null}
      </div>
    </label>
  )
}

interface BooleanFieldGroupProps {
  children: ReactNode
  className?: string
}

export function BooleanFieldGroup({ children, className }: BooleanFieldGroupProps) {
  return <div className={cn("flex flex-col gap-2", className)}>{children}</div>
}
