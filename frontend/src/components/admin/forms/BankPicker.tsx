import { CheckIcon, ChevronsUpDownIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { useDebounce } from "use-debounce"

import { useGetQuestionBanks } from "@/api/question-banks/question-banks"
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

interface BankPickerProps {
  value?: string
  onChange: (id: string) => void
  placeholder?: string
}

export function BankPicker({ value, onChange, placeholder }: BankPickerProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState("")
  const [debouncedSearch] = useDebounce(search, 300)

  const { data } = useGetQuestionBanks({ search: debouncedSearch || undefined })
  const banks = (data?.status === 200 && data.data.data?.items) || []
  const selected = banks.find((b) => b.id === value)

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={
          <Button
            variant="outline"
            role="combobox"
            className="w-full justify-between font-normal"
          />
        }
      >
        {selected ? (
          <span className="truncate">{selected.name}</span>
        ) : (
          <span className="text-muted-foreground">
            {placeholder ?? t("admin.bankPicker.placeholder")}
          </span>
        )}
        <ChevronsUpDownIcon className="text-muted-foreground ms-2 size-4 shrink-0 opacity-50" />
      </PopoverTrigger>
      <PopoverContent className="w-72 p-0" align="start">
        <Command>
          <CommandInput
            value={search}
            onValueChange={setSearch}
            placeholder={t("admin.questionBanks.searchPlaceholder")}
          />
          <CommandList>
            <CommandEmpty>{t("admin.questionBanks.noResults")}</CommandEmpty>
            <CommandGroup>
              {banks.map((b) => (
                <CommandItem
                  key={b.id}
                  value={b.name ?? ""}
                  onSelect={() => {
                    if (b.id) {
                      onChange(b.id)
                      setOpen(false)
                    }
                  }}
                >
                  <CheckIcon
                    className={cn(
                      "me-2 size-4",
                      value === b.id ? "opacity-100" : "opacity-0"
                    )}
                  />
                  <div className="min-w-0">
                    <div className="truncate text-sm">{b.name}</div>
                    {b.description && (
                      <div className="text-muted-foreground truncate text-xs">
                        {b.description}
                      </div>
                    )}
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
