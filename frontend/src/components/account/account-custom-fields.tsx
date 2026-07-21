import type { GithubCom4H1RZooraInternalDomainVisibleCustomField as VisibleCustomField } from "@/api/model"

import { useTranslation } from "react-i18next"

import { useGetUsersIdCustomFields } from "@/api/custom-fields/custom-fields"
import { useGetUsersMe } from "@/api/users/users"

export function AccountCustomFields() {
  const { t } = useTranslation()
  const { data: me } = useGetUsersMe()
  const userId = me?.status === 200 ? (me.data.data?.id ?? "") : ""

  const { data } = useGetUsersIdCustomFields(userId, { query: { enabled: !!userId } })
  const fields: VisibleCustomField[] = (data?.status === 200 && (data.data.data as VisibleCustomField[])) || []

  if (fields.length === 0) return null

  return (
    <section className="bg-card ring-foreground/10 rounded-2xl border p-6 ring-1 sm:p-8">
      <h2 className="mb-5 text-lg font-semibold tracking-tight">{t("org.customFields.title")}</h2>
      <dl className="grid grid-cols-1 gap-5 sm:grid-cols-2">
        {fields.map((f) => (
          <div key={f.field_id}>
            <dt className="text-muted-foreground text-sm">{f.label}</dt>
            <dd className="mt-1 font-medium">{renderValue(f)}</dd>
          </div>
        ))}
      </dl>
    </section>
  )
}

function renderValue(f: VisibleCustomField): string {
  const v = f.value
  if (v === null || v === undefined || v === "") return "—"
  if (f.field_type === "boolean") return v ? "✓" : "✗"
  return String(v)
}
