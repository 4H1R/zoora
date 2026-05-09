import i18n from "@/i18n"

export function orgHead(pageKey?: string) {
  const brand = i18n.t("common.brandName")
  const title = pageKey ? `${i18n.t(pageKey)} | ${brand}` : brand
  return {
    meta: [{ title }, { name: "description", content: title }],
  }
}
