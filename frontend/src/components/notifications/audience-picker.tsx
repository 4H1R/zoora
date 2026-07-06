import type { GithubCom4H1RZooraInternalDomainNotificationAudienceDTO as AudienceDTO } from "@/api/model"

import { CheckIcon, ChevronsUpDownIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { useDebounce } from "use-debounce"

import { useGetAdminClasses } from "@/api/admin-classes/admin-classes"
import { useGetAdminRoles } from "@/api/admin-roles/admin-roles"
import { useGetClasses } from "@/api/classes/classes"
import { GithubCom4H1RZooraInternalDomainNotificationAudienceDTOType as AudienceType } from "@/api/model"
import { useGetRoles } from "@/api/roles/roles"
import { OrganizationSelect } from "@/components/form/organization-select"
import { UserMultiSelect } from "@/components/notifications/user-multi-select"
import { Button } from "@/components/ui/button"
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command"
import { Label } from "@/components/ui/label"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group"
import { cn } from "@/lib/utils"

export type AudienceMode = "admin" | "manager" | "teacher"

// Which audience types each caller flavor may target, in display order.
const MODE_TYPES: Record<AudienceMode, (typeof AudienceType)[keyof typeof AudienceType][]> = {
  admin: [AudienceType.all, AudienceType.org, AudienceType.class, AudienceType.role, AudienceType.users],
  manager: [AudienceType.org, AudienceType.class, AudienceType.role, AudienceType.users],
  teacher: [AudienceType.class, AudienceType.users],
}

const TYPE_LABEL: Record<string, string> = {
  [AudienceType.all]: "notifications.send.audienceAll",
  [AudienceType.org]: "notifications.send.audienceOrg",
  [AudienceType.class]: "notifications.send.audienceClass",
  [AudienceType.role]: "notifications.send.audienceRole",
  [AudienceType.users]: "notifications.send.audienceUsers",
}

interface AudiencePickerProps {
  mode: AudienceMode
  value: AudienceDTO
  onChange: (value: AudienceDTO) => void
}

/** Adaptive audience selector. The available targets and the data sources adjust
 * to the caller: platform admins pick across orgs, org managers stay inside
 * their org, teachers reach only their classes and those classes' members. */
export function AudiencePicker({ mode, value, onChange }: AudiencePickerProps) {
  const { t } = useTranslation()
  const scope = mode === "admin" ? "admin" : "org"
  const types = MODE_TYPES[mode]

  const setType = (type: string) => {
    // Reset the type-specific fields when switching targets.
    onChange({ type: type as AudienceDTO["type"] })
  }

  return (
    <div className="flex flex-col gap-3">
      <ToggleGroup
        value={[value.type]}
        onValueChange={(v: string[]) => {
          const next = v.find((x) => x !== value.type)
          if (next) setType(next)
        }}
        spacing={1.5}
        className="flex flex-wrap"
      >
        {types.map((type) => (
          <ToggleGroupItem key={type} value={type} className="rounded-full border px-3 text-xs">
            {t(TYPE_LABEL[type])}
          </ToggleGroupItem>
        ))}
      </ToggleGroup>

      {value.type === AudienceType.org && mode === "admin" && (
        <FieldWrap label={t("notifications.send.selectOrg")}>
          <OrganizationSelect
            value={value.org_id}
            onChange={(org_id) => onChange({ ...value, org_id })}
            placeholder={t("notifications.send.selectOrg")}
          />
        </FieldWrap>
      )}

      {value.type === AudienceType.class && (
        <>
          {mode === "admin" && (
            <FieldWrap label={t("notifications.send.selectOrg")}>
              <OrganizationSelect
                value={value.org_id}
                onChange={(org_id) => onChange({ ...value, org_id, class_id: undefined })}
                placeholder={t("notifications.send.allOrgs")}
              />
            </FieldWrap>
          )}
          <FieldWrap label={t("notifications.send.selectClass")}>
            <ClassSelect
              scope={scope}
              organizationId={value.org_id}
              value={value.class_id}
              onChange={(class_id) => onChange({ ...value, class_id })}
            />
          </FieldWrap>
        </>
      )}

      {value.type === AudienceType.role && (
        <>
          {mode === "admin" && (
            <FieldWrap label={t("notifications.send.selectOrg")}>
              <OrganizationSelect
                value={value.org_id}
                onChange={(org_id) => onChange({ ...value, org_id, role_id: undefined })}
                placeholder={t("notifications.send.allOrgs")}
              />
            </FieldWrap>
          )}
          <FieldWrap label={t("notifications.send.selectRole")}>
            <RoleSelect
              scope={scope}
              organizationId={value.org_id}
              value={value.role_id}
              onChange={(role_id) => onChange({ ...value, role_id })}
            />
          </FieldWrap>
        </>
      )}

      {value.type === AudienceType.users && (
        <>
          {mode === "teacher" && (
            <FieldWrap label={t("notifications.send.selectClass")}>
              <ClassSelect
                scope="org"
                value={value.class_id}
                onChange={(class_id) => onChange({ ...value, class_id, user_ids: [] })}
              />
            </FieldWrap>
          )}
          <FieldWrap label={t("notifications.send.selectUsers")}>
            <UserMultiSelect
              scope={scope}
              organizationId={mode === "admin" ? value.org_id : undefined}
              classId={mode === "teacher" ? value.class_id : undefined}
              value={value.user_ids ?? []}
              onChange={(user_ids) => onChange({ ...value, user_ids })}
            />
          </FieldWrap>
        </>
      )}
    </div>
  )
}

function FieldWrap({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-1.5">
      <Label className="text-muted-foreground text-xs">{label}</Label>
      {children}
    </div>
  )
}

interface EntitySelectProps {
  scope: "admin" | "org"
  organizationId?: string
  value?: string
  onChange: (id: string) => void
}

function ClassSelect({ scope, organizationId, value, onChange }: EntitySelectProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState("")
  const [debouncedSearch] = useDebounce(search, 300)
  const isAdmin = scope === "admin"

  const adminQuery = useGetAdminClasses({ search: debouncedSearch || undefined }, { query: { enabled: isAdmin } })
  const orgQuery = useGetClasses({ search: debouncedSearch || undefined }, { query: { enabled: !isAdmin } })
  const data = isAdmin ? adminQuery.data : orgQuery.data
  const pageData = (data?.status === 200 && data.data.data) || undefined
  let items = pageData?.items ?? []
  // Admin classes carry organization_id; narrow client-side when an org filter
  // is set (the admin classes endpoint has no org param).
  if (isAdmin && organizationId) items = items.filter((c) => c.organization_id === organizationId)

  const selected = items.find((c) => c.id === value)

  return (
    <Combobox
      open={open}
      setOpen={setOpen}
      search={search}
      setSearch={setSearch}
      label={selected?.name}
      placeholder={t("notifications.send.selectClass")}
    >
      {items.map((c) => (
        <CommandItem
          key={c.id}
          value={c.id ?? ""}
          onSelect={() => {
            if (c.id) {
              onChange(c.id)
              setOpen(false)
            }
          }}
        >
          <CheckIcon className={cn("me-2 size-4 shrink-0", value === c.id ? "opacity-100" : "opacity-0")} />
          <span className="truncate text-sm">{c.name}</span>
        </CommandItem>
      ))}
    </Combobox>
  )
}

function RoleSelect({ scope, organizationId, value, onChange }: EntitySelectProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState("")
  const [debouncedSearch] = useDebounce(search, 300)
  const isAdmin = scope === "admin"

  const adminQuery = useGetAdminRoles(
    { search: debouncedSearch || undefined, organization_id: organizationId || undefined },
    { query: { enabled: isAdmin } }
  )
  const orgQuery = useGetRoles({ query: { enabled: !isAdmin } })

  let items = isAdmin
    ? (adminQuery.data?.status === 200 && adminQuery.data.data.data?.items) || []
    : (orgQuery.data?.status === 200 && orgQuery.data.data.data) || []
  if (!isAdmin && debouncedSearch) {
    const q = debouncedSearch.toLowerCase()
    items = items.filter((r) => (r.name ?? "").toLowerCase().includes(q))
  }

  const selected = items.find((r) => r.id === value)

  return (
    <Combobox
      open={open}
      setOpen={setOpen}
      search={search}
      setSearch={setSearch}
      label={selected?.name}
      placeholder={t("notifications.send.selectRole")}
    >
      {items.map((r) => (
        <CommandItem
          key={r.id}
          value={r.id ?? ""}
          onSelect={() => {
            if (r.id) {
              onChange(r.id)
              setOpen(false)
            }
          }}
        >
          <CheckIcon className={cn("me-2 size-4 shrink-0", value === r.id ? "opacity-100" : "opacity-0")} />
          <span className="truncate text-sm">{r.name}</span>
        </CommandItem>
      ))}
    </Combobox>
  )
}

interface ComboboxProps {
  open: boolean
  setOpen: (o: boolean) => void
  search: string
  setSearch: (s: string) => void
  label?: string
  placeholder: string
  children: React.ReactNode
}

function Combobox({ open, setOpen, search, setSearch, label, placeholder, children }: ComboboxProps) {
  const { t } = useTranslation()
  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={<Button variant="outline" role="combobox" className="w-full justify-between font-normal" />}
      >
        <span className={cn(!label && "text-muted-foreground", "truncate")}>{label ?? placeholder}</span>
        <ChevronsUpDownIcon className="text-muted-foreground ms-2 size-4 shrink-0 opacity-50" />
      </PopoverTrigger>
      <PopoverContent className="w-72 p-0" align="start">
        <Command shouldFilter={false}>
          <CommandInput value={search} onValueChange={setSearch} placeholder={t("common.search")} />
          <CommandList>
            <CommandEmpty>{t("common.noResults")}</CommandEmpty>
            <CommandGroup>{children}</CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
