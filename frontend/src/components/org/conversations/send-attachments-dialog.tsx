import type {
  EmojiPickerListCategoryHeaderProps,
  EmojiPickerListEmojiProps,
  EmojiPickerListRowProps,
} from "frimousse"

import { EmojiPicker } from "frimousse"
import { FileIcon, SmileIcon, XIcon } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import { formatBytes } from "@/components/org/files/utils"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { cn } from "@/lib/utils"

import { insertAtCaret } from "./lib/mentions"
import { isImage } from "./upload/compress"

interface SendAttachmentsDialogProps {
  files: File[]
  /** Send with a caption; `asDocument` forces file-chip rendering (no image preview). */
  onSend: (caption: string, asDocument: boolean) => void
  /** Drop a single staged file before sending. Closing the last one cancels. */
  onRemove: (index: number) => void
  /** Open the file picker to stage more files into the same batch. */
  onAddMore: () => void
  /** Discard the whole batch (Cancel / close / Esc). */
  onCancel: () => void
}

/**
 * Telegram-style "Send an image" modal. Opens whenever files are staged in the
 * composer: shows the previews, a "Send as a document" toggle, and a caption
 * field with its own emoji picker. The composer no longer renders an inline tray
 * — this dialog owns the whole pre-send flow. Enter (outside the emoji popover)
 * sends; the send caption + document flag are handed back to the composer.
 */
export function SendAttachmentsDialog({ files, onSend, onRemove, onAddMore, onCancel }: SendAttachmentsDialogProps) {
  const { t } = useTranslation()
  const [caption, setCaption] = useState("")
  const [asDocument, setAsDocument] = useState(false)
  const [emojiOpen, setEmojiOpen] = useState(false)

  const inputRef = useRef<HTMLInputElement>(null)
  const caretPosRef = useRef(0)

  const open = files.length > 0
  const allImages = files.every((f) => isImage(f))
  const title =
    allImages && !asDocument
      ? t("conversations.attachments.dialogImage", { count: files.length })
      : t("conversations.attachments.dialogFile", { count: files.length })

  function handleSend() {
    onSend(caption.trim(), asDocument)
  }

  function insertEmoji(emoji: string) {
    const { value, caret } = insertAtCaret(caption, caretPosRef.current, emoji)
    setCaption(value)
    caretPosRef.current = caret
    setEmojiOpen(false)
    requestAnimationFrame(() => {
      const el = inputRef.current
      if (!el) return
      el.focus()
      el.setSelectionRange(caret, caret)
    })
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (!next) onCancel()
      }}
    >
      <DialogContent className="sm:max-w-md">
        <DialogTitle>{title}</DialogTitle>

        {/* Preview grid — a lone item goes full-width, the rest tile. */}
        <div
          className={cn("grid gap-2", files.length <= 1 ? "grid-cols-1" : "max-h-72 grid-cols-3 overflow-y-auto p-2")}
          aria-label={t("conversations.attachments.tray")}
        >
          {files.map((file, index) => (
            <PreviewItem
              key={`${file.name}-${file.size}-${index}`}
              file={file}
              single={files.length <= 1}
              asDocument={asDocument}
              onRemove={() => onRemove(index)}
            />
          ))}
        </div>

        {/* Send-as-document toggle. */}
        <label className="flex cursor-pointer items-center gap-2 text-sm">
          <Checkbox
            checked={asDocument}
            onCheckedChange={(v) => setAsDocument(v === true)}
            aria-label={t("conversations.attachments.asDocument")}
          />
          <span>{t("conversations.attachments.asDocument")}</span>
        </label>

        {/* Caption row: text field + emoji picker. */}
        <div className="focus-within:ring-ring/40 flex items-end gap-1 rounded-lg border px-1 transition focus-within:ring-2">
          <input
            ref={inputRef}
            value={caption}
            onChange={(e) => {
              setCaption(e.target.value)
              caretPosRef.current = e.target.selectionStart ?? e.target.value.length
            }}
            onKeyUp={(e) => (caretPosRef.current = e.currentTarget.selectionStart ?? 0)}
            onClick={(e) => (caretPosRef.current = e.currentTarget.selectionStart ?? 0)}
            onKeyDown={(e) => {
              if (e.key === "Enter" && !emojiOpen) {
                e.preventDefault()
                handleSend()
              }
            }}
            placeholder={t("conversations.attachments.caption")}
            aria-label={t("conversations.attachments.caption")}
            className="text-foreground placeholder:text-muted-foreground min-h-9 flex-1 bg-transparent px-2 py-1.5 text-sm outline-none"
          />
          <Popover open={emojiOpen} onOpenChange={setEmojiOpen}>
            <PopoverTrigger
              render={
                <Button
                  type="button"
                  variant="ghost"
                  size="icon-sm"
                  className="text-muted-foreground shrink-0"
                  aria-label={t("conversations.composer.emoji")}
                />
              }
            >
              <SmileIcon />
            </PopoverTrigger>
            <PopoverContent align="end" side="top" className="w-fit p-0">
              <EmojiPicker.Root
                onEmojiSelect={({ emoji }) => insertEmoji(emoji)}
                className="isolate flex h-80 w-72 flex-col"
              >
                <EmojiPicker.Search
                  placeholder={t("conversations.composer.emojiSearch")}
                  className="bg-muted/60 placeholder:text-muted-foreground focus-visible:ring-ring/40 m-2 rounded-lg px-2.5 py-2 text-sm outline-none focus-visible:ring-2"
                />
                <EmojiPicker.Viewport className="relative flex-1 outline-hidden">
                  <EmojiPicker.Loading className="text-muted-foreground absolute inset-0 flex items-center justify-center text-sm">
                    {t("conversations.composer.emojiLoading")}
                  </EmojiPicker.Loading>
                  <EmojiPicker.Empty className="text-muted-foreground absolute inset-0 flex items-center justify-center text-sm">
                    {t("conversations.composer.emojiEmpty")}
                  </EmojiPicker.Empty>
                  <EmojiPicker.List
                    className="pb-2 select-none"
                    components={{
                      CategoryHeader: EmojiCategoryHeader,
                      Row: EmojiRow,
                      Emoji: EmojiButton,
                    }}
                  />
                </EmojiPicker.Viewport>
              </EmojiPicker.Root>
            </PopoverContent>
          </Popover>
        </div>

        {/* Actions: Add (start) — Cancel / Send (end). */}
        <div className="flex items-center justify-between">
          <Button type="button" variant="ghost" onClick={onAddMore}>
            {t("conversations.attachments.dialogAdd")}
          </Button>
          <div className="flex items-center gap-2">
            <Button type="button" variant="ghost" onClick={onCancel}>
              {t("conversations.attachments.dialogCancel")}
            </Button>
            <Button type="button" onClick={handleSend}>
              {t("conversations.attachments.dialogSend")}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

