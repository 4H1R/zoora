import type { Language } from "@/i18n"

import { ChevronsUpDown } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { Command, CommandItem, CommandList } from "@/components/ui/command"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { languages } from "@/i18n"

const flags: Record<Language, string> = {
  en: "🇺🇸",
  fa: "🇮🇷",
}

export function LanguageSwitcher() {
  const { i18n } = useTranslation()
  const [open, setOpen] = useState(false)

  const current: Language = (i18n.language as Language) in languages ? (i18n.language as Language) : "en"

  const select = (lang: Language) => {
    i18n.changeLanguage(lang)
    setOpen(false)
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={<Button variant="outline" size="sm" role="combobox" aria-expanded={open} className="gap-1.5" />}
      >
        <span>{flags[current]}</span>
        <span>{languages[current].label}</span>
        <ChevronsUpDown className="size-3.5 shrink-0 opacity-50" />
      </PopoverTrigger>
      <PopoverContent className="w-40 p-0" align="end">
        <Command>
          <CommandList>
            {(Object.keys(languages) as Language[]).map((lang) => (
              <CommandItem key={lang} value={lang} onSelect={() => select(lang)} data-checked={lang === current}>
                <span>{flags[lang]}</span>
                <span>{languages[lang].label}</span>
              </CommandItem>
            ))}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
