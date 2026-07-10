import type {
  GithubCom4H1RZooraInternalDomainRole as Role,
  GithubCom4H1RZooraInternalDomainUser as User,
} from "@/api/model"
import type { Resolver } from "react-hook-form"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { CheckIcon, ChevronsUpDownIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import { getGetAdminUsersQueryKey, usePostAdminUsers, usePutAdminUsersId } from "@/api/admin-users/admin-users"
import { useGetAdminRoles } from "@/api/admin-roles/admin-roles"
import { OrganizationSelect } from "@/components/form/organization-select"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command"
import { Field, FieldContent, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { useRoleName } from "@/lib/permissions"
import { cn } from "@/lib/utils"
import { useAdminStore } from "@/stores/admin"

const baseSchema = z.object({
  organization_id: z.string().uuid().optional(),
  name: z.string().min(2),
  username: z
    .string()
    .min(3)
    .max(30)
    .regex(/^[a-z0-9_.]+$/),
  password: z.string(),
  role_id: z.string().optional(),
  is_admin: z.boolean().default(false),
})

const createSchema = baseSchema.extend({ password: z.string().min(6) })
const editSchema = baseSchema.extend({ password: z.string().optional().default("") })

type UserFormValues = z.infer<typeof baseSchema>

interface UserFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  user?: User | null
}

export function UserFormDialog({ open, onOpenChange, user }: UserFormDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const activeOrganizationId = useAdminStore((s) => s.activeOrganizationId)
  const isEdit = !!user

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors },
  } = useForm<UserFormValues>({
    resolver: zodResolver(isEdit ? editSchema : createSchema) as Resolver<UserFormValues>,
    defaultValues: { organization_id: "", name: "", username: "", password: "", role_id: "", is_admin: false },
  })

  useEffect(() => {
    if (open) {
      reset({
        organization_id: user?.organization_id ?? activeOrganizationId ?? "",
        name: user?.name ?? "",
        username: user?.username ?? "",
        password: "",
        role_id: user?.role_id ?? "",
        is_admin: user?.is_admin ?? false,
      })
    }
  }, [open, user, reset, activeOrganizationId])

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetAdminUsersQueryKey() })
  }

  const createMutation = usePostAdminUsers({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.users.form.createSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const updateMutation = usePutAdminUsersId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.users.form.updateSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const isLoading = createMutation.isPending || updateMutation.isPending

  const onSubmit = handleSubmit((values) => {
    // Org-scoped users can never be platform admins; drop a stale checkbox
    // value left over from before an organization was picked.
    const orgId = isEdit ? user?.organization_id : values.organization_id
    const isAdminValue = orgId ? false : values.is_admin
    if (isEdit && user?.id) {
      updateMutation.mutate({
        id: user.id,
        data: {
          name: values.name,
          username: values.username,
          password: values.password || undefined,
          role_id: values.role_id || undefined,
          is_admin: isAdminValue,
        },
      })
    } else {
      createMutation.mutate({
        data: {
          organization_id: values.organization_id || undefined,
          name: values.name,
          username: values.username,
          password: values.password!,
          role_id: values.role_id || undefined,
          is_admin: isAdminValue,
        },
      })
    }
  })

  const isAdmin = watch("is_admin")
  const selectedOrgId = watch("organization_id")
  const selectedRoleId = watch("role_id")
  // Platform admins never belong to an organization — the backend rejects
  // is_admin on org-scoped users.
  const hasOrg = isEdit ? !!user?.organization_id : !!selectedOrgId

  const orgIdForRoles = isEdit ? user?.organization_id : selectedOrgId
  const { data: rolesData } = useGetAdminRoles(orgIdForRoles ? { organization_id: orgIdForRoles } : undefined)
  const rolesPage = rolesData?.data?.data as { items?: Role[] } | undefined
  const roles = rolesPage?.items ?? []

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("admin.users.form.editTitle") : t("admin.users.form.createTitle")}
      description={isEdit ? t("admin.users.form.editDescription") : t("admin.users.form.createDescription")}
      onSubmit={onSubmit}
      isLoading={isLoading}
      submitLabel={isEdit ? t("common.save") : t("common.create")}
    >
      <FieldGroup>
        {!isEdit && (
          <Field data-invalid={!!errors.organization_id || undefined}>
            <FieldLabel>{t("admin.users.form.organization")}</FieldLabel>
            <OrganizationSelect
              value={selectedOrgId || undefined}
              onChange={(id) => setValue("organization_id", id, { shouldValidate: true })}
              placeholder={t("admin.users.form.organizationPlaceholder")}
            />
            <FieldError errors={[errors.organization_id]} />
          </Field>
        )}

        <Field data-invalid={!!errors.name || undefined}>
          <FieldLabel>{t("admin.users.form.name")}</FieldLabel>
          <Input {...register("name")} placeholder={t("admin.users.form.namePlaceholder")} />
          <FieldError errors={[errors.name]} />
        </Field>

        <Field data-invalid={!!errors.username || undefined}>
          <FieldLabel>{t("admin.users.form.username")}</FieldLabel>
          <Input {...register("username")} placeholder={t("admin.users.form.usernamePlaceholder")} />
          <FieldError errors={[errors.username]} />
        </Field>

        <Field data-invalid={!!errors.password || undefined}>
          <FieldLabel>
            {t("admin.users.form.password")}
            {isEdit && (
              <span className="text-muted-foreground ms-1 text-xs font-normal">
                ({t("admin.users.form.passwordOptional")})
              </span>
            )}
          </FieldLabel>
          <Input
            {...register("password")}
            type="password"
            placeholder={
              isEdit ? t("admin.users.form.passwordOptionalPlaceholder") : t("admin.users.form.passwordPlaceholder")
            }
          />
          <FieldError errors={[errors.password]} />
        </Field>

        <Field>
          <FieldLabel>{t("admin.users.form.role")}</FieldLabel>
          <RoleSelect
            roles={roles}
            value={selectedRoleId}
            onChange={(id) => setValue("role_id", id, { shouldValidate: true })}
            placeholder={t("admin.users.form.rolePlaceholder")}
          />
        </Field>

        {!hasOrg && (
          <Field orientation="horizontal">
            <Checkbox
              checked={isAdmin}
              onCheckedChange={(checked) => setValue("is_admin", !!checked, { shouldValidate: true })}
            />
            <FieldContent>
              <FieldLabel>{t("admin.users.form.isAdmin")}</FieldLabel>
            </FieldContent>
          </Field>
        )}
      </FieldGroup>
    </ResourceFormDialog>
  )
}

