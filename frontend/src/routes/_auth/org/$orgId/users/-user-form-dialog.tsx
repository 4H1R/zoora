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

import { useGetRoles } from "@/api/roles/roles"
import { getGetUsersQueryKey, usePostUsers, usePutUsersId } from "@/api/users/users"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Button } from "@/components/ui/button"
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { cn } from "@/lib/utils"

const baseSchema = z.object({
  name: z.string().min(2),
  username: z.string().min(3),
  password: z.string(),
  role_id: z.string().optional(),
})

const createSchema = baseSchema.extend({ password: z.string().min(6) })
const editSchema = baseSchema.extend({ password: z.string().optional().default("") })

type UserFormValues = z.infer<typeof baseSchema>

interface UserFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  user?: User | null
  organizationId: string
}

export function UserFormDialog({ open, onOpenChange, user, organizationId }: UserFormDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
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
    defaultValues: { name: "", username: "", password: "", role_id: "" },
  })

  useEffect(() => {
    if (open) {
      reset({
        name: user?.name ?? "",
        username: user?.username ?? "",
        password: "",
        role_id: user?.role_id ?? "",
      })
    }
  }, [open, user, reset])

  const { data: rolesData } = useGetRoles({ organization_id: organizationId })
  const allRoles = (rolesData?.data?.data as Role[] | undefined) ?? []
  const roles = allRoles.filter((r) => !(r.is_preset && r.name === "Staff"))

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetUsersQueryKey() })
  }

  const createMutation = usePostUsers({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.users.form.createSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const updateMutation = usePutUsersId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.users.form.updateSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const isLoading = createMutation.isPending || updateMutation.isPending

  const onSubmit = handleSubmit((values) => {
    if (isEdit && user?.id) {
      updateMutation.mutate({
        id: user.id,
        data: {
          name: values.name,
          username: values.username,
          role_id: values.role_id || undefined,
        },
      })
    } else {
      createMutation.mutate({
        data: {
          organization_id: organizationId,
          name: values.name,
          username: values.username,
          password: values.password!,
          role_id: values.role_id || undefined,
        },
      })
    }
  })

  const selectedRoleId = watch("role_id")

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("org.users.form.editTitle") : t("org.users.form.createTitle")}
      description={isEdit ? t("org.users.form.editDescription") : t("org.users.form.createDescription")}
      onSubmit={onSubmit}
      isLoading={isLoading}
      submitLabel={isEdit ? t("common.save") : t("common.create")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.name || undefined}>
          <FieldLabel>{t("org.users.form.name")}</FieldLabel>
          <Input {...register("name")} placeholder={t("org.users.form.namePlaceholder")} />
          <FieldError errors={[errors.name]} />
        </Field>

        <Field data-invalid={!!errors.username || undefined}>
          <FieldLabel>{t("org.users.form.username")}</FieldLabel>
          <Input {...register("username")} placeholder={t("org.users.form.usernamePlaceholder")} />
          <FieldError errors={[errors.username]} />
        </Field>

        {!isEdit && (
          <Field data-invalid={!!errors.password || undefined}>
            <FieldLabel>{t("org.users.form.password")}</FieldLabel>
            <Input {...register("password")} type="password" placeholder={t("org.users.form.passwordPlaceholder")} />
            <FieldError errors={[errors.password]} />
          </Field>
        )}

        <Field>
          <FieldLabel>{t("org.users.form.role")}</FieldLabel>
          <RoleSelect
            roles={roles}
            value={selectedRoleId}
            onChange={(id) => setValue("role_id", id, { shouldValidate: true })}
            placeholder={t("org.users.form.rolePlaceholder")}
          />
        </Field>
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
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState("")

  const selected = roles.find((r) => r.id === value)

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={<Button variant="outline" role="combobox" className="w-full justify-between font-normal" />}
      >
        {selected ? (
          <span className="truncate">{selected.name}</span>
        ) : (
          <span className="text-muted-foreground">{placeholder ?? t("org.users.form.rolePlaceholder")}</span>
        )}
        <ChevronsUpDownIcon className="text-muted-foreground ms-2 size-4 shrink-0 opacity-50" />
      </PopoverTrigger>
      <PopoverContent className="w-72 p-0" align="start">
        <Command>
          <CommandInput value={search} onValueChange={setSearch} placeholder={t("org.users.form.rolePlaceholder")} />
          <CommandList>
            <CommandEmpty>{t("common.noResults")}</CommandEmpty>
            <CommandGroup>
              <CommandItem
                value={t("org.users.form.noRole")}
                onSelect={() => {
                  onChange("")
                  setOpen(false)
                  setSearch("")
                }}
              >
                <CheckIcon className={cn("me-2 size-4", !value ? "opacity-100" : "opacity-0")} />
                <span className="text-muted-foreground text-sm">{t("org.users.form.noRole")}</span>
              </CommandItem>
              {roles.map((role) => (
                <CommandItem
                  key={role.id}
                  value={role.name}
                  onSelect={() => {
                    if (role.id) {
                      onChange(role.id)
                      setOpen(false)
                      setSearch("")
                    }
                  }}
                >
                  <CheckIcon className={cn("me-2 size-4", value === role.id ? "opacity-100" : "opacity-0")} />
                  <span className="text-sm">{role.name}</span>
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
