import type { ChatMessage } from "./lib/messages"
import type { ReactNode } from "react"

import { PencilIcon, PinIcon, PinOffIcon, ReplyIcon, Trash2Icon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from "@/components/ui/context-menu"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { useCoarsePointer } from "@/hooks/use-coarse-pointer"

import { QUICK_EMOJIS } from "./lib/reactions"
import { useMessageActions } from "./use-message-actions"

interface MessageContextMenuProps {
  message: ChatMessage
  /** Whether the signed-in user authored this message — gates edit/delete. */
  isOwn: boolean
  convId: string
  /** The bubble to wrap; the tap/right-click surface. */
  children: ReactNode
}

/** A menu item component — satisfied by both Context and Dropdown item variants. */
type MenuItemComp = (props: {
  onClick?: () => void
  variant?: "default" | "destructive"
  className?: string
  children?: ReactNode
}) => ReactNode
type MenuSeparatorComp = (props: Record<string, never>) => ReactNode

/**
 * The message action surface. On touch, tapping the bubble opens a dropdown
 * anchored to it (Telegram-style); on desktop, right-clicking opens the native
 * context menu. Both render the exact same items — a quick reaction row plus
 * Reply / Pin / Edit / Delete — sharing handlers via {@link useMessageActions}.
 * Delete confirms through a single AlertDialog.
 */
export function MessageContextMenu({ message, isOwn, convId, children }: MessageContextMenuProps) {
  const { t } = useTranslation()
  const actions = useMessageActions(message, convId)
  const [confirmOpen, setConfirmOpen] = useState(false)
  const coarse = useCoarsePointer()

  // Items shared verbatim between the two menus; only the primitives differ.
  const items = (Item: MenuItemComp, Separator: MenuSeparatorComp) => (
    <>
      {/* Quick reaction row — one tap toggles the emoji, then closes. */}
      <div className="flex items-center gap-0.5 px-1 pb-1">
        {QUICK_EMOJIS.map((emoji) => (
          <Item key={emoji} onClick={() => actions.react(emoji)} className="size-8 justify-center p-0 text-base">
            {emoji}
          </Item>
        ))}
      </div>

      <Separator />

      <Item onClick={actions.reply}>
        <ReplyIcon className="rtl:-scale-x-100" />
        {t("conversations.actions.reply")}
      </Item>

      <Item onClick={actions.togglePin}>
        {message.is_pinned ? <PinOffIcon /> : <PinIcon />}
        {message.is_pinned ? t("conversations.actions.unpin") : t("conversations.actions.pin")}
      </Item>

      {isOwn && (
        <>
          <Item onClick={actions.edit}>
            <PencilIcon />
            {t("conversations.actions.edit")}
          </Item>

          <Separator />

          <Item variant="destructive" onClick={() => setConfirmOpen(true)}>
            <Trash2Icon />
            {t("conversations.actions.delete")}
          </Item>
        </>
      )}
    </>
  )

  return (
    <>
      {coarse ? (
        <DropdownMenu>
          {/* Tap the bubble to open; w-fit hugs it, select-none avoids stray selection. */}
          <DropdownMenuTrigger render={<div className="w-fit cursor-pointer select-none">{children}</div>} />
          <DropdownMenuContent align="center" side="top" className="min-w-44">
            {items(DropdownMenuItem, DropdownMenuSeparator)}
          </DropdownMenuContent>
        </DropdownMenu>
      ) : (
        <ContextMenu>
          {/* w-fit hugs the bubble; select-text keeps copy working. */}
          <ContextMenuTrigger className="w-fit select-text">{children}</ContextMenuTrigger>
          <ContextMenuContent className="min-w-44">{items(ContextMenuItem, ContextMenuSeparator)}</ContextMenuContent>
        </ContextMenu>
      )}

      <AlertDialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("conversations.actions.deleteConfirm.title")}</AlertDialogTitle>
            <AlertDialogDescription>{t("conversations.actions.deleteConfirm.description")}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t("common.cancel")}</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={() => {
                setConfirmOpen(false)
                actions.remove()
              }}
            >
              {t("common.delete")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
