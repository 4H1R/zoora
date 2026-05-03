import { ArrowDownIcon, ArrowUpDownIcon, ArrowUpIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"

export interface SortOption {
  id: string
  label: string
}

interface SortPickerProps {
  options: SortOption[]
  value?: { id: string; desc: boolean }
  onChange: (value: { id: string; desc: boolean } | undefined) => void
  label: string
  searchPlaceholder?: string
  emptyLabel?: string
}

export function SortPicker({ options, value, onChange, label, searchPlaceholder, emptyLabel }: SortPickerProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)

  const active = options.find((o) => o.id === value?.id)

  function handleSelect(id: string) {
    if (id === "__default__") {
      onChange(undefined)
      setOpen(false)
      return
    }
    onChange({ id, desc: value?.id === id ? !value.desc : false })
    setOpen(false)
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={
          <Button variant="outline" size="sm" className="h-8 gap-1.5 px-2.5 text-xs font-medium">
            <ArrowUpDownIcon className="size-3.5" />
            {label}:{" "}
            <span className="text-foreground font-semibold">
              {value ? (active?.label ?? options[0]?.label) : t("common.default")}
            </span>
            {value && (value.desc ? <ArrowDownIcon className="size-3" /> : <ArrowUpIcon className="size-3" />)}
          </Button>
        }
      />
      <PopoverContent align="end" className="w-44 p-0">
        <Command>
          <CommandInput placeholder={searchPlaceholder ?? t("common.search")} />
          <CommandList>
            <CommandEmpty>{emptyLabel ?? t("common.noResults")}</CommandEmpty>
            <CommandGroup>
              <CommandItem value="__default__" data-checked={!value} onSelect={() => handleSelect("__default__")}>
                <span className="flex-1">{t("common.default")}</span>
              </CommandItem>
              {options.map((opt) => (
                <CommandItem
                  key={opt.id}
                  value={opt.id}
                  data-checked={value?.id === opt.id}
                  onSelect={() => handleSelect(opt.id)}
                >
                  <span className="flex-1">{opt.label}</span>
                  {value?.id === opt.id &&
                    (value.desc ? (
                      <ArrowDownIcon className="text-muted-foreground size-3" />
                    ) : (
                      <ArrowUpIcon className="text-muted-foreground size-3" />
                    ))}
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
