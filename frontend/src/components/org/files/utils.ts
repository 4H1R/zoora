import type { LucideIcon } from "lucide-react"

import {
  FileQuestionIcon,
  FolderIcon,
  FolderHeartIcon,
  NotebookPenIcon,
  FileCheck2Icon,
  RadioIcon,
  FolderClockIcon,
  VideoIcon,
} from "lucide-react"

export const SHARED_FOLDER = "organization"

// Icon + tint per known folder (media model_type). Unknown types fall back to
// a plain folder so new backend model types never break the page.
export const FOLDER_STYLES: Record<string, { icon: LucideIcon; tint: string }> = {
  [SHARED_FOLDER]: { icon: FolderHeartIcon, tint: "bg-primary/10 text-primary" },
  live_room: { icon: VideoIcon, tint: "bg-sky-500/10 text-sky-600 dark:text-sky-400" },
  live_session: { icon: RadioIcon, tint: "bg-rose-500/10 text-rose-600 dark:text-rose-400" },
  offline_room: { icon: FolderClockIcon, tint: "bg-amber-500/10 text-amber-600 dark:text-amber-400" },
  practice: { icon: NotebookPenIcon, tint: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400" },
  practice_submission: { icon: FileCheck2Icon, tint: "bg-violet-500/10 text-violet-600 dark:text-violet-400" },
  question: { icon: FileQuestionIcon, tint: "bg-stone-500/10 text-stone-600 dark:text-stone-400" },
}

export const FALLBACK_FOLDER_STYLE = { icon: FolderIcon, tint: "bg-muted text-muted-foreground" }

export function folderStyle(modelType: string) {
  return FOLDER_STYLES[modelType] ?? FALLBACK_FOLDER_STYLE
}

export function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`
}
