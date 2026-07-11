import { createFileRoute } from "@tanstack/react-router"

import { AccountSettings } from "@/components/account/account-settings"
import { adminHead } from "@/lib/admin-head"

export const Route = createFileRoute("/_admin/admin/account")({
  head: () => adminHead("account.title"),
  component: () => <AccountSettings showProfile={false} />,
})
