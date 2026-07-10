import { createFileRoute } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

import { useGetConnectors } from "@/api/connectors/connectors"
import { GithubCom4H1RZooraInternalDomainConnectorType as ConnectorType } from "@/api/model"
import { ConnectorCard } from "@/components/connectors/connector-card"
import { PageHeader } from "@/components/page-header"
import { Skeleton } from "@/components/ui/skeleton"
import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute("/_auth/org/settings/notifications")({
  head: () => orgHead("notifications.connectors.title"),
  component: ConnectorSettingsPage,
})

const CHANNELS = [
  ConnectorType.ConnectorTelegram,
  ConnectorType.ConnectorBale,
  ConnectorType.ConnectorSMS,
  ConnectorType.ConnectorPush,
] as const

function ConnectorSettingsPage() {
  const { t } = useTranslation()
  const { data, isLoading } = useGetConnectors()
  const connectors = (data?.status === 200 && data.data.data) || []

  return (
    <div className="mx-auto w-full max-w-3xl">
      <PageHeader title={t("notifications.connectors.title")} />
      <p className="text-muted-foreground mt-1.5 text-sm">{t("notifications.connectors.description")}</p>

      <div className="mt-6 flex flex-col gap-3">
        {isLoading
          ? CHANNELS.map((c) => <Skeleton key={c} className="h-[75px] w-full rounded-xl" />)
          : CHANNELS.map((channel) => (
              <ConnectorCard key={channel} type={channel} connector={connectors.find((c) => c.type === channel)} />
            ))}
      </div>
    </div>
  )
}
