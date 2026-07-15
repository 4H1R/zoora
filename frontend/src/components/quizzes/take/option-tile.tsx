import { CheckCircle2Icon } from "lucide-react"

import { OptionImageThumb } from "@/components/admin/questions/OptionImage"
import { cn } from "@/lib/utils"

import { SystemImage } from "./system-image"

interface OptionTileProps {
  index: number
  label: string
  checked: boolean
  onClick: () => void
  imageMediaID?: string
  /** Anti-cheat image of the option value — when set, the text label is withheld. */
  systemImageMediaID?: string
}

export function OptionTile({ index, label, checked, onClick, imageMediaID, systemImageMediaID }: OptionTileProps) {
  const letter = String.fromCharCode(65 + index)
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={checked}
      className={cn(
        "group/option ring-foreground/15 hover:ring-foreground/40 bg-card relative flex items-center gap-4 rounded-xl px-4 py-3 text-start ring-1 transition-all hover:-translate-y-0.5",
        checked && "ring-foreground bg-foreground/[0.04] shadow-sm"
      )}
    >
      <span
        className={cn(
          "ring-foreground/20 flex size-8 shrink-0 items-center justify-center rounded-lg font-mono text-sm font-semibold ring-1 transition-colors",
          checked && "bg-foreground text-background ring-foreground"
        )}
      >
        {letter}
      </span>
      {systemImageMediaID ? (
        <span className="min-w-0 flex-1">
          <SystemImage mediaID={systemImageMediaID} className="max-h-14 w-auto" />
        </span>
      ) : (
        <>
          {imageMediaID && (
            <span className="shrink-0">
              <OptionImageThumb mediaID={imageMediaID} />
            </span>
          )}
          <span className="text-foreground text-base leading-snug">{label}</span>
        </>
      )}
      {checked && <CheckCircle2Icon className="text-foreground ms-auto size-5" />}
    </button>
  )
}
