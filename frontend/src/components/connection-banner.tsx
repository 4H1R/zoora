import { Loader2Icon, WifiIcon, WifiOffIcon } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import { useChatWs } from "@/components/org/conversations/chat-provider"
import { useOnline } from "@/hooks/use-online"
import { cn } from "@/lib/utils"

type Tone = "error" | "warning" | "success"

// Solid, fully-opaque fills so the strip is legible over whatever it sits on —
// a tinted `/10` wash let the header bleed through and killed the contrast.
const TONE: Record<Tone, string> = {
  error: "bg-destructive text-white",
  warning: "bg-warning text-warning-foreground",
  success: "bg-success text-white",
}

/**
 * The shared presentational strip: a slim, full-width status bar with a leading
 * icon and a live "pulse" dot. Animates in; live region announces the change to
 * screen readers.
 */
function StatusBar({
  tone,
  icon,
  label,
  hint,
  className,
  assertive,
  pulse = true,
}: {
  tone: Tone
  icon: React.ReactNode
  label: string
  hint?: string
  className?: string
  assertive?: boolean
  pulse?: boolean
}) {
  return (
    <div
      role="status"
      aria-live={assertive ? "assertive" : "polite"}
      className={cn(
        "flex items-center justify-center gap-2.5 px-4 py-2 text-sm font-medium shadow-sm",
        TONE[tone],
        className,
      )}
    >
      <span className="relative flex size-5 shrink-0 items-center justify-center">
        {icon}
        {pulse && (
          <span className="absolute -end-0.5 -top-0.5 flex size-2">
            <span className="absolute inline-flex size-full animate-ping rounded-full bg-current opacity-70" />
            <span className="relative inline-flex size-2 rounded-full bg-current" />
          </span>
        )}
      </span>
      <span className="text-pretty">
        {label}
        {hint && <span className="ms-1.5 font-normal opacity-80">{hint}</span>}
      </span>
    </div>
  )
}

const BACK_ONLINE_MS = 2500

/**
 * Platform-wide connectivity banner. When the browser goes offline NOTHING in
 * the app works — React Query can't refetch, mutations fail silently — so this
 * lives at the app root and covers every route. Pinned to the BOTTOM of the
 * viewport so it never collides with the app header. Offline shows a persistent
 * red strip; when the connection returns it flips to a green "back online" flash
 * for a moment (closure) and then unmounts itself.
 */
export function ConnectionBanner() {
  const { t } = useTranslation()
  const online = useOnline()
  const [showBackOnline, setShowBackOnline] = useState(false)
  const wasOffline = useRef(false)

  useEffect(() => {
    if (!online) {
      wasOffline.current = true
      setShowBackOnline(false)
      return
    }
    // Just recovered from a real outage → brief "back online" flash.
    if (wasOffline.current) {
      wasOffline.current = false
      setShowBackOnline(true)
      const id = setTimeout(() => setShowBackOnline(false), BACK_ONLINE_MS)
      return () => clearTimeout(id)
    }
  }, [online])

  if (online && !showBackOnline) return null

  return (
    <div className="fixed inset-x-0 bottom-0 z-50">
      {online ? (
        <StatusBar
          tone="success"
          pulse={false}
          className="animate-in fade-in-0 slide-in-from-bottom-2 fill-mode-both duration-300"
          icon={<WifiIcon className="size-4" />}
          label={t("connectivity.backOnline")}
        />
      ) : (
        <StatusBar
          tone="error"
          assertive
          className="animate-in fade-in-0 slide-in-from-bottom-2 fill-mode-both duration-300"
          icon={<WifiOffIcon className="size-4" />}
          label={t("connectivity.offline.title")}
          hint={t("connectivity.offline.hint")}
        />
      )}
    </div>
  )
}

/**
 * Conversations-only strip for the chat WebSocket. The socket reconnects with
 * backoff on its own; this just tells the user why new messages paused. Gated to
 * browser-online so it never stacks with the platform-wide offline banner above.
 * Renders in-flow at the top of the chat pane (hence the bottom border).
 */
export function ChatConnectionBanner() {
  const { t } = useTranslation()
  const { status } = useChatWs()
  const online = useOnline()

  if (status === "online" || !online) return null

  return (
    <StatusBar
      tone="warning"
      pulse={false}
      className="border-b"
      icon={<Loader2Icon className="size-4 animate-spin" />}
      label={t("connectivity.chat.reconnecting")}
    />
  )
}
