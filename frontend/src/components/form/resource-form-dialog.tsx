import type { ReactNode } from "react"

import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Spinner } from "@/components/ui/spinner"

interface ResourceFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description?: string
  children: ReactNode
  onSubmit: React.ComponentProps<"form">["onSubmit"]
  isLoading?: boolean
  submitLabel?: string
}

export function ResourceFormDialog({
  open,
  onOpenChange,
  title,
  description,
  children,
  onSubmit,
  isLoading,
  submitLabel,
}: ResourceFormDialogProps) {
  const { t } = useTranslation()

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description && <DialogDescription>{description}</DialogDescription>}
        </DialogHeader>
        <form onSubmit={onSubmit} className="flex flex-col gap-4">
          {children}
          <DialogFooter>
            <DialogClose render={<Button variant="outline" disabled={isLoading} />}>{t("common.cancel")}</DialogClose>
            <Button type="submit" disabled={isLoading}>
              {isLoading && <Spinner />}
              {submitLabel ?? t("common.save")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
