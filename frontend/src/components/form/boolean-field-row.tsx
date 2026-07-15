import type { ReactNode } from "react"

import { Checkbox } from "@/components/ui/checkbox"
import { cn } from "@/lib/utils"

interface BooleanFieldRowProps {
  label: ReactNode
  hint?: ReactNode
  /** Optional leading icon, shown in a badge that tints when checked. */
  icon?: ReactNode
  checked: boolean
  onCheckedChange: (checked: boolean) => void
  disabled?: boolean
  className?: string
}

export function BooleanFieldRow({
  label,
  hint,
  icon,
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
      {Boolean(icon) && (
        <span
          className={cn(
            "bg-muted text-muted-foreground mt-px flex size-8 shrink-0 items-center justify-center rounded-md transition-colors",
            "[&_svg]:size-4",
            "group-data-[checked=true]/bool-row:bg-primary/10 group-data-[checked=true]/bool-row:text-primary"
          )}
        >
          {icon}
        </span>
      )}
      <div className="flex min-w-0 flex-1 flex-col gap-0.5">
        <span className="text-sm leading-snug font-medium">{label}</span>
        {Boolean(hint) && <span className="text-muted-foreground text-xs leading-relaxed">{hint}</span>}
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
