import type { RoomRole } from "./room-role"
import type { RoomTab } from "./types"
import type { useRoomChat } from "./use-room-chat"
import type { RoomPolls } from "./use-room-polls"
import type { RoomQa } from "./use-room-qa"

import { BarChart3, MessageCircleQuestion, MessageSquare, Users, X } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Sheet, SheetContent, SheetTitle } from "@/components/ui/sheet"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useIsMobile } from "@/hooks/use-mobile"

import { ChatPanel } from "./panels/chat-panel"
import { PeoplePanel } from "./panels/people-panel"
import { PollsPanel } from "./panels/polls-panel"
import { QaPanel } from "./panels/qa-panel"

interface RoomPanelProps {
  tab: RoomTab
  setTab: (tab: RoomTab) => void
  open: boolean
  onClose: () => void
  chat: ReturnType<typeof useRoomChat>
  unread: number
  states: Record<string, { role: RoomRole; handRaised: boolean; handRaisedAt?: number }>
  isHost: boolean
  liveId: string
  onSetRole: (identity: string, role: "presenter" | "viewer") => void
  onMute: (identity: string, trackSid: string) => void
  onLowerHand: (identity: string) => void
  onRemove: (identity: string, name: string) => void
  polls: RoomPolls
  onVote: (value: string) => void
  answerPending: boolean
  qa: RoomQa
  myId: string
}

type TabsInnerProps = Pick<
  RoomPanelProps,
  | "tab"
  | "setTab"
  | "chat"
  | "unread"
  | "states"
  | "isHost"
  | "liveId"
  | "onSetRole"
  | "onMute"
  | "onLowerHand"
  | "onRemove"
  | "polls"
  | "onVote"
  | "answerPending"
  | "qa"
  | "myId"
>

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
  onLowerHand,
  onRemove,
  polls,
  onVote,
  answerPending,
  qa,
  myId,
}: TabsInnerProps) {
  const { t } = useTranslation()
  return (
    <Tabs value={tab} onValueChange={(v) => setTab(v as RoomTab)} className="flex min-h-0 flex-1 flex-col">
      <TabsList className="mx-2.5 mt-2.5 flex w-auto [scrollbar-width:none] gap-1 overflow-x-auto [&::-webkit-scrollbar]:hidden">
        <TabsTrigger value="chat" className="flex-1 gap-1.5">
          <MessageSquare className="size-4" />
          <span className="hidden sm:inline">{t("liveRoom.controls.chat")}</span>
          {unread > 0 && tab !== "chat" && (
            <span className="bg-primary text-primary-foreground ms-1 flex min-w-4 items-center justify-center rounded-full px-1 text-[10px] font-semibold">
              {unread > 9 ? "9+" : unread}
            </span>
          )}
        </TabsTrigger>
        <TabsTrigger value="qa" className="flex-1 gap-1.5">
          <MessageCircleQuestion className="size-4" />
          <span className="hidden sm:inline">{t("liveRoom.controls.qa")}</span>
          {qa.openCount > 0 && tab !== "qa" && (
            <span className="bg-primary text-primary-foreground ms-1 flex min-w-4 items-center justify-center rounded-full px-1 text-[10px] font-semibold">
              {qa.openCount > 9 ? "9+" : qa.openCount}
            </span>
          )}
        </TabsTrigger>
        <TabsTrigger value="people" className="flex-1 gap-1.5">
          <Users className="size-4" />
          <span className="hidden sm:inline">{t("liveRoom.controls.people")}</span>
        </TabsTrigger>
        <TabsTrigger value="polls" className="flex-1 gap-1.5">
          <BarChart3 className="size-4" />
          <span className="hidden sm:inline">{t("liveRoom.controls.polls")}</span>
        </TabsTrigger>
      </TabsList>
      <TabsContent value="chat" className="flex min-h-0 flex-1 flex-col">
        <ChatPanel chat={chat} canModerate={isHost} />
      </TabsContent>
      <TabsContent value="qa" className="flex min-h-0 flex-1 flex-col">
        <QaPanel qa={qa} isHost={isHost} myId={myId} />
      </TabsContent>
      <TabsContent value="people" className="flex min-h-0 flex-1 flex-col">
        <PeoplePanel
          states={states}
          isHost={isHost}
          onSetRole={onSetRole}
          onMute={onMute}
          onLowerHand={onLowerHand}
          onRemove={onRemove}
        />
      </TabsContent>
      <TabsContent value="polls" className="flex min-h-0 flex-1 flex-col">
        <PollsPanel liveId={liveId} isHost={isHost} polls={polls} onVote={onVote} answerPending={answerPending} />
      </TabsContent>
    </Tabs>
  )
}

export function RoomPanel({
  tab,
  setTab,
  open,
  onClose,
  chat,
  unread,
  states,
  isHost,
  liveId,
  onSetRole,
  onMute,
  onLowerHand,
  onRemove,
  polls,
  onVote,
  answerPending,
  qa,
  myId,
}: RoomPanelProps) {
  const { t } = useTranslation()
  const isMobile = useIsMobile()
  if (!open) return null

  if (isMobile) {
    return (
      <Sheet open={open} onOpenChange={(v) => !v && onClose()}>
        <SheetContent
          side="bottom"
          className="bg-card text-foreground flex h-[70dvh] flex-col gap-0 p-0 data-[side=bottom]:h-[70dvh]"
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
            onLowerHand={onLowerHand}
            onRemove={onRemove}
            polls={polls}
            onVote={onVote}
            answerPending={answerPending}
            qa={qa}
            myId={myId}
          />
        </SheetContent>
      </Sheet>
    )
  }

  return (
    <aside className="border-border bg-card flex h-full w-96 shrink-0 flex-col border-s">
      <div className="border-border flex items-center justify-between border-b px-3 py-2">
        <span className="text-foreground text-sm font-medium">{t("liveRoom.panel.title")}</span>
        <button
          type="button"
          onClick={onClose}
          aria-label={t("common.close")}
          className="text-muted-foreground hover:bg-accent hover:text-foreground flex size-8 items-center justify-center rounded-lg"
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
        onLowerHand={onLowerHand}
        onRemove={onRemove}
        polls={polls}
        onVote={onVote}
        answerPending={answerPending}
        qa={qa}
        myId={myId}
      />
    </aside>
  )
}
