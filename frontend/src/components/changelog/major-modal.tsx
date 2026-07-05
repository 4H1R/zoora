import { useQueryClient } from "@tanstack/react-query"
import { Link } from "@tanstack/react-router"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import {
  getGetChangelogStatusQueryKey,
  useGetChangelogStatus,
  usePostChangelogMarkSeen,
} from "@/api/changelog/changelog"
import { ChangelogMarkdown } from "@/components/changelog/markdown"
import { Button, buttonVariants } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"

export function MajorModal() {
  const { t, i18n } = useTranslation()
  const isFa = i18n.language.startsWith("fa")
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)
  const { data } = useGetChangelogStatus()
  const status = (data?.status === 200 && data.data.data) || undefined
  const entry = status?.latest_major

  const markSeen = usePostChangelogMarkSeen({
    mutation: {
      onSuccess: () =>
        queryClient.invalidateQueries({ queryKey: getGetChangelogStatusQueryKey() }),
    },
  })

  useEffect(() => {
    if (status?.has_major_unseen && entry) setOpen(true)
  }, [status?.has_major_unseen, entry])

  function dismiss() {
    setOpen(false)
    markSeen.mutate() // clears server marker → modal won't reappear
  }

  if (!entry) return null
  const title = (isFa && entry.title_fa) || entry.title_en
  const body = (isFa && entry.body_fa) || entry.body_en || ""

  return (
    <Dialog open={open} onOpenChange={(o) => !o && dismiss()}>
      <DialogContent className="max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {entry.version && <DialogDescription>{entry.version}</DialogDescription>}
        </DialogHeader>
        <ChangelogMarkdown>{body}</ChangelogMarkdown>
        <DialogFooter>
          <Link
            to="/org/whats-new"
            onClick={dismiss}
            className={buttonVariants({ variant: "outline" })}
          >
            {t("whatsNew.seeAll")}
          </Link>
          <Button onClick={dismiss}>{t("whatsNew.gotIt")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
