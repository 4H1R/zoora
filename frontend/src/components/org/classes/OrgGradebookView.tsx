import type {
  GithubCom4H1RZooraInternalDomainClass as Class,
  GithubCom4H1RZooraInternalDomainGradebookColumn as GradebookColumn,
  GithubCom4H1RZooraInternalDomainGradebookMatrixRow as GradebookRow,
} from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import {
  EllipsisVerticalIcon,
  LockIcon,
  PencilIcon,
  PlusIcon,
  Trash2Icon,
  TrophyIcon,
} from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { useAccess } from "react-access-engine"
import { toast } from "sonner"

import type { ErrorType } from "@/api/mutator/custom-instance"
import {
  getGetClassesIdGradebookQueryKey,
  useDeleteClassesIdGradebookColumnsColumnId,
  useGetClassesIdGradebook,
} from "@/api/gradebook/gradebook"
import { GithubCom4H1RZooraInternalDomainGradebookColumnType as ColumnType } from "@/api/model"
import { GradebookCellDialog } from "@/components/admin/gradebook/GradebookCellDialog"
import { GradebookColumnDialog } from "@/components/admin/gradebook/GradebookColumnDialog"
import { Eyebrow } from "@/components/eyebrow"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Skeleton } from "@/components/ui/skeleton"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { getEntityColor, getInitials } from "@/lib/data-table"
import { cn } from "@/lib/utils"

const AUTO_TYPES = new Set<string>([
  ColumnType.GradebookColumnAutoAttendance,
  ColumnType.GradebookColumnAutoPractice,
  ColumnType.GradebookColumnAutoQuiz,
])

const TYPE_BADGE: Record<string, "default" | "secondary" | "outline"> = {
  auto_attendance: "secondary",
  auto_practice: "secondary",
  auto_quiz: "secondary",
  manual_grade: "default",
  manual_attendance: "outline",
  manual_text: "outline",
}

interface CellTarget {
  column: GradebookColumn
  student: { id: string; name: string }
  value: string
}

interface OrgGradebookViewProps {
  classId: string
  cls?: Class
}

