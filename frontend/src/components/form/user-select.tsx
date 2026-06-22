import { CheckIcon, ChevronsUpDownIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { useDebounce } from "use-debounce"

import { useGetAdminUsers } from "@/api/admin-users/admin-users"
import { useGetUsers } from "@/api/users/users"
import { Button } from "@/components/ui/button"
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { cn } from "@/lib/utils"

interface UserSelectProps {
  value?: string
  onChange: (userId: string) => void
  placeholder?: string
  className?: string
  organizationId?: string
  /**
   * "admin" hits the platform-admin /admin/users endpoint (cross-org, requires IsAdmin).
   * "org" (default) hits /users, auto-scoped to the caller's organization (org Manager).
   */
  scope?: "admin" | "org"
}

export function UserSelect({
  value,
  onChange,
  placeholder,
  className,
  organizationId,
  scope = "org",
}: UserSelectProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState("")
  const [debouncedSearch] = useDebounce(search, 300)

  const isAdmin = scope === "admin"

  const adminQuery = useGetAdminUsers(
    {
      search: debouncedSearch || undefined,
      organization_id: organizationId || undefined,
    },
    { query: { enabled: isAdmin } },
  )
  const orgQuery = useGetUsers(
    {
      search: debouncedSearch || undefined,
    },
    { query: { enabled: !isAdmin } },
  )

  const data = isAdmin ? adminQuery.data : orgQuery.data
  const usersData = (data?.status === 200 && data.data.data) || undefined
  const users = usersData?.items ?? []

  const selected = users.find((u) => u.id === value)

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={
          <Button variant="outline" role="combobox" className={cn("w-full justify-between font-normal", className)} />
        }
      >
        {selected ? (
          <span className="truncate">{selected.name}</span>
        ) : (
          <span className="text-muted-foreground">{placeholder ?? t("admin.classes.form.teacherPlaceholder")}</span>
        )}
        <ChevronsUpDownIcon className="text-muted-foreground ms-2 size-4 shrink-0 opacity-50" />
      </PopoverTrigger>
      <PopoverContent className="w-72 p-0" align="start">
        <Command>
          <CommandInput value={search} onValueChange={setSearch} placeholder={t("admin.users.searchPlaceholder")} />
          <CommandList>
            <CommandEmpty>{t("admin.users.noResults")}</CommandEmpty>
            <CommandGroup>
              {users.map((user) => (
                <CommandItem
                  key={user.id}
                  value={user.name ?? ""}
                  onSelect={() => {
                    if (user.id) {
                      onChange(user.id)
                      setOpen(false)
                    }
                  }}
                >
                  <div className="min-w-0 flex-1">
                    <div className="truncate text-sm">{user.name}</div>
                    {user.username && (
                      <div className="text-muted-foreground truncate font-mono text-xs">{user.username}</div>
                    )}
                  </div>
                  <CheckIcon
                    className={cn("ms-2 size-4 shrink-0", value === user.id ? "opacity-100" : "opacity-0")}
                  />
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
