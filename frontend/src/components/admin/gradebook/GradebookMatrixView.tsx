import type {
  GithubCom4H1RZooraInternalDomainGradebookColumn as GradebookColumn,
  GithubCom4H1RZooraInternalDomainGradebookMatrixRow as GradebookRow,
} from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { EllipsisVerticalIcon, PencilIcon, PlusIcon, Trash2Icon, TrophyIcon, UsersIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetClassesIdGradebookQueryKey,
  useDeleteClassesIdGradebookColumnsColumnId,
  useGetClassesIdGradebook,
} from "@/api/gradebook/gradebook"
import { GithubCom4H1RZooraInternalDomainGradebookColumnType as ColumnType } from "@/api/model"
import { GradebookCellDialog } from "@/components/admin/gradebook/GradebookCellDialog"
import { GradebookColumnDialog } from "@/components/admin/gradebook/GradebookColumnDialog"
import { StatCards } from "@/components/data-table/stat-cards"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
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

interface GradebookMatrixViewProps {
  classId: string
}

export function GradebookMatrixView({ classId }: GradebookMatrixViewProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const [columnDialogOpen, setColumnDialogOpen] = useState(false)
  const [editingColumn, setEditingColumn] = useState<GradebookColumn | null>(null)
  const [deleteColumn, setDeleteColumn] = useState<GradebookColumn | null>(null)
  const [cellTarget, setCellTarget] = useState<CellTarget | null>(null)

  const { data, isLoading } = useGetClassesIdGradebook(classId)
  const matrix = (data?.status === 200 && data.data.data) || undefined
  const columns = matrix?.columns ?? []
  const rows = matrix?.rows ?? []

  const deleteMutation = useDeleteClassesIdGradebookColumnsColumnId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.gradebook.form.deleteSuccess"))
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
    if (!column.id || !row.student_id) return
    if (AUTO_TYPES.has(column.type ?? "")) return
    setCellTarget({
      column,
      student: {
        id: row.student_id,
        name: row.student?.name ?? row.student_id,
      },
      value: row.cells?.[column.id] ?? "",
    })
  }

  const statCards = [
    {
      icon: <TrophyIcon />,
      label: t("admin.gradebook.stats.columns"),
      value: columns.length,
      loading: isLoading,
    },
    {
      icon: <UsersIcon />,
      label: t("admin.gradebook.stats.students"),
      value: rows.length,
      loading: isLoading,
    },
  ]

  return (
    <>
      <div className="flex items-center justify-end">
        <Button size="sm" onClick={openCreate}>
          <PlusIcon data-icon="inline-start" />
          {t("admin.gradebook.newColumn")}
        </Button>
      </div>

      <StatCards stats={statCards} />

      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          {columns.length === 0 && rows.length === 0 && !isLoading ? (
            <div className="flex flex-col items-center justify-center gap-3 py-16 text-center">
              <TrophyIcon className="text-muted-foreground size-8 opacity-40" />
              <p className="text-muted-foreground text-sm">
                {t("admin.gradebook.noResults")}
              </p>
              <p className="text-muted-foreground text-xs">
                {t("admin.gradebook.noResultsHint")}
              </p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="bg-card sticky start-0 z-10 min-w-56">
                    {t("admin.gradebook.student")}
                  </TableHead>
                  {columns.map((col) => (
                    <TableHead key={col.id} className="min-w-40">
                      <div className="flex items-center justify-between gap-2">
                        <div className="flex min-w-0 flex-col gap-1">
                          <span className="truncate text-sm font-medium">{col.title}</span>
                          <Badge
                            variant={TYPE_BADGE[col.type ?? ""] ?? "outline"}
                            className="w-fit text-[10px]"
                          >
                            {t(`admin.gradebook.types.${col.type}`)}
                          </Badge>
                        </div>
                        <DropdownMenu>
                          <DropdownMenuTrigger
                            render={
                              <Button variant="ghost" size="icon-xs">
                                <EllipsisVerticalIcon />
                              </Button>
                            }
                          />
                          <DropdownMenuContent align="end" className="min-w-40">
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
                            <DropdownMenuSeparator />
                            <DropdownMenuGroup>
                              <DropdownMenuItem
                                variant="destructive"
                                onClick={() => setDeleteColumn(col)}
                              >
                                <Trash2Icon data-icon="inline-start" />
                                {t("common.delete")}
                              </DropdownMenuItem>
                            </DropdownMenuGroup>
                          </DropdownMenuContent>
                        </DropdownMenu>
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
                          <div className="truncate text-sm font-medium">
                            {row.student?.name ?? "—"}
                          </div>
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
                      return (
                        <TableCell
                          key={col.id}
                          className={cn(
                            "text-sm",
                            !isAuto &&
                              "hover:bg-muted/50 cursor-pointer transition-colors"
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
      </Card>

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
          if (deleteColumn?.id) {
            deleteMutation.mutate({ id: classId, columnId: deleteColumn.id })
          }
        }}
        isLoading={deleteMutation.isPending}
      />
    </>
  )
}