export function OrgGradebookView({ classId, cls }: OrgGradebookViewProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { can, user } = useAccess()

  const [columnDialogOpen, setColumnDialogOpen] = useState(false)
  const [editingColumn, setEditingColumn] = useState<GradebookColumn | null>(null)
  const [deleteColumn, setDeleteColumn] = useState<GradebookColumn | null>(null)
  const [cellTarget, setCellTarget] = useState<CellTarget | null>(null)

  const { data, isLoading, isError, error } = useGetClassesIdGradebook(classId)
  const matrix = (data?.status === 200 && data.data.data) || undefined
  const columns = matrix?.columns ?? []
  const rows = matrix?.rows ?? []

  const isOwner = !!cls?.user_id && cls.user_id === user.id
  const canManage = can("gradebook:update_any") || isOwner
  const canDelete = can("gradebook:delete_any") || isOwner

  const errStatus = (error as ErrorType<unknown> | undefined)?.response?.status
  const forbidden = isError && errStatus === 403

  const deleteMutation = useDeleteClassesIdGradebookColumnsColumnId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.class.gradebook.deleteSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetClassesIdGradebookQueryKey(classId) })
        setDeleteColumn(null)
      },
    },
  })

  const openCreate = () => {
    setEditingColumn(null)
    setColumnDialogOpen(true)
  }

  const handleColumnDialogChange = (open: boolean) => {
    setColumnDialogOpen(open)
    if (!open) setEditingColumn(null)
  }

  const handleCellClick = (column: GradebookColumn, row: GradebookRow) => {
    if (!canManage) return
    if (!column.id || !row.student_id) return
    if (AUTO_TYPES.has(column.type ?? "")) return
    setCellTarget({
      column,
      student: { id: row.student_id, name: row.student?.name ?? row.student_id },
      value: row.cells?.[column.id] ?? "",
    })
  }

  if (forbidden) {
    return (
      <div className="bg-card ring-foreground/10 flex flex-col items-center gap-3 rounded-2xl px-6 py-16 text-center ring-1">
        <LockIcon className="text-muted-foreground size-8" />
        <h3 className="text-foreground text-lg font-semibold tracking-tight">
          {t("org.class.gradebook.noAccessTitle")}
        </h3>
        <p className="text-muted-foreground max-w-md text-sm leading-relaxed">
          {t("org.class.gradebook.noAccessHint")}
        </p>
      </div>
    )
  }

  return (
    <section className="flex flex-col gap-5">
      <div className="flex items-end justify-between gap-4">
        <div className="flex flex-col gap-1.5">
          <Eyebrow>{t("org.class.gradebook.eyebrow")}</Eyebrow>
          <h2 className="text-2xl font-semibold tracking-tight">{t("org.class.gradebook.title")}</h2>
        </div>
        {canManage ? (
          <Button onClick={openCreate}>
            <PlusIcon className="size-4" />
            {t("org.class.gradebook.newColumn")}
          </Button>
        ) : null}
      </div>

      <section className="bg-card ring-foreground/10 grid grid-cols-2 divide-x divide-dashed overflow-hidden rounded-2xl ring-1 rtl:divide-x-reverse">
        <div className="flex flex-col gap-2 px-5 py-5">
          <Eyebrow>{t("org.class.gradebook.stats.columns")}</Eyebrow>
          <span className="text-3xl font-semibold tracking-tight tabular-nums">{columns.length}</span>
        </div>
        <div className="flex flex-col gap-2 px-5 py-5">
          <Eyebrow>{t("org.class.gradebook.stats.students")}</Eyebrow>
          <span className="text-3xl font-semibold tracking-tight tabular-nums">{rows.length}</span>
        </div>
      </section>

      <div className="bg-card ring-foreground/10 overflow-hidden rounded-2xl ring-1">
        <div className="overflow-x-auto">
          {isLoading ? (
            <div className="flex flex-col gap-3 p-6">
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
            </div>
          ) : columns.length === 0 && rows.length === 0 ? (
            <div className="flex flex-col items-center justify-center gap-3 py-16 text-center">
              <TrophyIcon className="text-muted-foreground size-8 opacity-40" />
              <p className="text-muted-foreground text-sm">{t("org.class.gradebook.noResults")}</p>
              <p className="text-muted-foreground text-xs">{t("org.class.gradebook.noResultsHint")}</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="bg-card sticky start-0 z-10 min-w-56">
                    {t("org.class.gradebook.student")}
                  </TableHead>
                  {columns.map((col) => (
                    <TableHead key={col.id} className="min-w-40">
                      <div className="flex items-center justify-between gap-2">
                        <div className="flex min-w-0 flex-col gap-1">
                          <span className="truncate text-sm font-medium">{col.title}</span>
                          <Badge variant={TYPE_BADGE[col.type ?? ""] ?? "outline"} className="w-fit text-[10px]">
                            {t(`org.class.gradebook.types.${col.type}`)}
                          </Badge>
                        </div>
                        {canManage || canDelete ? (
                          <DropdownMenu>
                            <DropdownMenuTrigger
                              render={
                                <Button variant="ghost" size="icon-xs">
                                  <EllipsisVerticalIcon />
                                </Button>
                              }
                            />
                            <DropdownMenuContent align="end" className="min-w-40">
                              {canManage ? (
                                <DropdownMenuGroup>
                                  <DropdownMenuItem
                                    onClick={() => {
                                      setEditingColumn(col)
                                      setColumnDialogOpen(true)
                                    }}
                                  >
                                    <PencilIcon data-icon="inline-start" />
                                    {t("common.edit")}
                                  </DropdownMenuItem>
                                </DropdownMenuGroup>
                              ) : null}
                              {canDelete ? (
                                <>
                                  {canManage ? <DropdownMenuSeparator /> : null}
                                  <DropdownMenuGroup>
                                    <DropdownMenuItem
                                      variant="destructive"
                                      onClick={() => setDeleteColumn(col)}
                                    >
                                      <Trash2Icon data-icon="inline-start" />
                                      {t("common.delete")}
                                    </DropdownMenuItem>
                                  </DropdownMenuGroup>
                                </>
                              ) : null}
                            </DropdownMenuContent>
                          </DropdownMenu>
                        ) : null}
                      </div>
                    </TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {rows.map((row) => (
                  <TableRow key={row.student_id}>
                    <TableCell className="bg-card sticky start-0 z-10">
                      <div className="flex items-center gap-3">
                        <div
                          className={cn(
                            "flex size-8 shrink-0 items-center justify-center rounded-lg text-xs font-semibold text-white",
                            getEntityColor(row.student?.name ?? row.student_id ?? "")
                          )}
                        >
                          {getInitials(row.student?.name ?? row.student_id ?? "")}
                        </div>
                        <div className="min-w-0">
                          <div className="truncate text-sm font-medium">{row.student?.name ?? "—"}</div>
                          {row.student?.username && (
                            <div className="text-muted-foreground truncate text-xs">
                              {row.student.username}
                            </div>
                          )}
                        </div>
                      </div>
                    </TableCell>
                    {columns.map((col) => {
                      const value = col.id ? row.cells?.[col.id] : undefined
                      const isAuto = AUTO_TYPES.has(col.type ?? "")
                      const editable = canManage && !isAuto
                      return (
                        <TableCell
                          key={col.id}
                          className={cn(
                            "text-sm",
                            editable && "hover:bg-muted/50 cursor-pointer transition-colors"
                          )}
                          onClick={() => handleCellClick(col, row)}
                        >
                          {value ? (
                            isAuto ? (
                              <Badge variant="outline" className="font-normal">
                                {value}
                              </Badge>
                            ) : (
                              <span className="font-medium">{value}</span>
                            )
                          ) : (
                            <span className="text-muted-foreground">—</span>
                          )}
                        </TableCell>
                      )
                    })}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </div>
      </div>

      {canManage || canDelete ? (
        <>
          <GradebookColumnDialog
            open={columnDialogOpen}
            onOpenChange={handleColumnDialogChange}
            classId={classId}
            column={editingColumn}
          />
          <GradebookCellDialog
            open={!!cellTarget}
            onOpenChange={(open) => {
              if (!open) setCellTarget(null)
            }}
            classId={classId}
            columnId={cellTarget?.column.id}
            studentId={cellTarget?.student.id}
            studentName={cellTarget?.student.name}
            columnTitle={cellTarget?.column.title}
            initialValue={cellTarget?.value}
          />
          <DeleteConfirmDialog
            open={!!deleteColumn}
            onOpenChange={(open) => {
              if (!open) setDeleteColumn(null)
            }}
            resourceName={deleteColumn?.title ?? ""}
            onConfirm={() => {
              if (deleteColumn?.id) deleteMutation.mutate({ id: classId, columnId: deleteColumn.id })
            }}
            isLoading={deleteMutation.isPending}
          />
        </>
      ) : null}
    </section>
  )
}
