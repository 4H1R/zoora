import { useEffect, useRef } from "react"
import { useTranslation } from "react-i18next"
import { registerSW } from "virtual:pwa-register"
import { toast } from "sonner"

/**
 * Registers the service worker and surfaces update / offline-ready state as
 * sonner toasts. We use registerType 'prompt' (see vite.config.js) so a new
 * build never hard-reloads the page underneath a live meeting — the user opts
 * in via the toast action instead.
 */
export function PWAUpdater() {
  const { t } = useTranslation()
  const registered = useRef(false)

  useEffect(() => {
    if (registered.current) return
    registered.current = true

    const updateSW = registerSW({
      onNeedRefresh() {
        toast.info(t("pwa.updateTitle"), {
          duration: Infinity,
          action: {
            label: t("pwa.updateAction"),
            onClick: () => updateSW(true),
          },
        })
      },
      onOfflineReady() {
        toast.success(t("pwa.offlineReady"))
      },
    })
  }, [t])

  return null
}
