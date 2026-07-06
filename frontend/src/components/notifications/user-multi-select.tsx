import { CheckIcon, ChevronsUpDownIcon, XIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { useDebounce } from "use-debounce"

import { useGetAdminUsers } from "@/api/admin-users/admin-users"
import { useGetClassesIdMembers } from "@/api/classes/classes"
import { useGetUsers } from "@/api/users/users"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { cn } from "@/lib/utils"

type Option = { id: string; name: string; username?: string }

interface UserMultiSelectProps {
  value: string[]
  onChange: (ids: string[]) => void
  /** "admin" hits the cross-org endpoint; "org" is caller-scoped. */
  scope?: "admin" | "org"
  /** Admin-only filter to a single organization. */
  organizationId?: string
  /** When set, the pool is restricted to a class's members (teacher flow). */
  classId?: string
  className?: string
}

/** Searchable multi-select over org / admin users, or the members of a given
 * class. Emits the selected user ids and renders removable chips. */
export function UserMultiSelect({
  value,
  onChange,
  scope = "org",
  organizationId,
  classId,
  className,
}: UserMultiSelectProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState("")
  const [debouncedSearch] = useDebounce(search, 300)

  const byClass = !!classId
  const isAdmin = scope === "admin" && !byClass

  const membersQuery = useGetClassesIdMembers(
    classId ?? "",
    { search: debouncedSearch || undefined, page_size: 50 },
    { query: { enabled: byClass && !!classId } }
  )
  const adminQuery = useGetAdminUsers(
    { search: debouncedSearch || undefined, organization_id: organizationId || undefined },
    { query: { enabled: isAdmin } }
  )
  const orgQuery = useGetUsers(
    { search: debouncedSearch || undefined },
    { query: { enabled: !byClass && !isAdmin } }
  )

  let options: Option[] = []
  if (byClass) {
    const data = (membersQuery.data?.status === 200 && membersQuery.data.data.data) || undefined
    options = (data?.items ?? [])
      .map((m) => m.user)
      .filter((u): u is NonNullable<typeof u> => !!u?.id)
      .map((u) => ({ id: u.id as string, name: u.name ?? "", username: u.username }))
  } else {
    const q = isAdmin ? adminQuery.data : orgQuery.data
    const data = (q?.status === 200 && q.data.data) || undefined
    options = (data?.items ?? [])
      .filter((u) => !!u.id)
      .map((u) => ({ id: u.id as string, name: u.name ?? "", username: u.username }))
  }

  const selectedOptions = value
    .map((id) => options.find((o) => o.id === id) ?? { id, name: id })
    .filter(Boolean) as Option[]

  const toggle = (id: string) => {
    if (value.includes(id)) onChange(value.filter((v) => v !== id))
    else onChange([...value, id])
  }

  return (
    <div className={cn("flex flex-col gap-2", className)}>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger
          render={
            <Button variant="outline" role="combobox" className="w-full justify-between font-normal" />
          }
        >
          <span className={cn(value.length === 0 && "text-muted-foreground")}>
            {value.length > 0
              ? t("notifications.send.selectUsers") + ` (${value.length})`
              : t("notifications.send.selectUsers")}
          </span>
          <ChevronsUpDownIcon className="text-muted-foreground ms-2 size-4 shrink-0 opacity-50" />
        </PopoverTrigger>
        <PopoverContent className="w-72 p-0" align="start">
          <Command shouldFilter={false}>
            <CommandInput
              value={search}
              onValueChange={setSearch}
              placeholder={t("common.search")}
            />
            <CommandList>
              <CommandEmpty>{t("common.noResults")}</CommandEmpty>
              <CommandGroup>
                {options.map((user) => {
                  const checked = value.includes(user.id)
                  return (
                    <CommandItem key={user.id} value={user.id} onSelect={() => toggle(user.id)}>
                      <CheckIcon
                        className={cn("me-2 size-4 shrink-0", checked ? "opacity-100" : "opacity-0")}
                      />
                      <div className="min-w-0 flex-1">
                        <div className="truncate text-sm">{user.name}</div>
                        {user.username && (
                          <div className="text-muted-foreground truncate font-mono text-xs">
                            {user.username}
                          </div>
                        )}
                      </div>
                    </CommandItem>
                  )
                })}
              </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>

      {selectedOptions.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {selectedOptions.map((user) => (
            <Badge key={user.id} variant="secondary" className="gap-1 ps-2 pe-1">
              <span className="max-w-40 truncate">{user.name}</span>
              <button
                type="button"
                aria-label={user.name}
                className="hover:bg-foreground/10 -me-0.5 grid size-4 place-items-center rounded-full"
                onClick={() => toggle(user.id)}
              >
                <XIcon className="size-3" />
              </button>
            </Badge>
          ))}
        </div>
      )}
    </div>
  )
}
