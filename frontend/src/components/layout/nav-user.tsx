import type { GithubCom4H1RZooraInternalDomainUser } from "@/api/model"

import { Link } from "@tanstack/react-router"
import { ChevronsUpDownIcon, LogOutIcon, SettingsIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useLogout } from "@/api/auth/logout"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { useRoleName } from "@/lib/permissions"

export function NavUser({
  user,
  accountTo = "/org/account",
}: {
  user?: GithubCom4H1RZooraInternalDomainUser
  accountTo?: "/org/account" | "/admin/account"
}) {
  const { t } = useTranslation()
  const logout = useLogout()
  const roleName = useRoleName()
  const displayName = user?.name ?? user?.username ?? "—"
  const fallback = displayName.slice(0, 2).toUpperCase()
  const role = user?.is_admin
    ? t("admin.roleAdmin")
    : user?.role?.name
      ? roleName(user.role.name)
      : t("admin.roleMember")

  return (
    <div className="border-t pt-2.5">
      <DropdownMenu>
        <DropdownMenuTrigger
          render={
            <button
              type="button"
              className="hover:bg-muted data-[popup-open]:bg-muted -mx-1 flex w-[calc(100%+0.5rem)] cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 text-start transition-colors"
            />
          }
        >
          <div className="grid size-7 shrink-0 place-items-center rounded-full bg-green-100 text-[11px] leading-7 font-semibold text-green-800">
            {fallback}
          </div>
          <div className="min-w-0 flex-1">
            <div className="text-foreground truncate text-xs font-medium">{displayName}</div>
            <div className="text-muted-foreground text-[11px]">{role}</div>
          </div>
          <ChevronsUpDownIcon className="text-muted-foreground size-3.5 shrink-0" />
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" side="top" className="min-w-52">
          <DropdownMenuItem render={<Link to={accountTo} />}>
            <SettingsIcon className="size-4" />
            {t("account.menuAction")}
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem variant="destructive" onClick={() => logout.mutate()}>
            <LogOutIcon className="size-4" />
            {t("admin.logOut")}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
