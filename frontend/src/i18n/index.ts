import i18n from "i18next"
import LanguageDetector from "i18next-browser-languagedetector"
import { initReactI18next } from "react-i18next"

import en from "./locales/en.json"
import fa from "./locales/fa.json"

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

export default i18n
