import type { useChat } from "@livekit/components-react"
import { BarChart3, MessageSquare, Users, X } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Sheet, SheetContent, SheetTitle } from "@/components/ui/sheet"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { cn } from "@/lib/utils"

import { ChatPanel } from "./panels/chat-panel"
import { PeoplePanel } from "./panels/people-panel"
import { PollsPanel } from "./panels/polls-panel"
import type { RoomTab } from "./types"

interface RoomPanelProps {
  tab: RoomTab
  setTab: (tab: RoomTab) => void
  open: boolean
  onClose: () => void
  chat: ReturnType<typeof useChat>
  unread: number
}

function TabsInner({
  tab,
  setTab,
  chat,
  unread,
}: Pick<RoomPanelProps, "tab" | "setTab" | "chat" | "unread">) {
  const { t } = useTranslation()
  return (
    <Tabs
      value={tab}
      onValueChange={(v) => setTab(v as RoomTab)}
      className="flex min-h-0 flex-1 flex-col"
    >
      <TabsList className="mx-2.5 mt-2.5 w-auto grid grid-cols-3">
        <TabsTrigger value="chat" className="gap-1.5">
          <MessageSquare className="size-4" />
          <span className="hidden sm:inline">{t("liveRoom.controls.chat")}</span>
          {unread > 0 && tab !== "chat" && (
            <span className="ms-1 flex min-w-4 items-center justify-center rounded-full bg-primary px-1 text-[10px] font-semibold text-primary-foreground">
              {unread > 9 ? "9+" : unread}
            </span>
          )}
        </TabsTrigger>
        <TabsTrigger value="people" className="gap-1.5">
          <Users className="size-4" />
          <span className="hidden sm:inline">{t("liveRoom.controls.people")}</span>
        </TabsTrigger>
        <TabsTrigger value="polls" className="gap-1.5">
          <BarChart3 className="size-4" />
          <span className="hidden sm:inline">{t("liveRoom.controls.polls")}</span>
        </TabsTrigger>
      </TabsList>
      <TabsContent value="chat" className="flex min-h-0 flex-1 flex-col">
        <ChatPanel chat={chat} />
      </TabsContent>
      <TabsContent value="people" className="flex min-h-0 flex-1 flex-col">
        <PeoplePanel />
      </TabsContent>
      <TabsContent value="polls" className="flex min-h-0 flex-1 flex-col">
        <PollsPanel />
      </TabsContent>
    </Tabs>
  )
}

export function RoomPanel({ tab, setTab, open, onClose, chat, unread }: RoomPanelProps) {
  const { t } = useTranslation()
  if (!open) return null

  return (
    <>
      {/* Desktop: side dock — hidden on mobile */}
      <aside className="hidden h-full w-80 shrink-0 flex-col border-s border-white/10 bg-zinc-900/70 backdrop-blur-xl md:flex">
        <div className="flex items-center justify-between border-b border-white/10 px-3 py-2">
          <span className="text-sm font-medium text-zinc-200">{t("liveRoom.panel.title")}</span>
          <button
            type="button"
            onClick={onClose}
            aria-label={t("common.close")}
            className="flex size-8 items-center justify-center rounded-lg text-zinc-400 hover:bg-white/5 hover:text-zinc-100"
          >
            <X className="size-4" />
          </button>
        </div>
        <TabsInner tab={tab} setTab={setTab} chat={chat} unread={unread} />
      </aside>

      {/* Mobile: bottom Sheet — hidden on desktop */}
      <Sheet open={open} onOpenChange={(v) => !v && onClose()}>
        <SheetContent
          side="bottom"
          className={cn(
            "flex h-[70dvh] flex-col gap-0 bg-zinc-900 p-0 text-zinc-100 md:hidden",
          )}
          showCloseButton={false}
        >
          <SheetTitle className="sr-only">{t("liveRoom.panel.title")}</SheetTitle>
          <TabsInner tab={tab} setTab={setTab} chat={chat} unread={unread} />
        </SheetContent>
      </Sheet>
    </>
  )
}