// Emoji-picker slot renderers, hoisted to module scope so they aren't redefined
// per render (frimousse consumes them via the `components` prop).
function EmojiCategoryHeader({ category, ...props }: EmojiPickerListCategoryHeaderProps) {
  return (
    <div className="bg-popover text-muted-foreground px-2 pt-2 pb-1 text-xs font-medium" {...props}>
      {category.label}
    </div>
  )
}

function EmojiRow({ children, ...props }: EmojiPickerListRowProps) {
  return (
    <div className="scroll-my-1 px-1" {...props}>
      {children}
    </div>
  )
}

function EmojiButton({ emoji, ...props }: EmojiPickerListEmojiProps) {
  return (
    <button
      className={cn("flex size-8 items-center justify-center rounded-md text-lg", emoji.isActive && "bg-accent")}
      {...props}
    >
      {emoji.emoji}
    </button>
  )
}

function PreviewItem({
  file,
  single,
  asDocument,
  onRemove,
}: {
  file: File
  single: boolean
  asDocument: boolean
  onRemove: () => void
}) {
  const { t } = useTranslation()
  const image = !asDocument && isImage(file)
  const [blobUrl, setBlobUrl] = useState<string | null>(null)

  useEffect(() => {
    if (!image) return
    const url = URL.createObjectURL(file)
    setBlobUrl(url)
    return () => URL.revokeObjectURL(url)
  }, [file, image])

  return (
    <div className={cn("group/preview relative", single && image && "mx-auto w-fit max-w-full")}>
      <div
        className={cn(
          "border-border bg-muted flex items-center justify-center overflow-hidden rounded-xl border",
          image
            ? single
              ? "max-h-72 w-fit max-w-full"
              : "aspect-square"
            : single
              ? "aspect-video w-full"
              : "aspect-square",
          !image && "flex-col gap-1 p-2"
        )}
      >
        {image && blobUrl ? (
          <img
            src={blobUrl}
            alt={file.name}
            className={cn("object-contain", single ? "max-h-72 w-auto max-w-full" : "size-full")}
          />
        ) : (
          <>
            <FileIcon className="text-muted-foreground size-6 shrink-0" />
            <span className="text-muted-foreground line-clamp-2 w-full px-1 text-center text-xs leading-tight break-all">
              {file.name}
            </span>
            <span className="text-muted-foreground text-[0.625rem] tabular-nums">{formatBytes(file.size)}</span>
          </>
        )}
      </div>
      <button
        type="button"
        onClick={onRemove}
        aria-label={t("conversations.attachments.remove")}
        className="bg-foreground/80 text-background hover:bg-foreground absolute -end-1.5 -top-1.5 flex size-5 items-center justify-center rounded-full opacity-0 shadow-sm transition group-hover/preview:opacity-100 focus-visible:opacity-100"
      >
        <XIcon className="size-3" />
      </button>
    </div>
  )
}
