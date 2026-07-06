import type { ErrorType } from "@/api/mutator/custom-instance"

import { useQueryClient } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetConnectorsQueryKey,
  usePostConnectorsSmsRequestOtp,
  usePostConnectorsSmsVerifyOtp,
} from "@/api/connectors/connectors"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Field, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Spinner } from "@/components/ui/spinner"

interface SmsOtpDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

/** Two-step SMS verification: request an OTP for a phone number, then confirm
 * the 6-digit code. Rate-limit (429) responses surface a friendly retry hint. */
export function SmsOtpDialog({ open, onOpenChange }: SmsOtpDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [step, setStep] = useState<"phone" | "code">("phone")
  const [phone, setPhone] = useState("")
  const [code, setCode] = useState("")

  useEffect(() => {
    if (!open) {
      setStep("phone")
      setPhone("")
      setCode("")
    }
  }, [open])

  const handleError = (error: ErrorType<unknown>) => {
    if (error.response?.status === 429) toast.error(t("notifications.connectors.rateLimited"))
    else toast.error(t("common.error", "Something went wrong"))
  }

  const requestOtp = usePostConnectorsSmsRequestOtp({
    mutation: {
      onSuccess: () => {
        setStep("code")
        toast.success(t("notifications.connectors.codeSent"))
      },
      onError: handleError,
    },
  })

  const verifyOtp = usePostConnectorsSmsVerifyOtp({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: getGetConnectorsQueryKey() })
        toast.success(t("notifications.connectors.connected"))
        onOpenChange(false)
      },
      onError: handleError,
    },
  })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>{t("notifications.connectors.sms")}</DialogTitle>
          <DialogDescription>
            {step === "phone"
              ? t("notifications.connectors.phone")
              : t("notifications.connectors.codeSent")}
          </DialogDescription>
        </DialogHeader>

        {step === "phone" ? (
          <form
            onSubmit={(e) => {
              e.preventDefault()
              if (phone.trim()) requestOtp.mutate({ data: { phone: phone.trim() } })
            }}
          >
            <Field>
              <FieldLabel>{t("notifications.connectors.phone")}</FieldLabel>
              <Input
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                inputMode="tel"
                autoComplete="tel"
                dir="ltr"
                placeholder="09xxxxxxxxx"
              />
            </Field>
            <DialogFooter className="mt-5">
              <Button type="submit" disabled={requestOtp.isPending || !phone.trim()}>
                {requestOtp.isPending && <Spinner />}
                {t("notifications.connectors.sendCode")}
              </Button>
            </DialogFooter>
          </form>
        ) : (
          <form
            onSubmit={(e) => {
              e.preventDefault()
              if (code.trim()) verifyOtp.mutate({ data: { code: code.trim() } })
            }}
          >
            <Field>
              <FieldLabel>{t("notifications.connectors.codeField")}</FieldLabel>
              <Input
                value={code}
                onChange={(e) => setCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
                inputMode="numeric"
                autoComplete="one-time-code"
                dir="ltr"
                className="text-center text-lg tracking-[0.4em] tabular-nums"
                placeholder="······"
              />
            </Field>
            <DialogFooter className="mt-5">
              <Button type="submit" disabled={verifyOtp.isPending || code.length < 4}>
                {verifyOtp.isPending && <Spinner />}
                {t("notifications.connectors.verify")}
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  )
}
