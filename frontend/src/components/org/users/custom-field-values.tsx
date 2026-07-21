import type { GithubCom4H1RZooraInternalDomainUserCustomFieldDefinition as CustomFieldDefinition } from "@/api/model"

import { useTranslation } from "react-i18next"

import { useGetCustomFieldDefinitions } from "@/api/custom-fields/custom-fields"
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"

export type CustomFieldValues = Record<string, unknown>

export function CustomFieldValuesEditor({
  value,
  onChange,
}: {
  value: CustomFieldValues
  onChange: (next: CustomFieldValues) => void
}) {
  const { t } = useTranslation()
  const { data } = useGetCustomFieldDefinitions()
  const defs: CustomFieldDefinition[] = (data?.status === 200 && (data.data.data as CustomFieldDefinition[])) || []

  if (defs.length === 0) return null

  const set = (id: string, v: unknown) => onChange({ ...value, [id]: v })

  return (
    <div className="space-y-3">
      <p className="text-muted-foreground text-sm font-medium">{t("org.customFields.valuesTitle")}</p>
      <FieldGroup>
        {defs.map((def) => {
          const id = def.id as string
          const current = value[id]
          return (
            <Field key={id}>
              <FieldLabel>
                {def.label}
                {def.is_required ? <span className="text-destructive"> *</span> : null}
              </FieldLabel>

              {def.field_type === "boolean" ? (
                <Switch checked={!!current} onCheckedChange={(c) => set(id, c)} />
              ) : def.field_type === "select" ? (
                <Select
                  value={(current as string) ?? ""}
                  onValueChange={(v) => set(id, v)}
                  items={(def.options ?? []).map((o) => ({ value: o, label: o }))}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {(def.options ?? []).map((o) => (
                      <SelectItem key={o} value={o}>
                        {o}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              ) : def.field_type === "number" ? (
                <Input
                  type="number"
                  value={current === undefined || current === null ? "" : String(current)}
                  onChange={(e) => set(id, e.target.value === "" ? null : Number(e.target.value))}
                />
              ) : def.field_type === "date" ? (
                <Input type="date" value={(current as string) ?? ""} onChange={(e) => set(id, e.target.value || null)} />
              ) : (
                <Input value={(current as string) ?? ""} onChange={(e) => set(id, e.target.value)} />
              )}
              {def.description ? <p className="text-muted-foreground text-xs">{def.description}</p> : null}
            </Field>
          )
        })}
      </FieldGroup>
    </div>
  )
}
