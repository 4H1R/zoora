import type { GithubCom4H1RZooraInternalDomainQuestionBank as Bank } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { CheckIcon, CopyIcon, Share2Icon } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetQuestionBanksIdShareCodeQueryKey,
  useDeleteQuestionBanksIdShareCode,
  useGetQuestionBanksIdShareCode,
  usePostQuestionBanksIdShareCode,
} from "@/api/question-banks/question-banks"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Spinner } from "@/components/ui/spinner"
import { formatSessionDate } from "@/lib/session-status"
import { cn } from "@/lib/utils"

const EXPIRY_OPTIONS = [
  { id: "7d", days: 7 },
  { id: "30d", days: 30 },
  { id: "never", days: undefined },
] as const

type ExpiryId = (typeof EXPIRY_OPTIONS)[number]["id"]

interface QuestionBankShareDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  bank: Bank | null
}

/** Generate, copy, and revoke a bank's share code. Redeeming the code clones
 * the bank (questions + images) into the redeemer's organization. */
export function QuestionBankShareDialog({ open, onOpenChange, bank }: QuestionBankShareDialogProps) {
  const { t, i18n } = useTranslation()
  const queryClient = useQueryClient()
  const [expiry, setExpiry] = useState<ExpiryId>("7d")
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    if (open) {
      setExpiry("7d")
      setCopied(false)
    }
  }, [open])

  const bankId = bank?.id ?? ""
  const codeQuery = useGetQuestionBanksIdShareCode(bankId, {
    query: { enabled: open && !!bankId, retry: false, staleTime: 0 },
  })
  const shareCode = (codeQuery.data?.status === 200 && codeQuery.data.data.data) || null
  // 404 = no active code yet; anything else pending/errored keeps actions quiet.
  const hasCode = !!shareCode?.code

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetQuestionBanksIdShareCodeQueryKey(bankId) })
  }

  const generateMutation = usePostQuestionBanksIdShareCode({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.session.questionBanks.share.generateSuccess"))
        setCopied(false)
        invalidate()
      },
    },
  })

  const revokeMutation = useDeleteQuestionBanksIdShareCode({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.session.questionBanks.share.revokeSuccess"))
        invalidate()
      },
    },
  })

  const handleGenerate = () => {
    if (!bankId) return
    const days = EXPIRY_OPTIONS.find((o) => o.id === expiry)?.days
    generateMutation.mutate({ id: bankId, data: days ? { expires_in_days: days } : {} })
  }

  const handleCopy = async () => {
    if (!shareCode?.code) return
    await navigator.clipboard.writeText(shareCode.code)
    setCopied(true)
    toast.success(t("org.session.questionBanks.share.copied"))
    setTimeout(() => setCopied(false), 2000)
  }

  const expiresStr = shareCode?.expires_at
    ? t("org.session.questionBanks.share.expiresAt", {
        date: formatSessionDate(shareCode.expires_at, i18n.language, "short"),
      })
    : t("org.session.questionBanks.share.neverExpires")

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Share2Icon className="text-primary size-4 shrink-0" />
            {t("org.session.questionBanks.share.title")}
          </DialogTitle>
          <DialogDescription>{t("org.session.questionBanks.share.description")}</DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-4">
          <p className="truncate text-sm font-medium" dir="auto">
            {bank?.name}
          </p>

          {codeQuery.isPending ? (
            <div className="flex items-center justify-center py-6">
              <Spinner className="size-5" />
            </div>
          ) : hasCode ? (
            <>
              <div className="flex flex-col gap-1.5">
                <span className="text-muted-foreground text-xs font-medium">
                  {t("org.session.questionBanks.share.codeLabel")}
                </span>
                <div className="flex items-center gap-2">
                  <Input
                    readOnly
                    value={shareCode.code}
                    className="font-mono text-base tracking-[0.2em]"
                    dir="ltr"
                  />
                  <Button type="button" size="icon" variant="outline" onClick={handleCopy}>
                    {copied ? <CheckIcon /> : <CopyIcon />}
                  </Button>
                </div>
                <span className="text-muted-foreground text-xs">{expiresStr}</span>
              </div>
              <div className="flex items-center justify-between gap-2">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                  disabled={revokeMutation.isPending}
                  onClick={() => revokeMutation.mutate({ id: bankId })}
                >
                  {revokeMutation.isPending && <Spinner className="size-4" />}
                  {t("org.session.questionBanks.share.revoke")}
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  disabled={generateMutation.isPending}
                  onClick={handleGenerate}
                  title={t("org.session.questionBanks.share.regenerateHint")}
                >
                  {generateMutation.isPending && <Spinner className="size-4" />}
                  {t("org.session.questionBanks.share.regenerate")}
                </Button>
              </div>
            </>
          ) : (
            <>
              <p className="text-muted-foreground text-sm">{t("org.session.questionBanks.share.noCodeHint")}</p>
              <div className="flex flex-col gap-1.5">
                <span className="text-muted-foreground text-xs font-medium">
                  {t("org.session.questionBanks.share.expiryLabel")}
                </span>
                <div className="grid grid-cols-3 gap-2">
                  {EXPIRY_OPTIONS.map((option) => (
                    <Button
                      key={option.id}
                      type="button"
                      size="sm"
                      variant={expiry === option.id ? "default" : "outline"}
                      className={cn(expiry !== option.id && "text-muted-foreground")}
                      onClick={() => setExpiry(option.id)}
                    >
                      {t(`org.session.questionBanks.share.expiry.${option.id}`)}
                    </Button>
                  ))}
                </div>
              </div>
              <Button type="button" disabled={generateMutation.isPending} onClick={handleGenerate}>
                {generateMutation.isPending && <Spinner className="size-4" />}
                {t("org.session.questionBanks.share.generate")}
              </Button>
            </>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
