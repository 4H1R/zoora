import type { GetMediaIdDownloadUrlExpiry } from "@/api/model"
import type { GithubCom4H1RZooraInternalDomainMedia as Media } from "@/api/model"

import { CheckIcon, CopyIcon, LinkIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { useGetMediaIdDownloadUrl } from "@/api/media/media"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Spinner } from "@/components/ui/spinner"
import { cn } from "@/lib/utils"

const EXPIRY_OPTIONS = ["1h", "24h", "7d"] as const

interface ShareDialogProps {
  media: Media | null
  onOpenChange: (open: boolean) => void
}

/** Share a temporary presigned link for a media file with a selectable lifetime. */
export function ShareDialog({ media, onOpenChange }: ShareDialogProps) {
  const { t } = useTranslation()
  const [expiry, setExpiry] = useState<GetMediaIdDownloadUrlExpiry>("1h")
  const [copied, setCopied] = useState(false)

  // Reset per file so a fresh dialog never shows the previous link state.
  useEffect(() => {
    if (media) {
      setExpiry("1h")
      setCopied(false)
    }
  }, [media])

  const { data, isFetching } = useGetMediaIdDownloadUrl(
    media?.id ?? "",
    { expiry },
    { query: { enabled: !!media?.id, staleTime: 0 } }
  )
  const url = (data?.status === 200 && data.data.data?.url) || ""

  const handleCopy = async () => {
    if (!url) return
    await navigator.clipboard.writeText(url)
    setCopied(true)
    toast.success(t("filesPage.share.copied"))
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Dialog open={!!media} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <LinkIcon className="text-primary size-4 shrink-0" />
            {t("filesPage.share.title")}
          </DialogTitle>
          <DialogDescription>{t("filesPage.share.description")}</DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-4">
          <p className="truncate text-sm font-medium" dir="auto">
            {media?.name || media?.file_name}
          </p>

          <div className="flex flex-col gap-1.5">
            <span className="text-muted-foreground text-xs font-medium">{t("filesPage.share.expiry")}</span>
            <div className="grid grid-cols-3 gap-2">
              {EXPIRY_OPTIONS.map((option) => (
                <Button
                  key={option}
                  type="button"
                  size="sm"
                  variant={expiry === option ? "default" : "outline"}
                  className={cn(expiry !== option && "text-muted-foreground")}
                  onClick={() => setExpiry(option)}
                >
                  {t(`filesPage.share.expiry${option}`)}
                </Button>
              ))}
            </div>
          </div>

          <div className="flex items-center gap-2">
            <Input readOnly value={url} placeholder={isFetching ? "…" : ""} className="font-mono text-xs" dir="ltr" />
            <Button type="button" size="icon" variant="outline" disabled={!url} onClick={handleCopy}>
              {isFetching ? <Spinner className="size-4" /> : copied ? <CheckIcon /> : <CopyIcon />}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
