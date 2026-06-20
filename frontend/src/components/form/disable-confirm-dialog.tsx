import { BanIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogMedia,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Label } from "@/components/ui/label"
import { Spinner } from "@/components/ui/spinner"
import { Textarea } from "@/components/ui/textarea"

interface DisableConfirmDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  resourceName: string
  reason: string
  onReasonChange: (reason: string) => void
  onConfirm: () => void
  isLoading?: boolean
}

export function DisableConfirmDialog({
  open,
  onOpenChange,
  resourceName,
  reason,
  onReasonChange,
  onConfirm,
  isLoading,
}: DisableConfirmDialogProps) {
  const { t } = useTranslation()

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent onOutsideClick={() => !isLoading && onOpenChange(false)}>
        <AlertDialogHeader>
          <AlertDialogMedia className="bg-destructive/10 text-destructive">
            <BanIcon />
          </AlertDialogMedia>
          <AlertDialogTitle>{t("common.disableConfirm.title")}</AlertDialogTitle>
          <AlertDialogDescription>
            {t("common.disableConfirm.description", { name: resourceName })}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <div className="grid gap-2">
          <Label htmlFor="disable-reason">{t("common.disableConfirm.reasonLabel")}</Label>
          <Textarea
            id="disable-reason"
            value={reason}
            onChange={(e) => onReasonChange(e.target.value)}
            placeholder={t("common.disableConfirm.reasonPlaceholder")}
            disabled={isLoading}
            maxLength={500}
          />
        </div>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isLoading}>{t("common.cancel")}</AlertDialogCancel>
          <AlertDialogAction variant="destructive" onClick={onConfirm} disabled={isLoading}>
            {isLoading && <Spinner />}
            {t("common.disable")}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
