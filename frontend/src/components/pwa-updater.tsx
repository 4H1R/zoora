import { useEffect, useRef } from "react"
import { useTranslation } from "react-i18next"
import { registerSW } from "virtual:pwa-register"
import { toast } from "sonner"

/**
 * Registers the service worker and surfaces offline-ready state as a sonner
 * toast. registerType is 'autoUpdate' (see vite.config.js): a new build
 * activates and reloads automatically, so onNeedRefresh below is a no-op safety
 * net rather than the primary update path.
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
