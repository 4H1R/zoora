import type {
  GithubCom4H1RZooraInternalDomainPermission as Permission,
  GithubCom4H1RZooraInternalDomainRole as Role,
} from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { CheckIcon, ChevronsUpDownIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  getGetRolesQueryKey,
  getGetRolesStatsQueryKey,
  useGetPermissions,
  usePostRoles,
  usePutRolesId,
} from "@/api/roles/roles"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { cn } from "@/lib/utils"

const roleSchema = z.object({
  name: z.string().min(2),
  permission_ids: z.array(z.string()).min(1),
})

type RoleFormValues = z.infer<typeof roleSchema>

interface RoleFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  role?: Role | null
  organizationId: string
}

export function RoleFormDialog({ open, onOpenChange, role, organizationId }: RoleFormDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const isEdit = !!role
  const [permOpen, setPermOpen] = useState(false)

  const { data: permData } = useGetPermissions()
  const permissions = (permData?.data?.data as Permission[] | undefined) ?? []

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors },
  } = useForm<RoleFormValues>({
    resolver: zodResolver(roleSchema),
    defaultValues: { name: "", permission_ids: [] },
  })

  useEffect(() => {
    if (open) {
      reset({
        name: role?.name ?? "",
        permission_ids: role?.permissions?.map((p) => p.id!).filter(Boolean) ?? [],
      })
    }
  }, [open, role, reset])

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetRolesQueryKey() })
    queryClient.invalidateQueries({ queryKey: getGetRolesStatsQueryKey() })
  }

  const createMutation = usePostRoles({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.roles.form.createSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const updateMutation = usePutRolesId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.roles.form.updateSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const isLoading = createMutation.isPending || updateMutation.isPending

  const onSubmit = handleSubmit((values) => {
    if (isEdit && role?.id) {
      updateMutation.mutate({ id: role.id, data: { name: values.name, permission_ids: values.permission_ids } })
    } else {
      createMutation.mutate({
        data: { organization_id: organizationId, name: values.name, permission_ids: values.permission_ids },
      })
    }
  })

  const selectedIds = watch("permission_ids")

  const togglePermission = (id: string) => {
    const current = watch("permission_ids")
    if (current.includes(id)) {
      setValue(
        "permission_ids",
        current.filter((x) => x !== id),
        { shouldValidate: true }
      )
    } else {
      setValue("permission_ids", [...current, id], { shouldValidate: true })
    }
  }

  const selectedPerms = permissions.filter((p) => p.id && selectedIds.includes(p.id))

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("org.roles.form.editTitle") : t("org.roles.form.createTitle")}
      description={isEdit ? t("org.roles.form.editDescription") : t("org.roles.form.createDescription")}
      onSubmit={onSubmit}
      isLoading={isLoading}
      submitLabel={isEdit ? t("common.save") : t("common.create")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.name || undefined}>
          <FieldLabel>{t("org.roles.form.name")}</FieldLabel>
          <Input {...register("name")} placeholder={t("org.roles.form.namePlaceholder")} />
          <FieldError errors={[errors.name]} />
        </Field>

        <Field data-invalid={!!errors.permission_ids || undefined}>
          <FieldLabel>{t("org.roles.form.permissions")}</FieldLabel>
          <Popover open={permOpen} onOpenChange={setPermOpen}>
            <PopoverTrigger
              render={
                <Button
                  variant="outline"
                  role="combobox"
                  className="h-auto min-h-9 w-full justify-between px-3 py-1.5 font-normal"
                />
              }
            >
              {selectedPerms.length > 0 ? (
                <div className="flex flex-wrap gap-1">
                  {selectedPerms.map((p) => (
                    <Badge key={p.id} variant="secondary" className="text-[11px]">
                      {p.name}
                    </Badge>
                  ))}
                </div>
              ) : (
                <span className="text-muted-foreground">{t("org.roles.form.permissionsPlaceholder")}</span>
              )}
              <ChevronsUpDownIcon className="text-muted-foreground ms-2 size-4 shrink-0 opacity-50" />
            </PopoverTrigger>
            <PopoverContent className="w-72 p-0" align="start">
              <Command>
                <CommandInput placeholder={t("org.roles.searchPlaceholder")} />
                <CommandList>
                  <CommandEmpty>{t("org.roles.noResults")}</CommandEmpty>
                  <CommandGroup>
                    {permissions.map((perm) => {
                      const isSelected = perm.id ? selectedIds.includes(perm.id) : false
                      return (
                        <CommandItem
                          key={perm.id}
                          value={perm.name}
                          onSelect={() => perm.id && togglePermission(perm.id)}
                        >
                          <CheckIcon className={cn("me-2 size-4", isSelected ? "opacity-100" : "opacity-0")} />
                          <span className="text-sm">{perm.name}</span>
                        </CommandItem>
                      )
                    })}
                  </CommandGroup>
                </CommandList>
              </Command>
            </PopoverContent>
          </Popover>
          <FieldError errors={[errors.permission_ids]} />
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}
