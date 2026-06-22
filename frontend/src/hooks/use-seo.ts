import { useEffect } from "react"

import i18n from "@/i18n"

/**
 * Imperatively sets the document title and meta description for screens
 * rendered outside the route tree's `head` API (e.g. `notFoundComponent`,
 * `errorComponent`), where `<HeadContent />` cannot pick up route head data.
 */
export function useSeo(titleKey: string, descriptionKey: string) {
  const title = `${i18n.t(titleKey)} | ${i18n.t("common.brandName")}`
  const description = i18n.t(descriptionKey)

  useEffect(() => {
    document.title = title

    let meta = document.querySelector<HTMLMetaElement>('meta[name="description"]')
    if (!meta) {
      meta = document.createElement("meta")
      meta.name = "description"
      document.head.appendChild(meta)
    }
    meta.content = description
  }, [title, description])
}
