import { useQueryClient } from "@tanstack/react-query"
import { CheckCircle2Icon, ExternalLinkIcon, Loader2Icon } from "lucide-react"
import { QRCodeSVG } from "qrcode.react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetConnectorsQueryKey,
  useGetConnectors,
  usePostConnectorsBaleLink,
  usePostConnectorsTelegramLink,
} from "@/api/connectors/connectors"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Spinner } from "@/components/ui/spinner"

interface BotLinkDialogProps {
  channel: "telegram" | "bale"
  open: boolean
  onOpenChange: (open: boolean) => void
}

/** Deep-link + QR flow for the Telegram/Bale bots. Requests a one-time link on
 * open, then polls the connector list until the user presses Start in the bot. */
export function BotLinkDialog({ channel, open, onOpenChange }: BotLinkDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [deepLink, setDeepLink] = useState<string | undefined>()
  const [linked, setLinked] = useState(false)
  const [alreadyConnected, setAlreadyConnected] = useState(false)

  const telegramLink = usePostConnectorsTelegramLink()
  const baleLink = usePostConnectorsBaleLink()
  const mutation = channel === "telegram" ? telegramLink : baleLink

  // Request a fresh link each time the dialog opens.
  useEffect(() => {
    if (!open) {
      setDeepLink(undefined)
      setLinked(false)
      return
    }
    mutation.mutate(undefined, {
      onSuccess: (res) => {
        if (res.status === 200) setDeepLink(res.data.data?.deep_link)
      },
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, channel])

  // Poll connectors while open; close once this channel becomes connected.
  const { data } = useGetConnectors({
    query: { enabled: open && !linked, refetchInterval: 3000 },
  })
  const connectors = (data?.status === 200 && data.data.data) || []
  const connected = connectors.some((c) => c.type === channel)

  // Snapshot whether the channel was already connected when we opened, so a
  // stale prior link doesn't instantly report success.
  useEffect(() => {
    if (open && data) setAlreadyConnected((prev) => prev || connected)
    if (!open) setAlreadyConnected(false)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, data])

  useEffect(() => {
    if (open && connected && !alreadyConnected && !linked) {
      setLinked(true)
      queryClient.invalidateQueries({ queryKey: getGetConnectorsQueryKey() })
      toast.success(t("notifications.connectors.connected"))
      const id = setTimeout(() => onOpenChange(false), 1200)
      return () => clearTimeout(id)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [connected, alreadyConnected, open, linked])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>{t(`notifications.connectors.${channel}`)}</DialogTitle>
          <DialogDescription>{t("notifications.connectors.linkHint")}</DialogDescription>
        </DialogHeader>

        {linked ? (
          <div className="flex flex-col items-center gap-3 py-8">
            <CheckCircle2Icon className="text-primary size-12" />
            <p className="font-medium">{t("notifications.connectors.connected")}</p>
          </div>
        ) : mutation.isPending || !deepLink ? (
          <div className="flex justify-center py-12">
            <Spinner />
          </div>
        ) : (
          <div className="flex flex-col items-center gap-5 py-2">
            <div className="rounded-xl border bg-white p-3">
              <QRCodeSVG value={deepLink} size={168} />
            </div>

            <Button className="w-full" render={<a href={deepLink} target="_blank" rel="noopener noreferrer" />}>
              <ExternalLinkIcon />
              {t("notifications.connectors.openLink")}
            </Button>

            <div className="text-muted-foreground flex items-center gap-2 text-xs">
              <Loader2Icon className="size-3.5 animate-spin" />
              <span>{t("notifications.connectors.waiting")}</span>
            </div>
            <p className="text-muted-foreground text-center text-xs">{t("notifications.connectors.linkExpires")}</p>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
