import { CheckIcon, ChevronsUpDownIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { useDebounce } from "use-debounce"

import { useGetAdminOrganizations } from "@/api/admin-organizations/admin-organizations"
import { Button } from "@/components/ui/button"
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { getInitials } from "@/components/user-avatar"
import { getEntityColor } from "@/lib/data-table"
import { cn } from "@/lib/utils"

interface OrganizationSelectProps {
  value?: string
  onChange: (orgId: string) => void
  placeholder?: string
  className?: string
}

export function OrganizationSelect({ value, onChange, placeholder, className }: OrganizationSelectProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState("")
  const [debouncedSearch] = useDebounce(search, 300)

  const { data } = useGetAdminOrganizations({ search: debouncedSearch || undefined })
  const orgsData = (data?.status === 200 && data.data.data) || undefined
  const organizations = orgsData?.items ?? []

  const selected = organizations.find((o) => o.id === value)

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={
          <Button variant="outline" role="combobox" className={cn("w-full justify-between font-normal", className)} />
        }
      >
        {selected ? (
          <div className="flex min-w-0 items-center gap-2">
            <div
              className={cn(
                "flex size-5 shrink-0 items-center justify-center rounded-[5px] text-[10px] font-semibold text-white",
                getEntityColor(selected.name)
              )}
            >
              {getInitials(selected.name)}
            </div>
            <span className="truncate">{selected.name}</span>
          </div>
        ) : (
          <span className="text-muted-foreground">{placeholder ?? t("admin.orgs.switcher.search")}</span>
        )}
        <ChevronsUpDownIcon className="text-muted-foreground ms-2 size-4 shrink-0 opacity-50" />
      </PopoverTrigger>
      <PopoverContent className="w-72 p-0" align="start">
        <Command>
          <CommandInput value={search} onValueChange={setSearch} placeholder={t("admin.orgs.switcher.search")} />
          <CommandList>
            <CommandEmpty>{t("admin.organizations.noResults", "No organizations found")}</CommandEmpty>
            <CommandGroup>
              {organizations.map((org) => (
                <CommandItem
                  key={org.id}
                  value={org.name}
                  onSelect={() => {
                    if (org.id) {
                      onChange(org.id)
                      setOpen(false)
                      setSearch("")
                    }
                  }}
                >
                  <CheckIcon className={cn("me-2 size-4 shrink-0", value === org.id ? "opacity-100" : "opacity-0")} />
                  <div
                    className={cn(
                      "me-2 flex size-5 shrink-0 items-center justify-center rounded-[5px] text-[10px] font-semibold text-white",
                      getEntityColor(org.name)
                    )}
                  >
                    {getInitials(org.name)}
                  </div>
                  <span className="text-sm">{org.name}</span>
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
