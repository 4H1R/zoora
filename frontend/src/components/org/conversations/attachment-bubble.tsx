import type { ChatMessage, LocalAttachment } from "./lib/messages"

import { AlertCircleIcon, DownloadIcon, FileIcon, XIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { formatBytes } from "@/components/org/files/utils"
import { Dialog, DialogContent, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { cn } from "@/lib/utils"

import { mediaIdStrings } from "./lib/messages"
import { ProgressRing } from "./attachment-progress-ring"
import { isImage as isImageType } from "./upload/compress"
import { blurhashToDataUrl } from "./upload/blurhash"
import { useMediaMeta, useMediaUrl } from "./use-media-attachment"
import { useSendAttachments } from "./use-send-attachments"

interface AttachmentBubbleProps {
  message: ChatMessage
  convId: string
  isOwn: boolean
}

/**
 * Media slot for a message bubble. Prefers the client-only `_attachments`
 * previews while a bubble is optimistic (blob thumbnails + progress rings),
 * then falls back to the server-confirmed `media_ids` (resolved lazily to
 * presigned display URLs). Renders nothing when a message has no media.
 */
export function AttachmentBubble({ message, convId, isOwn }: AttachmentBubbleProps) {
  // "Send as a document" forces every attachment to render as a file chip,
  // regardless of mime type (Telegram parity).
  const asDocument = message.as_document ?? false
  const local = message._attachments
  if (local && local.length > 0) {
    return (
      <LocalAttachments
        attachments={local}
        convId={convId}
        msgId={message.id ?? ""}
        isOwn={isOwn}
        asDocument={asDocument}
      />
    )
  }

  const ids = mediaIdStrings(message)
  if (ids.length > 0) return <ConfirmedAttachments ids={ids} isOwn={isOwn} asDocument={asDocument} />

  return null
}

// Column count for the attachment grid: a lone item goes full-bleed, everything
// else tiles two-up.
function gridClass(count: number): string {
  return count <= 1 ? "grid-cols-1 max-w-64" : "grid-cols-2 max-w-80"
}

/* -------------------------------------------------------------------------- */
/*  Optimistic (pre-confirmation) attachments                                  */
/* -------------------------------------------------------------------------- */

function LocalAttachments({
  attachments,
  convId,
  msgId,
  isOwn,
  asDocument,
}: {
  attachments: LocalAttachment[]
  convId: string
  msgId: string
  isOwn: boolean
  asDocument: boolean
}) {
  const { cancelAttachment } = useSendAttachments(convId)

  return (
    <div className={cn("mb-1 grid gap-1", gridClass(attachments.length))}>
      {attachments.map((a) => {
        const image = !asDocument && (!!a.blobUrl || isImageType({ type: a.contentType }))
        const single = attachments.length <= 1
        return image ? (
          <LocalImageCell
            key={a.localId}
            att={a}
            single={single}
            onCancel={() => cancelAttachment(msgId, a.localId)}
          />
        ) : (
          <div key={a.localId} className={cn(attachments.length > 1 && "col-span-2")}>
            <FileChip
              name={a.name}
              size={a.size}
              isOwn={isOwn}
              status={a.status}
              progress={a.progress}
              onCancel={() => cancelAttachment(msgId, a.localId)}
            />
          </div>
        )
      })}
    </div>
  )
}

function LocalImageCell({
  att,
  single,
  onCancel,
}: {
  att: LocalAttachment
  single: boolean
  onCancel: () => void
}) {
  const { t } = useTranslation()
  const uploading = att.status === "uploading"
  const errored = att.status === "error"

  return (
    <div className={cn("group/att relative overflow-hidden rounded-xl", single ? "" : "aspect-square")}>
      <BlurhashImage
        src={att.blobUrl ?? undefined}
        blurhash={att.blurhash}
        alt={att.name}
        className={cn(single ? "max-h-80 w-full" : "size-full", uploading && "brightness-90")}
      />

      {/* Upload overlay: dim + centered progress ring + cancel X. */}
      {uploading && (
        <div className="absolute inset-0 flex items-center justify-center bg-black/25">
          <ProgressRing value={att.progress} className="size-11" />
          <button
            type="button"
            onClick={onCancel}
            aria-label={t("conversations.attachments.cancel")}
            className="absolute inset-0 m-auto flex size-11 items-center justify-center"
          >
            <XIcon className="size-4 text-white" />
          </button>
        </div>
      )}

      {errored && (
        <div className="bg-destructive/70 absolute inset-0 flex items-center justify-center">
          <AlertCircleIcon className="size-6 text-white" />
        </div>
      )}
    </div>
  )
}

/* -------------------------------------------------------------------------- */
/*  Confirmed (server-persisted) attachments                                   */
/* -------------------------------------------------------------------------- */

function ConfirmedAttachments({ ids, isOwn, asDocument }: { ids: string[]; isOwn: boolean; asDocument: boolean }) {
  return (
    <div className={cn("mb-1 grid gap-1", gridClass(ids.length))}>
      {ids.map((id) => (
        <ConfirmedAttachment key={id} id={id} isOwn={isOwn} multiple={ids.length > 1} asDocument={asDocument} />
      ))}
    </div>
  )
}

function ConfirmedAttachment({
  id,
  isOwn,
  multiple,
  asDocument,
}: {
  id: string
  isOwn: boolean
  multiple: boolean
  asDocument: boolean
}) {
  const { data: meta } = useMediaMeta(id)
  const { data: url } = useMediaUrl(id)

  // Metadata still resolving — reserve a square so the grid doesn't jump.
  if (meta === undefined) {
    return <div className={cn("bg-muted animate-pulse rounded-xl", multiple ? "aspect-square" : "aspect-video")} />
  }

  const isImg = !asDocument && (meta?.mime_type ?? "").startsWith("image/")

  if (isImg) {
    return (
      <LightboxImage
        src={url ?? undefined}
        alt={meta?.name ?? meta?.file_name ?? ""}
        className={cn("overflow-hidden rounded-xl", multiple ? "aspect-square" : "max-h-80 w-full")}
      />
    )
  }

  return (
    <div className={cn(multiple && "col-span-2")}>
      <FileChip
        name={meta?.name ?? meta?.file_name ?? id}
        size={meta?.size}
        isOwn={isOwn}
        downloadUrl={url ?? undefined}
      />
    </div>
  )
}

/* -------------------------------------------------------------------------- */
/*  Shared primitives                                                           */
/* -------------------------------------------------------------------------- */

// Decode a blurhash to a data URL once per hash (canvas work kept out of render).
function useBlurhashDataUrl(hash?: string | null): string | null {
  const [url, setUrl] = useState<string | null>(null)
  useEffect(() => {
    setUrl(hash ? blurhashToDataUrl(hash) : null)
  }, [hash])
  return url
}

/**
 * An image that fades in over its blurhash placeholder (or a muted box). Used
 * for both blob previews and resolved download URLs.
 */
function BlurhashImage({
  src,
  blurhash,
  alt,
  className,
  onClick,
}: {
  src?: string
  blurhash?: string | null
  alt: string
  className?: string
  onClick?: () => void
}) {
  const [loaded, setLoaded] = useState(false)
  const placeholder = useBlurhashDataUrl(blurhash)

  return (
    <div className={cn("bg-muted relative overflow-hidden", className)}>
      {placeholder && (
        <img
          src={placeholder}
          alt=""
          aria-hidden
          className={cn(
            "absolute inset-0 size-full scale-110 object-cover blur-md transition-opacity duration-500",
            loaded && "opacity-0"
          )}
        />
      )}
      {src && (
        <img
          src={src}
          alt={alt}
          onClick={onClick}
          onLoad={() => setLoaded(true)}
          className={cn(
            "relative size-full object-cover transition-opacity duration-500",
            onClick && "cursor-zoom-in",
            loaded ? "opacity-100" : "opacity-0"
          )}
        />
      )}
    </div>
  )
}

/** A confirmed image thumbnail that opens a full-size lightbox on click. */
function LightboxImage({ src, alt, className }: { src?: string; alt: string; className?: string }) {
  const { t } = useTranslation()

  if (!src) return <div className={cn("bg-muted animate-pulse", className)} />

  return (
    <Dialog>
      <DialogTrigger
        render={<button type="button" aria-label={t("conversations.attachments.open")} className={className} />}
      >
        <img src={src} alt={alt} className="size-full object-cover" />
      </DialogTrigger>
      <DialogContent className="border-0 bg-transparent p-0 shadow-none sm:max-w-3xl">
        <DialogTitle className="sr-only">{alt || t("conversations.attachments.image")}</DialogTitle>
        <img src={src} alt={alt} className="max-h-[80vh] w-full rounded-lg object-contain" />
      </DialogContent>
    </Dialog>
  )
}

/** A non-image attachment rendered as a labelled chip. */
function FileChip({
  name,
  size,
  isOwn,
  downloadUrl,
  status,
  progress = 0,
  onCancel,
}: {
  name: string
  size?: number
  isOwn: boolean
  downloadUrl?: string
  status?: LocalAttachment["status"]
  progress?: number
  onCancel?: () => void
}) {
  const { t } = useTranslation()
  const uploading = status === "uploading"
  const errored = status === "error"

  const body = (
    <>
      <span
        className={cn(
          "relative flex size-9 shrink-0 items-center justify-center rounded-lg",
          isOwn ? "bg-primary-foreground/15" : "bg-background"
        )}
      >
        {uploading ? (
          <ProgressRing value={progress} className="text-foreground size-9" />
        ) : errored ? (
          <AlertCircleIcon className="text-destructive size-4" />
        ) : downloadUrl ? (
          <DownloadIcon className="size-4" />
        ) : (
          <FileIcon className="size-4" />
        )}
      </span>
      <span className="flex min-w-0 flex-col text-start">
        <span className="truncate text-sm font-medium">{name}</span>
        {size !== undefined && (
          <span className={cn("text-xs tabular-nums", isOwn ? "text-primary-foreground/70" : "text-muted-foreground")}>
            {formatBytes(size)}
          </span>
        )}
      </span>
      {uploading && onCancel && (
        <button
          type="button"
          onClick={(e) => {
            e.preventDefault()
            onCancel()
          }}
          aria-label={t("conversations.attachments.cancel")}
          className="text-muted-foreground hover:text-foreground ms-auto shrink-0"
        >
          <XIcon className="size-4" />
        </button>
      )}
    </>
  )

  const shell = cn(
    "flex items-center gap-2.5 rounded-xl px-2.5 py-2 transition",
    isOwn ? "bg-primary-foreground/10" : "bg-background/60 border-border border"
  )

  if (downloadUrl && !uploading) {
    return (
      <a href={downloadUrl} target="_blank" rel="noopener noreferrer" download={name} className={cn(shell, "hover:bg-muted")}>
        {body}
      </a>
    )
  }

  return <div className={shell}>{body}</div>
}
