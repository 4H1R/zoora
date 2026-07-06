import type { GithubCom4H1RZooraInternalDomainUserConnector as UserConnector } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { BellRingIcon, MessageCircleIcon, SendIcon, SmartphoneIcon, Trash2Icon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetConnectorsQueryKey,
  useDeleteConnectorsId,
  usePatchConnectorsId,
  usePostConnectorsPush,
} from "@/api/connectors/connectors"
import { GithubCom4H1RZooraInternalDomainConnectorType as ConnectorType } from "@/api/model"
import { BotLinkDialog } from "@/components/connectors/bot-link-dialog"
import { SmsOtpDialog } from "@/components/connectors/sms-otp-dialog"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogMedia,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import { Spinner } from "@/components/ui/spinner"
import { Switch } from "@/components/ui/switch"
import { isPushConfigured } from "@/config/env"
import { enablePush } from "@/lib/push"
import { cn } from "@/lib/utils"

const META: Record<string, { icon: typeof SendIcon; labelKey: string; tint: string }> = {
  [ConnectorType.ConnectorTelegram]: {
    icon: SendIcon,
    labelKey: "notifications.connectors.telegram",
    tint: "bg-sky-500/10 text-sky-600 dark:text-sky-400",
  },
  [ConnectorType.ConnectorBale]: {
    icon: MessageCircleIcon,
    labelKey: "notifications.connectors.bale",
    tint: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400",
  },
  [ConnectorType.ConnectorSMS]: {
    icon: SmartphoneIcon,
    labelKey: "notifications.connectors.sms",
    tint: "bg-violet-500/10 text-violet-600 dark:text-violet-400",
  },
  [ConnectorType.ConnectorPush]: {
    icon: BellRingIcon,
    labelKey: "notifications.connectors.push",
    tint: "bg-amber-500/10 text-amber-600 dark:text-amber-400",
  },
}

interface ConnectorCardProps {
  type: (typeof ConnectorType)[keyof typeof ConnectorType]
  connector?: UserConnector
}

/** One channel row on the connector settings page: shows connection state,
 * an enable toggle, disconnect action, and a connect flow tailored per channel. */
export function ConnectorCard({ type, connector }: ConnectorCardProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [pushPending, setPushPending] = useState(false)

  const meta = META[type]
  const Icon = meta.icon
  const connected = !!connector
  const verified = !!connector?.verified_at

  const invalidate = () => queryClient.invalidateQueries({ queryKey: getGetConnectorsQueryKey() })

  const patchMutation = usePatchConnectorsId({ mutation: { onSuccess: invalidate } })
  const deleteMutation = useDeleteConnectorsId({
    mutation: {
      onSuccess: () => {
        invalidate()
        setConfirmOpen(false)
      },
    },
  })
  const pushMutation = usePostConnectorsPush({ mutation: { onSuccess: invalidate } })

  const handleToggle = (next: boolean) => {
    if (!connector?.id) return
    patchMutation.mutate({ id: connector.id, data: { enabled: next } })
  }

  const handleConnect = async () => {
    if (type === ConnectorType.ConnectorPush) {
      setPushPending(true)
      try {
        const result = await enablePush()
        if (result.ok) {
          pushMutation.mutate({ data: { token: result.token } })
        } else if (result.reason === "denied") {
          toast.error(t("notifications.connectors.pushDenied"))
        } else {
          toast.error(t("notifications.connectors.pushUnsupported"))
        }
      } finally {
        setPushPending(false)
      }
      return
    }
    setDialogOpen(true)
  }

  // The push card hides its connect button when Firebase is unconfigured.
  const pushUnavailable = type === ConnectorType.ConnectorPush && !isPushConfigured

  return (
    <div className="bg-card ring-foreground/10 flex items-center gap-4 rounded-xl p-4 ring-1">
      <div className={cn("grid size-11 shrink-0 place-items-center rounded-lg", meta.tint)}>
        <Icon className="size-5" />
      </div>

      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="font-medium">{t(meta.labelKey)}</span>
          {connected && (
            <span
              className={cn(
                "rounded-full px-1.5 py-0.5 text-[10px] font-medium",
                verified ? "bg-primary/10 text-primary" : "bg-amber-500/10 text-amber-600 dark:text-amber-400"
              )}
            >
              {verified ? t("notifications.connectors.connected") : t("notifications.connectors.notVerified")}
            </span>
          )}
        </div>
        {connector?.target && (
          <div className="text-muted-foreground truncate font-mono text-xs" dir="ltr">
            {connector.target}
          </div>
        )}
      </div>

      <div className="flex shrink-0 items-center gap-3">
        {connected ? (
          <>
            <Switch
              checked={!!connector?.enabled}
              onCheckedChange={handleToggle}
              disabled={patchMutation.isPending}
              aria-label={t("notifications.connectors.enabled")}
            />
            <Button
              variant="ghost"
              size="icon-sm"
              aria-label={t("notifications.connectors.disconnect")}
              onClick={() => setConfirmOpen(true)}
            >
              <Trash2Icon />
            </Button>
          </>
        ) : (
          !pushUnavailable && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleConnect}
              disabled={pushPending || pushMutation.isPending}
            >
              {(pushPending || pushMutation.isPending) && <Spinner />}
              {t("notifications.connectors.connect")}
            </Button>
          )
        )}
        {pushUnavailable && (
          <span className="text-muted-foreground text-xs">{t("notifications.connectors.pushUnavailable")}</span>
        )}
      </div>

      {(type === ConnectorType.ConnectorTelegram || type === ConnectorType.ConnectorBale) && (
        <BotLinkDialog channel={type} open={dialogOpen} onOpenChange={setDialogOpen} />
      )}
      {type === ConnectorType.ConnectorSMS && <SmsOtpDialog open={dialogOpen} onOpenChange={setDialogOpen} />}

      <AlertDialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <AlertDialogContent onOutsideClick={() => !deleteMutation.isPending && setConfirmOpen(false)}>
          <AlertDialogHeader>
            <AlertDialogMedia className="bg-destructive/10 text-destructive">
              <Trash2Icon />
            </AlertDialogMedia>
            <AlertDialogTitle>{t("notifications.connectors.disconnectConfirm")}</AlertDialogTitle>
            <AlertDialogDescription>{t("notifications.connectors.disconnectConfirmBody")}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteMutation.isPending}>{t("common.cancel")}</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              disabled={deleteMutation.isPending}
              onClick={() => connector?.id && deleteMutation.mutate({ id: connector.id })}
            >
              {deleteMutation.isPending && <Spinner />}
              {t("notifications.connectors.disconnect")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
