import { useQueryClient } from "@tanstack/react-query"
import { TicketIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetQuestionBanksQueryKey,
  useGetQuestionBanksShareCodesCode,
  usePostQuestionBanksRedeem,
} from "@/api/question-banks/question-banks"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Spinner } from "@/components/ui/spinner"

interface QuestionBankRedeemDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

/** Import a shared question bank: enter a code, preview what it contains,
 * then clone it into the caller's organization. */
export function QuestionBankRedeemDialog({ open, onOpenChange }: QuestionBankRedeemDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [code, setCode] = useState("")
  const [previewCode, setPreviewCode] = useState("")

  useEffect(() => {
    if (open) {
      setCode("")
      setPreviewCode("")
    }
  }, [open])

  const previewQuery = useGetQuestionBanksShareCodesCode(previewCode, {
    query: { enabled: open && previewCode.length >= 4, retry: false, staleTime: 0 },
  })
  const preview = (previewQuery.data?.status === 200 && previewQuery.data.data.data) || null

  const redeemMutation = usePostQuestionBanksRedeem({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.session.questionBanks.redeem.success"))
        queryClient.invalidateQueries({ queryKey: getGetQuestionBanksQueryKey() })
        onOpenChange(false)
      },
    },
  })

  const trimmed = code.trim()
  const checking = previewQuery.isFetching
  const invalid = !!previewCode && !checking && (previewQuery.isError || !preview)

  const handleCheck = () => {
    if (trimmed.length < 4) return
    setPreviewCode(trimmed)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <TicketIcon className="text-primary size-4 shrink-0" />
            {t("org.session.questionBanks.redeem.title")}
          </DialogTitle>
          <DialogDescription>{t("org.session.questionBanks.redeem.description")}</DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-4">
          <form
            className="flex items-center gap-2"
            onSubmit={(e) => {
              e.preventDefault()
              handleCheck()
            }}
          >
            <Input
              value={code}
              onChange={(e) => {
                setCode(e.target.value)
                setPreviewCode("")
              }}
              placeholder={t("org.session.questionBanks.redeem.codePlaceholder")}
              className="font-mono uppercase"
              dir="ltr"
              autoFocus
            />
            <Button type="submit" variant="outline" disabled={trimmed.length < 4 || checking}>
              {checking ? <Spinner className="size-4" /> : t("org.session.questionBanks.redeem.check")}
            </Button>
          </form>

          {invalid && <p className="text-destructive text-sm">{t("org.session.questionBanks.redeem.invalid")}</p>}

          {preview && (
            <>
              <div className="bg-muted/50 ring-foreground/10 flex flex-col gap-1.5 rounded-xl p-4 ring-1">
                <p className="font-medium" dir="auto">
                  {preview.bank_name}
                </p>
                {preview.description && (
                  <p className="text-muted-foreground line-clamp-2 text-sm" dir="auto">
                    {preview.description}
                  </p>
                )}
                <p className="text-muted-foreground font-mono text-xs">
                  {t("org.session.questionBanks.redeem.questionCount", { count: preview.question_count ?? 0 })}
                </p>
              </div>
              <Button
                type="button"
                disabled={redeemMutation.isPending}
                onClick={() => redeemMutation.mutate({ data: { code: previewCode } })}
              >
                {redeemMutation.isPending && <Spinner className="size-4" />}
                {t("org.session.questionBanks.redeem.import")}
              </Button>
            </>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
