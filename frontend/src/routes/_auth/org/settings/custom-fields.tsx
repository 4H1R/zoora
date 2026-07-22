import type { GithubCom4H1RZooraInternalDomainUserCustomFieldDefinition as CustomFieldDefinition } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { ArchiveIcon, PencilIcon, PlusIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetCustomFieldDefinitionsQueryKey,
  useDeleteCustomFieldDefinitionsId,
  useGetCustomFieldDefinitions,
} from "@/api/custom-fields/custom-fields"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"
import { cn } from "@/lib/utils"

import { CustomFieldSheet } from "./-custom-field-sheet"

export const Route = createFileRoute("/_auth/org/settings/custom-fields")({
  head: () => orgHead("org.customFields.title"),
  component: CustomFieldsPage,
})

// Leading accent colour per field type — the one memorable visual anchor of the ledger.
const TYPE_ACCENT: Record<string, string> = {
  text: "border-s-sky-400",
  number: "border-s-amber-400",
  date: "border-s-violet-400",
  boolean: "border-s-emerald-400",
  select: "border-s-rose-400",
}

function CustomFieldsPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const allowed = useOrgGuard("custom_fields:manage")

  const { data, isLoading } = useGetCustomFieldDefinitions()
  const defs: CustomFieldDefinition[] = (data?.status === 200 && (data.data.data as CustomFieldDefinition[])) || []

  const [editing, setEditing] = useState<CustomFieldDefinition | null>(null)
  const [sheetOpen, setSheetOpen] = useState(false)
  const [archiving, setArchiving] = useState<CustomFieldDefinition | null>(null)

  const archiveMutation = useDeleteCustomFieldDefinitionsId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.customFields.archiveSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetCustomFieldDefinitionsQueryKey() })
        setArchiving(null)
      },
      onError: () => toast.error(t("org.customFields.errors.generic")),
    },
  })

  if (!allowed) return null

  const atLimit = defs.length >= 10

  const openCreate = () => {
    setEditing(null)
    setSheetOpen(true)
  }

  return (
    <div className="space-y-8">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
        <div className="min-w-0 space-y-1">
          <h1 className="text-2xl font-bold tracking-tight">{t("org.customFields.title")}</h1>
          <p className="text-muted-foreground text-sm">{t("org.customFields.subtitle")}</p>
        </div>
        <Button onClick={openCreate} disabled={atLimit} title={atLimit ? t("org.customFields.limitReached") : undefined}>
          <PlusIcon className="me-2 size-4" />
          {t("org.customFields.add")}
        </Button>
      </div>

      {isLoading ? (
        <div className="space-y-3">
          {[0, 1, 2].map((i) => (
            <Skeleton key={i} className="h-[4.5rem] w-full rounded-xl" />
          ))}
        </div>
      ) : defs.length === 0 ? (
        <Card className="flex flex-col items-center justify-center gap-3 border-dashed py-16 text-center">
          <p className="text-muted-foreground">{t("org.customFields.empty")}</p>
          <Button variant="outline" onClick={openCreate}>
            <PlusIcon className="me-2 size-4" />
            {t("org.customFields.add")}
          </Button>
        </Card>
      ) : (
        <ul className="space-y-3">
          {defs.map((def, i) => (
            <li
              key={def.id}
              className="animate-in fade-in slide-in-from-bottom-2"
              style={{ animationDelay: `${i * 40}ms`, animationFillMode: "backwards" }}
            >
              <Card
                className={cn(
                  "flex flex-row items-center gap-4 border-s-4 p-4",
                  TYPE_ACCENT[def.field_type ?? "text"] ?? "border-s-muted"
                )}
              >
                <span className="bg-muted text-muted-foreground rounded-md px-2 py-1 font-mono text-xs tracking-wide uppercase">
                  {t(`org.customFields.types.${def.field_type}`)}
                </span>
                <div className="min-w-0 flex-1">
                  <p className="truncate font-medium">{def.label}</p>
                  {def.description ? <p className="text-muted-foreground truncate text-sm">{def.description}</p> : null}
                </div>
                <div className="flex items-center gap-2">
                  {def.is_required ? <Badge variant="secondary">{t("org.customFields.required")}</Badge> : null}
                  {def.is_unique ? <Badge variant="outline">{t("org.customFields.unique")}</Badge> : null}
                </div>
                <div className="flex items-center gap-1">
                  <Button
                    size="icon"
                    variant="ghost"
                    aria-label={t("org.customFields.editTitle")}
                    onClick={() => {
                      setEditing(def)
                      setSheetOpen(true)
                    }}
                  >
                    <PencilIcon className="size-4" />
                  </Button>
                  <Button
                    size="icon"
                    variant="ghost"
                    aria-label={t("org.customFields.archive")}
                    onClick={() => setArchiving(def)}
                  >
                    <ArchiveIcon className="size-4" />
                  </Button>
                </div>
              </Card>
            </li>
          ))}
        </ul>
      )}

      <CustomFieldSheet open={sheetOpen} onOpenChange={setSheetOpen} definition={editing} />

      <AlertDialog open={!!archiving} onOpenChange={(o) => !o && setArchiving(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{archiving?.label}</AlertDialogTitle>
            <AlertDialogDescription>{t("org.customFields.archiveConfirm")}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t("common.cancel")}</AlertDialogCancel>
            <AlertDialogAction
              disabled={archiveMutation.isPending}
              onClick={() => archiving?.id && archiveMutation.mutate({ id: archiving.id })}
            >
              {t("org.customFields.archive")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
