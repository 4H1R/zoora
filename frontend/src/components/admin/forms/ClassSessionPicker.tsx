import { CheckIcon, ChevronsUpDownIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { useDebounce } from "use-debounce"

import { useGetAdminClasses } from "@/api/admin-classes/admin-classes"
import { useGetClassesIdSessions } from "@/api/classes/classes"
import { Button } from "@/components/ui/button"
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { cn } from "@/lib/utils"

interface ClassPickerProps {
  value?: string
  onChange: (id: string) => void
  placeholder?: string
  disabled?: boolean
}

export function ClassPicker({ value, onChange, placeholder, disabled }: ClassPickerProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState("")
  const [debouncedSearch] = useDebounce(search, 300)

  const { data } = useGetAdminClasses({ search: debouncedSearch || undefined })
  const classes = (data?.status === 200 && data.data.data?.items) || []
  const selected = classes.find((c) => c.id === value)

  return (
    <Popover open={open} onOpenChange={(o) => !disabled && setOpen(o)}>
      <PopoverTrigger
        render={
          <Button
            variant="outline"
            role="combobox"
            disabled={disabled}
            className="w-full justify-between font-normal"
          />
        }
      >
        {selected ? (
          <span className="truncate">{selected.name}</span>
        ) : (
          <span className="text-muted-foreground">{placeholder ?? t("admin.classSessionPicker.classPlaceholder")}</span>
        )}
        <ChevronsUpDownIcon className="text-muted-foreground ms-2 size-4 shrink-0 opacity-50" />
      </PopoverTrigger>
      <PopoverContent className="w-72 p-0" align="start">
        <Command>
          <CommandInput value={search} onValueChange={setSearch} placeholder={t("admin.classes.searchPlaceholder")} />
          <CommandList>
            <CommandEmpty>{t("admin.classes.noResults")}</CommandEmpty>
            <CommandGroup>
              {classes.map((cls) => (
                <CommandItem
                  key={cls.id}
                  value={cls.name ?? ""}
                  onSelect={() => {
                    if (cls.id) {
                      onChange(cls.id)
                      setOpen(false)
                    }
                  }}
                >
                  <CheckIcon className={cn("me-2 size-4", value === cls.id ? "opacity-100" : "opacity-0")} />
                  <div className="min-w-0">
                    <div className="truncate text-sm">{cls.name}</div>
                    {cls.user?.name && <div className="text-muted-foreground truncate text-xs">{cls.user.name}</div>}
                  </div>
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}

interface SessionPickerProps {
  classId?: string
  value?: string
  onChange: (id: string) => void
  placeholder?: string
  disabled?: boolean
}

export function SessionPicker({ classId, value, onChange, placeholder, disabled }: SessionPickerProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState("")
  const [debouncedSearch] = useDebounce(search, 300)

  const isDisabled = disabled || !classId

  const { data } = useGetClassesIdSessions(
    classId ?? "",
    { search: debouncedSearch || undefined },
    { query: { enabled: !!classId } }
  )
  const sessions = (data?.status === 200 && data.data.data?.items) || []
  const selected = sessions.find((s) => s.id === value)

  return (
    <Popover open={open} onOpenChange={(o) => !isDisabled && setOpen(o)}>
      <PopoverTrigger
        render={
          <Button
            variant="outline"
            role="combobox"
            disabled={isDisabled}
            className="w-full justify-between font-normal"
          />
        }
      >
        {selected ? (
          <span className="truncate">{selected.name}</span>
        ) : (
          <span className="text-muted-foreground">
            {classId
              ? (placeholder ?? t("admin.classSessionPicker.sessionPlaceholder"))
              : t("admin.classSessionPicker.selectClassFirst")}
          </span>
        )}
        <ChevronsUpDownIcon className="text-muted-foreground ms-2 size-4 shrink-0 opacity-50" />
      </PopoverTrigger>
      <PopoverContent className="w-72 p-0" align="start">
        <Command>
          <CommandInput value={search} onValueChange={setSearch} placeholder={t("admin.sessions.searchPlaceholder")} />
          <CommandList>
            <CommandEmpty>{t("admin.sessions.noResults")}</CommandEmpty>
            <CommandGroup>
              {sessions.map((s) => (
                <CommandItem
                  key={s.id}
                  value={s.name ?? ""}
                  onSelect={() => {
                    if (s.id) {
                      onChange(s.id)
                      setOpen(false)
                    }
                  }}
                >
                  <CheckIcon className={cn("me-2 size-4", value === s.id ? "opacity-100" : "opacity-0")} />
                  <span className="truncate text-sm">{s.name}</span>
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
