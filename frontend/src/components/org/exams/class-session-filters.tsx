import { useTranslation } from "react-i18next"

import { useGetClasses, useGetClassesIdSessions } from "@/api/classes/classes"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { formatSessionDate } from "@/lib/session-status"

const ALL = "all"

/** Class dropdown for exam filters. Emits undefined for "all". */
export function ClassFilterSelect({
  value,
  onChange,
}: {
  value?: string
  onChange: (classId?: string) => void
}) {
  const { t } = useTranslation()
  const classesQ = useGetClasses()
  const classes = (classesQ.data?.status === 200 && classesQ.data.data.data?.items) || []

  const items = [
    { value: ALL, label: t("org.exams.filter.allClasses") },
    ...classes.map((c) => ({ value: c.id ?? "", label: c.name ?? "" })),
  ]

  return (
    <Select items={items} value={value ?? ALL} onValueChange={(v) => onChange(v && v !== ALL ? v : undefined)}>
      <SelectTrigger size="sm" className="w-40">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {items.map((item) => (
          <SelectItem key={item.value} value={item.value}>
            {item.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}

/** Session dropdown, dependent on the chosen class. Disabled until a class is picked. */
export function SessionFilterSelect({
  classId,
  value,
  onChange,
}: {
  classId?: string
  value?: string
  onChange: (sessionId?: string) => void
}) {
  const { t, i18n } = useTranslation()
  const sessionsQ = useGetClassesIdSessions(classId ?? "", { page_size: 200 }, { query: { enabled: !!classId } })
  const sessions = (sessionsQ.data?.status === 200 && sessionsQ.data.data.data?.items) || []

  const items = [
    { value: ALL, label: t("org.exams.filter.allSessions") },
    ...sessions.map((s) => ({
      value: s.id ?? "",
      label: s.name || (s.start_time ? formatSessionDate(s.start_time, i18n.language, "short") : (s.id ?? "")),
    })),
  ]

  return (
    <Select
      items={items}
      value={value ?? ALL}
      onValueChange={(v) => onChange(v && v !== ALL ? v : undefined)}
      disabled={!classId}
    >
      <SelectTrigger size="sm" className="w-40" title={!classId ? t("org.exams.filter.chooseClassFirst") : undefined}>
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {items.map((item) => (
          <SelectItem key={item.value} value={item.value}>
            {item.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
