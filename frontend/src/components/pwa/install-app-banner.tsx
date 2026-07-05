import { useState } from "react"
import { DownloadIcon, PlusSquareIcon, ShareIcon, XIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { usePwaInstall } from "@/components/pwa/use-pwa-install"
import { Button } from "@/components/ui/button"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import { cn } from "@/lib/utils"

/**
 * Contextual, dismissible "install this app" banner. Renders only when the PWA
 * is genuinely installable (see usePwaInstall) — never on unsupported browsers
 * or once already installed. On iOS Safari it opens a Share-sheet instructions
 * sheet since installation there is manual.
 */
export function InstallAppBanner({ className }: { className?: string }) {
  const { t } = useTranslation()
  const { canShow, ios, promptInstall, dismiss } = usePwaInstall()
  const [iosOpen, setIosOpen] = useState(false)

  if (!canShow) return null

  async function onInstall() {
    if (ios) {
      setIosOpen(true)
      return
    }
    await promptInstall()
  }

  return (
    <>
      <div
        className={cn(
          "group relative isolate overflow-hidden rounded-2xl border p-4 sm:p-5",
          "flex flex-col gap-3.5 sm:flex-row sm:items-center sm:gap-4",
          "border-primary/25 bg-gradient-to-br from-primary/12 via-card to-card shadow-sm",
          "animate-in fade-in-0 slide-in-from-top-2 fill-mode-both duration-500",
          className,
        )}
      >
        {/* Slow shimmer sweep — draws the eye without shouting. Forced LTR so it
            always travels left→right regardless of page direction. */}
        <div
          aria-hidden
          dir="ltr"
          className="pointer-events-none absolute inset-0 -z-10 overflow-hidden"
        >
          <div className="animate-install-sweep absolute inset-y-0 left-0 w-1/4 bg-gradient-to-r from-transparent via-primary/10 to-transparent" />
        </div>

        <div className="flex min-w-0 flex-1 items-start gap-3 sm:gap-4">
          {/* Glowing app icon. */}
          <div className="relative grid size-11 shrink-0 place-items-center rounded-xl bg-primary/12 text-primary ring-1 ring-primary/20 ring-inset">
            <span
              aria-hidden
              className="absolute inset-0 animate-pulse rounded-xl bg-primary/25 blur-md"
            />
            <DownloadIcon className="relative size-5" />
          </div>

          <div className="flex min-w-0 flex-1 flex-col gap-0.5">
            <p className="text-sm font-semibold tracking-tight text-balance">
              {t("pwa.install.title")}
            </p>
            <p className="text-muted-foreground text-xs text-pretty sm:text-sm">
              {t("pwa.install.description")}
            </p>
          </div>

          {/* Dismiss lives in the top-end corner on mobile; folds into the row on sm+. */}
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={dismiss}
            aria-label={t("pwa.install.dismiss")}
            className="text-muted-foreground hover:text-foreground -me-1 shrink-0 sm:hidden"
          >
            <XIcon />
          </Button>
        </div>

        <div className="flex items-center gap-2 sm:shrink-0">
          <Button onClick={onInstall} className="flex-1 shadow-sm sm:flex-none">
            <DownloadIcon />
            {t("pwa.install.action")}
          </Button>

          <Button
            variant="ghost"
            size="icon-sm"
            onClick={dismiss}
            aria-label={t("pwa.install.dismiss")}
            className="text-muted-foreground hover:text-foreground -me-1 hidden shrink-0 sm:inline-flex"
          >
            <XIcon />
          </Button>
        </div>
      </div>

      {/* iOS: no install event — show the Share → Add to Home Screen recipe. */}
      <Sheet open={iosOpen} onOpenChange={setIosOpen}>
        <SheetContent side="bottom" className="gap-0">
          <SheetHeader>
            <SheetTitle>{t("pwa.install.ios.title")}</SheetTitle>
            <SheetDescription>{t("pwa.install.ios.description")}</SheetDescription>
          </SheetHeader>
          <ol className="flex flex-col gap-3 px-4 pb-6">
            <li className="flex items-center gap-3">
              <span className="bg-primary/12 text-primary grid size-8 shrink-0 place-items-center rounded-lg text-sm font-semibold">
                1
              </span>
              <span className="text-sm">{t("pwa.install.ios.step1")}</span>
              <ShareIcon className="text-primary ms-auto size-5 shrink-0" />
            </li>
            <li className="flex items-center gap-3">
              <span className="bg-primary/12 text-primary grid size-8 shrink-0 place-items-center rounded-lg text-sm font-semibold">
                2
              </span>
              <span className="text-sm">{t("pwa.install.ios.step2")}</span>
              <PlusSquareIcon className="text-primary ms-auto size-5 shrink-0" />
            </li>
          </ol>
        </SheetContent>
      </Sheet>
    </>
  )
}
