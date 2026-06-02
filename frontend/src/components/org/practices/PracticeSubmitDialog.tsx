import type { GithubCom4H1RZooraInternalDomainPracticeRoom as PracticeRoom } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetPracticesQueryKey,
  usePostPracticesIdSubmissions,
} from "@/api/practices/practices"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Textarea } from "@/components/ui/textarea"

interface PracticeSubmitDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  practice: PracticeRoom | null
}

export function PracticeSubmitDialog({ open, onOpenChange, practice }: PracticeSubmitDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [content, setContent] = useState("")

  useEffect(() => {
    if (open) setContent("")
  }, [open])

  const submitMutation = usePostPracticesIdSubmissions({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.session.practices.submit.success"))
        queryClient.invalidateQueries({ queryKey: getGetPracticesQueryKey() })
        onOpenChange(false)
      },
      onError: () => {
        toast.error(t("org.session.practices.submit.failed"))
      },
    },
  })

  const handleSubmit = () => {
    if (!practice?.id) return
    submitMutation.mutate({ id: practice.id, data: { content } })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t("org.session.practices.submit.title")}</DialogTitle>
          <DialogDescription>
            {practice?.title
              ? t("org.session.practices.submit.descriptionNamed", { title: practice.title })
              : t("org.session.practices.submit.description")}
          </DialogDescription>
        </DialogHeader>

        <Textarea
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder={t("org.session.practices.submit.placeholder")}
          rows={8}
        />

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={submitMutation.isPending}>
            {t("common.cancel")}
          </Button>
          <Button onClick={handleSubmit} disabled={submitMutation.isPending || !content.trim()}>
            {t("org.session.practices.submit.submit")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
