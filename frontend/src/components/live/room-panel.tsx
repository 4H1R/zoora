import { BarChart3, MessageSquare, Users, X } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Sheet, SheetContent, SheetTitle } from "@/components/ui/sheet"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useIsMobile } from "@/hooks/use-mobile"

import { ChatPanel } from "./panels/chat-panel"
import { PeoplePanel } from "./panels/people-panel"
import { PollsPanel } from "./panels/polls-panel"
import type { RoomRole } from "./room-role"
import type { RoomTab } from "./types"
import type { useRoomChat } from "./use-room-chat"
import type { RoomPolls } from "./use-room-polls"

interface RoomPanelProps {
  tab: RoomTab
  setTab: (tab: RoomTab) => void
  open: boolean
  onClose: () => void
  chat: ReturnType<typeof useRoomChat>
  unread: number
  states: Record<string, { role: RoomRole; handRaised: boolean }>
  isHost: boolean
  liveId: string
  onSetRole: (identity: string, role: "presenter" | "viewer") => void
  onMute: (identity: string, trackSid: string) => void
  polls: RoomPolls
  onVote: (value: string) => void
  answerPending: boolean
}

type TabsInnerProps = Pick<RoomPanelProps, "tab" | "setTab" | "chat" | "unread" | "states" | "isHost" | "liveId" | "onSetRole" | "onMute" | "polls" | "onVote" | "answerPending">

function TabsInner({
  tab,
  setTab,
  chat,
  unread,
  states,
  isHost,
  liveId,
  onSetRole,
  onMute,
  polls,
  onVote,
  answerPending,
}: TabsInnerProps) {
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
        <ChatPanel chat={chat} canModerate={isHost} />
      </TabsContent>
      <TabsContent value="people" className="flex min-h-0 flex-1 flex-col">
        <PeoplePanel
          states={states}
          isHost={isHost}
          onSetRole={onSetRole}
          onMute={onMute}
        />
      </TabsContent>
      <TabsContent value="polls" className="flex min-h-0 flex-1 flex-col">
        <PollsPanel
          liveId={liveId}
          isHost={isHost}
          polls={polls}
          onVote={onVote}
          answerPending={answerPending}
        />
      </TabsContent>
    </Tabs>
  )
}

export function RoomPanel({ tab, setTab, open, onClose, chat, unread, states, isHost, liveId, onSetRole, onMute, polls, onVote, answerPending }: RoomPanelProps) {
  const { t } = useTranslation()
  const isMobile = useIsMobile()
  if (!open) return null

  if (isMobile) {
    return (
      <Sheet open={open} onOpenChange={(v) => !v && onClose()}>
        <SheetContent
          side="bottom"
          className="flex h-[70dvh] flex-col gap-0 bg-card p-0 text-foreground"
        >
          <SheetTitle className="sr-only">{t("liveRoom.panel.title")}</SheetTitle>
          <TabsInner
            tab={tab}
            setTab={setTab}
            chat={chat}
            unread={unread}
            states={states}
            isHost={isHost}
            liveId={liveId}
            onSetRole={onSetRole}
            onMute={onMute}
            polls={polls}
            onVote={onVote}
            answerPending={answerPending}
          />
        </SheetContent>
      </Sheet>
    )
  }

  return (
    <aside className="flex h-full w-80 shrink-0 flex-col border-s border-border bg-card/70 backdrop-blur-xl">
      <div className="flex items-center justify-between border-b border-border px-3 py-2">
        <span className="text-sm font-medium text-foreground">{t("liveRoom.panel.title")}</span>
        <button
          type="button"
          onClick={onClose}
          aria-label={t("common.close")}
          className="flex size-8 items-center justify-center rounded-lg text-muted-foreground hover:bg-accent hover:text-foreground"
        >
          <X className="size-4" />
        </button>
      </div>
      <TabsInner
        tab={tab}
        setTab={setTab}
        chat={chat}
        unread={unread}
        states={states}
        isHost={isHost}
        liveId={liveId}
        onSetRole={onSetRole}
        onMute={onMute}
        polls={polls}
        onVote={onVote}
        answerPending={answerPending}
      />
    </aside>
  )
}
