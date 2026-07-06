import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useEffect } from "react"
import { useTranslation } from "react-i18next"

import { NotificationComposer } from "@/components/notifications/notification-composer"
import { SentHistory } from "@/components/notifications/sent-history"
import { PageHeader } from "@/components/page-header"
import { Card } from "@/components/ui/card"
import { useCanAny } from "@/lib/access"
import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute("/_auth/org/notifications/send")({
  head: () => orgHead("notifications.send.title"),
  component: SendNotificationPage,
})

function SendNotificationPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()

  const canSend = useCanAny(["notifications:send", "notifications:send_any"])
  const canSendAny = useCanAny(["notifications:send_any"])
  const mode = canSendAny ? "manager" : "teacher"

  // Page-level gate: users without any send permission are bounced to the inbox.
  useEffect(() => {
    if (!canSend) navigate({ to: "/org/notifications" })
  }, [canSend, navigate])

  if (!canSend) return null

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-8">
      <PageHeader title={t("notifications.send.title")} />

      <Card className="p-6">
        <NotificationComposer mode={mode} />
      </Card>

      <SentHistory />
    </div>
  )
}
