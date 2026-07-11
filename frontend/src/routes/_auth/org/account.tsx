import { createFileRoute } from "@tanstack/react-router"

import { AccountSettings } from "@/components/account/account-settings"
import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute("/_auth/org/account")({
  head: () => orgHead("account.title"),
  component: AccountSettings,
})
