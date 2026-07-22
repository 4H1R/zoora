import { useAccess } from "react-access-engine"

import { useGetClasses } from "@/api/classes/classes"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"

/** Classes the caller can manage: every class the API returns when they hold
 * an org-wide permission, otherwise only the classes they own (teach) —
 * GET /classes also returns classes the caller is merely enrolled in. */
export function useManagerClasses(showAll: boolean) {
  const { user } = useAccess()
  const classesQ = useGetClasses({ page_size: 100, order_by: "name", order_dir: "asc" })
  const items = (classesQ.data?.status === 200 && classesQ.data.data.data?.items) || []
  const classes = showAll ? items : items.filter((cls) => cls.user_id === user.id)
  return { classes, isLoading: classesQ.isPending }
}

/** Mandatory class dropdown for manager views — no "all classes" option. */
export function ManagerClassPicker({
  classes,
  value,
  onChange,
}: {
  classes: { id?: string; name?: string }[]
  value?: string
  onChange: (classId: string) => void
}) {
  const items = classes.map((cls) => ({ value: cls.id ?? "", label: cls.name ?? "" }))

  return (
    <Select items={items} value={value ?? ""} onValueChange={(v) => v && onChange(v)}>
      <SelectTrigger size="sm" className="w-56">
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
