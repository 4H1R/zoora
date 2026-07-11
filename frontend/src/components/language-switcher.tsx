import type { Language } from "@/i18n"

import { ChevronsUpDown } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { languages } from "@/i18n"

const flags: Record<Language, string> = {
  en: "🇺🇸",
  fa: "🇮🇷",
}

export function LanguageSwitcher() {
  const { i18n } = useTranslation()

  const current: Language = (i18n.language as Language) in languages ? (i18n.language as Language) : "en"

  return (
    <DropdownMenu>
      <DropdownMenuTrigger render={<Button variant="outline" size="sm" role="combobox" className="gap-1.5" />}>
        <span>{flags[current]}</span>
        <span>{languages[current].label}</span>
        <ChevronsUpDown className="size-3.5 shrink-0 opacity-50" />
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-40" align="end">
        <DropdownMenuRadioGroup value={current} onValueChange={(lang) => i18n.changeLanguage(lang as Language)}>
          {(Object.keys(languages) as Language[]).map((lang) => (
            <DropdownMenuRadioItem key={lang} value={lang}>
              <span>{flags[lang]}</span>
              <span>{languages[lang].label}</span>
            </DropdownMenuRadioItem>
          ))}
        </DropdownMenuRadioGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