interface RoleSelectProps {
  roles: Role[]
  value?: string
  onChange: (roleId: string) => void
  placeholder?: string
}

function RoleSelect({ roles, value, onChange, placeholder }: RoleSelectProps) {
  const { t } = useTranslation()
  const roleName = useRoleName()
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState("")

  const selected = roles.find((r) => r.id === value)

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={<Button variant="outline" role="combobox" className="w-full justify-between font-normal" />}
      >
        {selected ? (
          <span className="truncate">{selected.name ? roleName(selected.name) : ""}</span>
        ) : (
          <span className="text-muted-foreground">{placeholder ?? t("admin.users.form.rolePlaceholder")}</span>
        )}
        <ChevronsUpDownIcon className="text-muted-foreground ms-2 size-4 shrink-0 opacity-50" />
      </PopoverTrigger>
      <PopoverContent className="w-72 p-0" align="start">
        <Command>
          <CommandInput value={search} onValueChange={setSearch} placeholder={t("admin.users.form.rolePlaceholder")} />
          <CommandList>
            <CommandEmpty>{t("common.noResults")}</CommandEmpty>
            <CommandGroup>
              {roles.map((r) => (
                <CommandItem
                  key={r.id}
                  value={r.name ? roleName(r.name) : r.name}
                  onSelect={() => {
                    if (r.id) {
                      onChange(r.id)
                      setOpen(false)
                      setSearch("")
                    }
                  }}
                >
                  <CheckIcon className={cn("me-2 size-4", value === r.id ? "opacity-100" : "opacity-0")} />
                  <span className="text-sm">{r.name ? roleName(r.name) : ""}</span>
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
