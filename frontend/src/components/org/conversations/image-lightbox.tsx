import { useTranslation } from "react-i18next"

import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog"
import { useImageLightbox } from "@/stores/image-lightbox"

/**
 * The single full-size image viewer. Driven by {@link useImageLightbox} and
 * mounted once at the conversations route root, well outside any message
 * bubble's context menu, so dismissing it can't reopen that menu.
 */
export function ImageLightbox() {
  const { t } = useTranslation()
  const { src, alt, close } = useImageLightbox()
  const open = src !== null

  return (
    <Dialog open={open} onOpenChange={(next) => !next && close()}>
      <DialogContent className="border-0 bg-transparent p-0 shadow-none sm:max-w-3xl">
        <DialogTitle className="sr-only">{alt || t("conversations.attachments.image")}</DialogTitle>
        {src && <img src={src} alt={alt} className="max-h-[80vh] w-full rounded-lg object-contain" />}
      </DialogContent>
    </Dialog>
  )
}
