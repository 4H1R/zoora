import i18n from "@/i18n"

export function adminHead(pageKey: string) {
  const title = `${i18n.t(pageKey)} | ${i18n.t("common.brandName")} ${i18n.t("admin.panel")}`
  return {
    meta: [{ title }, { name: "description", content: title }],
  }
}
