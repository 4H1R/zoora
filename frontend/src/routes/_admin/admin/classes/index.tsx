import type { GithubCom4H1RZooraInternalDomainClass as Class } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { PlusIcon, SchoolIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetAdminClassesQueryKey,
  useDeleteAdminClassesId,
  useGetAdminClasses,
} from "@/api/admin-classes/admin-classes"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { StatCards } from "@/components/data-table/stat-cards"
import { TableFilter } from "@/components/data-table/table-filter"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { adminHead } from "@/lib/admin-head"
import { adminSearchSchema, useAdminTable } from "@/lib/data-table"

import { ClassFormDialog } from "./-class-form-dialog"
import { useClassColumns } from "./-columns"

export const Route = createFileRoute("/_admin/admin/classes/")({
  head: () => adminHead("admin.classes.title"),
  validateSearch: adminSearchSchema,
  component: ClassesPage,
})

function ClassesPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { search, order_by, order_dir, page } = Route.useSearch()

  const currentPage = page ?? 1

  const [formOpen, setFormOpen] = useState(false)
  const [editingClass, setEditingClass] = useState<Class | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<Class | null>(null)

  const handleEdit = (cls: Class) => {
    setEditingClass(cls)
    setFormOpen(true)
  }

  const handleDelete = (cls: Class) => {
    setDeleteTarget(cls)
  }

  const handleCreate = () => {
    setEditingClass(null)
    setFormOpen(true)
  }

  const { data, isLoading } = useGetAdminClasses({
    search: search || undefined,
    page: currentPage,
    order_by: order_by || undefined,
    order_dir: order_dir || undefined,
  })

  const deleteMutation = useDeleteAdminClassesId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.classes.form.deleteSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetAdminClassesQueryKey() })
        setDeleteTarget(null)
      },
    },
  })

  const classesData = (data?.status === 200 && data.data.data) || undefined
  const classes = classesData?.items ?? []
  const total = classesData?.total ?? 0

  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []
  const columns = useClassColumns({ onEdit: handleEdit, onDelete: handleDelete })

  const table = useAdminTable({
    data: classes,
    columns,
    rowCount: total,
    sorting,
  })

  const statCards = [
    {
      icon: <SchoolIcon />,
      label: t("admin.classes.stats.total"),
      value: total,
      loading: isLoading,
    },
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("admin.classes.title")}
        actions={
          <Button size="sm" onClick={handleCreate}>
            <PlusIcon data-icon="inline-start" />
            {t("admin.classes.newClass")}
          </Button>
        }
      />
      <StatCards stats={statCards} />
      <TableFilter
        table={table}
        searchPlaceholder={t("admin.classes.searchPlaceholder")}
        sortLabel={t("admin.classes.toolbar.sort")}
        columnsLabel={t("admin.classes.toolbar.columns")}
        toggleColumnsLabel={t("admin.classes.toolbar.toggleColumns")}
      />
      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<SchoolIcon className="size-8 opacity-40" />}
            emptyTitle={t("admin.classes.noResults")}
            emptyHint={t("admin.classes.noResultsHint")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>

      <ClassFormDialog open={formOpen} onOpenChange={setFormOpen} cls={editingClass} />

      <DeleteConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open: boolean) => {
          if (!open) setDeleteTarget(null)
        }}
        resourceName={deleteTarget?.name ?? ""}
        onConfirm={() => {
          if (deleteTarget?.id) deleteMutation.mutate({ id: deleteTarget.id })
        }}
        isLoading={deleteMutation.isPending}
      />
    </div>
  )
}
