import i18n from "i18next"
import LanguageDetector from "i18next-browser-languagedetector"
import { initReactI18next } from "react-i18next"

import en from "./locales/en.json"
import fa from "./locales/fa.json"
import { configureZodLocale } from "./zod"

export const languages = {
  en: { label: "English", dir: "ltr" as const },
  fa: { label: "فارسی", dir: "rtl" as const },
}

export type Language = keyof typeof languages

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources: {
      en: { translation: en },
      fa: { translation: fa },
    },
    fallbackLng: "en",
    supportedLngs: ["en", "fa"],
    interpolation: {
      escapeValue: false,
    },
    detection: {
      order: ["localStorage", "navigator"],
      caches: ["localStorage"],
    },
  })

// Keep Zod's global error map in sync with the active language so validation
// messages are always localized. See ./zod.ts.
configureZodLocale(i18n.resolvedLanguage ?? i18n.language)
i18n.on("languageChanged", configureZodLocale)

export default i18n
