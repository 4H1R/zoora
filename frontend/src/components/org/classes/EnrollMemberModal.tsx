import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { CheckIcon, ChevronsUpDownIcon, UserPlusIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { useDebounce } from "use-debounce"
import { z } from "zod"

import {
  getGetClassesIdMembersQueryKey,
  usePostClassesIdMembers,
} from "@/api/classes/classes"
import { useGetUsers } from "@/api/users/users"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { UserAvatar } from "@/components/user-avatar"
import { Button } from "@/components/ui/button"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { cn } from "@/lib/utils"

const schema = z.object({
  user_id: z.string().uuid(),
})

type FormValues = z.infer<typeof schema>

interface EnrollMemberModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  classId: string
}

export function EnrollMemberModal({ open, onOpenChange, classId }: EnrollMemberModalProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { user_id: "" },
  })

  useEffect(() => {
    if (!open) form.reset({ user_id: "" })
  }, [open])

  const mutation = usePostClassesIdMembers({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.class.enrollMember.success"))
        queryClient.invalidateQueries({ queryKey: getGetClassesIdMembersQueryKey(classId) })
        onOpenChange(false)
      },
      onError: (err) => {
        const status = (err as { status?: number })?.status
        if (status === 409) {
          toast.error(t("org.class.enrollMember.errorConflict"))
        } else if (status === 403) {
          toast.error(t("org.class.enrollMember.errorForbidden"))
        } else {
          toast.error(t("org.class.enrollMember.errorGeneric"))
        }
      },
    },
  })

  const errors = form.formState.errors
  const selectedUserId = form.watch("user_id")

  const onSubmit = form.handleSubmit((values) => {
    mutation.mutate({ id: classId, data: { user_id: values.user_id } })
  })

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("org.class.enrollMember.title")}
      description={t("org.class.enrollMember.description")}
      onSubmit={onSubmit}
      isLoading={mutation.isPending}
      submitLabel={t("org.class.enrollMember.submit")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.user_id || undefined}>
          <FieldLabel>{t("org.class.enrollMember.userLabel")}</FieldLabel>
          <OrgUserPicker
            value={selectedUserId}
            onChange={(id) => form.setValue("user_id", id, { shouldValidate: true })}
          />
          <FieldError errors={[errors.user_id]} />
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}

interface OrgUserPickerProps {
  value?: string
  onChange: (userId: string) => void
}

function OrgUserPicker({ value, onChange }: OrgUserPickerProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState("")
  const [debouncedSearch] = useDebounce(search, 300)

  const { data } = useGetUsers({
    search: debouncedSearch || undefined,
    page_size: 20,
  })
  const usersData = (data?.status === 200 && data.data.data) || undefined
  const users = usersData?.items ?? []
  const selected = users.find((u) => u.id === value)

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={
          <Button
            type="button"
            variant="outline"
            role="combobox"
            className="h-11 w-full justify-between font-normal"
          />
        }
      >
        {selected ? (
          <span className="inline-flex min-w-0 items-center gap-2.5">
            <UserAvatar name={selected.name ?? ""} size="sm" />
            <span className="flex min-w-0 flex-col items-start">
              <span className="truncate text-sm font-medium">{selected.name}</span>
              {selected.username && (
                <span className="text-muted-foreground truncate font-mono text-xs">
                  @{selected.username}
                </span>
              )}
            </span>
          </span>
        ) : (
          <span className="text-muted-foreground inline-flex items-center gap-2">
            <UserPlusIcon className="size-4" />
            {t("org.class.enrollMember.userPlaceholder")}
          </span>
        )}
        <ChevronsUpDownIcon className="text-muted-foreground ms-2 size-4 shrink-0 opacity-50" />
      </PopoverTrigger>
      <PopoverContent className="w-[--radix-popover-trigger-width] min-w-72 p-0" align="start">
        <Command>
          <CommandInput
            value={search}
            onValueChange={setSearch}
            placeholder={t("org.class.enrollMember.searchPlaceholder")}
          />
          <CommandList>
            <CommandEmpty>{t("org.class.enrollMember.noResults")}</CommandEmpty>
            <CommandGroup>
              {users.map((user) => (
                <CommandItem
                  key={user.id}
                  value={`${user.name ?? ""} ${user.username ?? ""}`}
                  onSelect={() => {
                    if (user.id) {
                      onChange(user.id)
                      setOpen(false)
                    }
                  }}
                >
                  <CheckIcon
                    className={cn(
                      "me-2 size-4",
                      value === user.id ? "opacity-100" : "opacity-0"
                    )}
                  />
                  <UserAvatar name={user.name ?? ""} size="sm" />
                  <div className="ms-2 flex min-w-0 flex-col">
                    <span className="truncate text-sm">{user.name}</span>
                    {user.username && (
                      <span className="text-muted-foreground truncate font-mono text-xs">
                        @{user.username}
                      </span>
                    )}
                  </div>
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
