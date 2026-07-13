import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { CheckIcon, ChevronsUpDownIcon, UserPlusIcon, XIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { useDebounce } from "use-debounce"
import { z } from "zod"

import {
  getGetClassesIdMembersQueryKey,
  postClassesIdMembers,
} from "@/api/classes/classes"
import { useGetUsers } from "@/api/users/users"
import type { GithubCom4H1RZooraInternalDomainUser as OrgUser } from "@/api/model"
import type { ErrorType } from "@/api/mutator/custom-instance"
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

interface EnrollMemberModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  classId: string
}

export function EnrollMemberModal({ open, onOpenChange, classId }: EnrollMemberModalProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [isSubmitting, setIsSubmitting] = useState(false)

  const schema = z.object({
    user_ids: z.array(z.string().uuid()).min(1, t("org.class.enrollMember.required")),
  })

  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: { user_ids: [] },
  })

  // Chips need name/username after the search results move on, so keep the
  // full user objects alongside the id list the form validates.
  const [selectedUsers, setSelectedUsers] = useState<OrgUser[]>([])

  useEffect(() => {
    if (!open) {
      form.reset({ user_ids: [] })
      setSelectedUsers([])
    }
  }, [open])

  const errors = form.formState.errors

  const setSelection = (users: OrgUser[]) => {
    setSelectedUsers(users)
    form.setValue(
      "user_ids",
      users.map((u) => u.id!).filter(Boolean),
      { shouldValidate: form.formState.isSubmitted }
    )
  }

  const toggleUser = (user: OrgUser) => {
    if (!user.id) return
    const exists = selectedUsers.some((u) => u.id === user.id)
    setSelection(exists ? selectedUsers.filter((u) => u.id !== user.id) : [...selectedUsers, user])
  }

  const onSubmit = form.handleSubmit(async ({ user_ids }) => {
    setIsSubmitting(true)
    try {
      const results = await Promise.allSettled(
        user_ids.map((userId) => postClassesIdMembers(classId, { user_id: userId }))
      )
      const failedIds = user_ids.filter((_, i) => results[i].status === "rejected")
      const succeeded = user_ids.length - failedIds.length

      if (succeeded > 0) {
        toast.success(t("org.class.enrollMember.successCount", { count: succeeded }))
        queryClient.invalidateQueries({ queryKey: getGetClassesIdMembersQueryKey(classId) })
      }

      if (failedIds.length === 0) {
        onOpenChange(false)
        return
      }

      const statuses = results
        .filter((r): r is PromiseRejectedResult => r.status === "rejected")
        .map((r) => (r.reason as ErrorType<unknown>)?.response?.status)
      if (statuses.every((s) => s === 409)) {
        toast.error(t("org.class.enrollMember.errorConflict"))
      } else if (statuses.every((s) => s === 403)) {
        toast.error(t("org.class.enrollMember.errorForbidden"))
      } else {
        toast.error(t("org.class.enrollMember.errorPartial", { count: failedIds.length }))
      }
      // Keep the dialog open with only the failed users selected for retry.
      setSelection(selectedUsers.filter((u) => failedIds.includes(u.id ?? "")))
    } finally {
      setIsSubmitting(false)
    }
  })

  const count = selectedUsers.length

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("org.class.enrollMember.title")}
      description={t("org.class.enrollMember.description")}
      onSubmit={onSubmit}
      isLoading={isSubmitting}
      submitLabel={
        count > 0
          ? t("org.class.enrollMember.submitCount", { count })
          : t("org.class.enrollMember.submit")
      }
    >
      <FieldGroup>
        <Field data-invalid={!!errors.user_ids || undefined}>
          <FieldLabel>{t("org.class.enrollMember.userLabel")}</FieldLabel>
          <OrgUserPicker selected={selectedUsers} onToggle={toggleUser} />
          {count > 0 && (
            <div className="flex flex-wrap gap-1.5 pt-1">
              {selectedUsers.map((user) => (
                <span
                  key={user.id}
                  className="bg-muted inline-flex items-center gap-1.5 rounded-full py-0.5 ps-0.5 pe-1 text-sm"
                >
                  <UserAvatar name={user.name ?? ""} size="sm" />
                  <span className="max-w-32 truncate">{user.name}</span>
                  <button
                    type="button"
                    onClick={() => toggleUser(user)}
                    aria-label={t("org.class.enrollMember.removeUser", { name: user.name ?? "" })}
                    className="text-muted-foreground hover:bg-accent hover:text-foreground focus-visible:ring-ring rounded-full p-0.5 transition-colors focus-visible:ring-2 focus-visible:outline-none"
                  >
                    <XIcon className="size-3.5" />
                  </button>
                </span>
              ))}
            </div>
          )}
          <FieldError errors={[errors.user_ids]} />
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}

interface OrgUserPickerProps {
  selected: OrgUser[]
  onToggle: (user: OrgUser) => void
}

function OrgUserPicker({ selected, onToggle }: OrgUserPickerProps) {
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
  const selectedIds = new Set(selected.map((u) => u.id))

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
        {selected.length > 0 ? (
          <span className="inline-flex items-center gap-2">
            <UserPlusIcon className="text-muted-foreground size-4" />
            {t("org.class.enrollMember.selectedCount", { count: selected.length })}
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
                  onSelect={() => onToggle(user)}
                >
                  <CheckIcon
                    className={cn(
                      "me-2 size-4",
                      selectedIds.has(user.id) ? "opacity-100" : "opacity-0"
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
