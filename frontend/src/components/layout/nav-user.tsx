import type { GithubCom4H1RZooraInternalDomainUser } from "@/api/model"

import { LogOutIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useLogout } from "@/api/auth/logout"

export function NavUser({ user }: { user?: GithubCom4H1RZooraInternalDomainUser }) {
  const { t } = useTranslation()
  const logout = useLogout()
  const displayName = user?.name ?? user?.username ?? "—"
  const fallback = displayName.slice(0, 2).toUpperCase()
  const role = user?.is_admin ? t("admin.roleAdmin") : (user?.role?.name ?? t("admin.roleMember"))

  return (
    <div className="flex items-center gap-2.5 border-t pt-2.5">
      <div className="grid size-7 shrink-0 place-items-center rounded-full bg-green-100 text-[11px] leading-7 font-semibold text-green-800">
        {fallback}
      </div>
      <div className="min-w-0 flex-1">
        <div className="text-foreground truncate text-xs font-medium">{displayName}</div>
        <div className="text-muted-foreground text-[11px]">{role}</div>
      </div>
      <button
        type="button"
        title={t("admin.logOut")}
        onClick={() => logout.mutate()}
        className="text-muted-foreground hover:bg-muted hover:text-foreground grid size-6.5 shrink-0 cursor-pointer place-items-center rounded-md"
      >
        <LogOutIcon className="size-3.5 stroke-[1.75]" />
      </button>
    </div>
  )
}
