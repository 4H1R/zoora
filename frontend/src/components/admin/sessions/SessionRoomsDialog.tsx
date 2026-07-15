import type { GithubCom4H1RZooraInternalDomainClassSession as Session } from "@/api/model"

import { Link } from "@tanstack/react-router"
import { ClipboardListIcon, DumbbellIcon, FileVideoIcon, VideoIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetAdminLiveRooms } from "@/api/admin-livesessions/admin-livesessions"
import { useGetAdminOfflines } from "@/api/admin-offlines/admin-offlines"
import { useGetAdminPractices } from "@/api/admin-practices/admin-practices"
import { useGetAdminQuizzes } from "@/api/admin-quizzes/admin-quizzes"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useFormatDate } from "@/lib/data-table"

interface SessionRoomsDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  session: Session | null
  classId: string
}

export function SessionRoomsDialog({ open, onOpenChange, session, classId }: SessionRoomsDialogProps) {
  const { t } = useTranslation()
  const sessionId = session?.id

  const { data: liveData, isLoading: liveLoading } = useGetAdminLiveRooms(
    { class_session_id: sessionId ?? "" },
    { query: { enabled: open && !!sessionId } }
  )
  const { data: offlineData, isLoading: offlineLoading } = useGetAdminOfflines(
    { class_session_id: sessionId ?? "" },
    { query: { enabled: open && !!sessionId } }
  )
  const { data: practiceData, isLoading: practiceLoading } = useGetAdminPractices(
    { class_session_id: sessionId ?? "" },
    { query: { enabled: open && !!sessionId } }
  )
  const { data: quizData, isLoading: quizLoading } = useGetAdminQuizzes(
    { class_id: classId },
    { query: { enabled: open && !!classId } }
  )

  const liveRooms = (liveData?.status === 200 && liveData.data.data?.items) || []
  const offlineRooms = (offlineData?.status === 200 && offlineData.data.data?.items) || []
  const practiceRooms = (practiceData?.status === 200 && practiceData.data.data?.items) || []
  const allQuizzes = (quizData?.status === 200 && quizData.data.data?.items) || []
  // Quizzes are class-scoped. Filter to those whose rooms reference this session.
  const sessionQuizzes = sessionId
    ? allQuizzes.filter((q) => (session?.quiz_rooms ?? []).some((r) => r.quiz_id === q.id))
    : []
  // Fall back to all class quizzes when session has no preloaded rooms.
  const quizzesToShow = sessionQuizzes.length > 0 ? sessionQuizzes : allQuizzes

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="flex max-h-[85vh] w-full flex-col overflow-hidden sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>
            {t("admin.sessions.manageRooms.title")}
            {session?.name && <span className="text-muted-foreground ms-2 text-sm font-normal">· {session.name}</span>}
          </DialogTitle>
          <DialogDescription>{t("admin.sessions.manageRooms.description")}</DialogDescription>
        </DialogHeader>

        <Tabs defaultValue="live" className="min-h-0 flex-1 overflow-hidden">
          <TabsList className="w-full max-w-full overflow-x-auto">
            <TabsTrigger value="live">
              <VideoIcon data-icon="inline-start" />
              {t("admin.sessions.manageRooms.tabs.live")}
              <Badge variant="secondary" className="ms-1.5">
                {liveRooms.length}
              </Badge>
            </TabsTrigger>
            <TabsTrigger value="offline">
              <FileVideoIcon data-icon="inline-start" />
              {t("admin.sessions.manageRooms.tabs.offline")}
              <Badge variant="secondary" className="ms-1.5">
                {offlineRooms.length}
              </Badge>
            </TabsTrigger>
            <TabsTrigger value="practice">
              <DumbbellIcon data-icon="inline-start" />
              {t("admin.sessions.manageRooms.tabs.practice")}
              <Badge variant="secondary" className="ms-1.5">
                {practiceRooms.length}
              </Badge>
            </TabsTrigger>
            <TabsTrigger value="quiz">
              <ClipboardListIcon data-icon="inline-start" />
              {t("admin.sessions.manageRooms.tabs.quiz")}
              <Badge variant="secondary" className="ms-1.5">
                {quizzesToShow.length}
              </Badge>
            </TabsTrigger>
          </TabsList>

          <TabsContent value="live" className="min-h-0 overflow-y-auto">
            <RoomList
              isLoading={liveLoading}
              empty={t("admin.sessions.manageRooms.empty.live")}
              items={liveRooms.map((r) => ({
                id: r.id ?? "",
                primary: r.livekit_room_name ?? r.class_session?.name ?? "—",
                secondary: r.status,
              }))}
              viewAll={
                <Link to="/admin/classes/$classId/live-rooms" params={{ classId }} onClick={() => onOpenChange(false)}>
                  <Button variant="outline" size="sm">
                    {t("admin.sessions.manageRooms.viewAll")}
                  </Button>
                </Link>
              }
            />
          </TabsContent>

          <TabsContent value="offline" className="min-h-0 overflow-y-auto">
            <RoomList
              isLoading={offlineLoading}
              empty={t("admin.sessions.manageRooms.empty.offline")}
              items={offlineRooms.map((r) => ({
                id: r.id ?? "",
                primary: r.title ?? "—",
                secondary: r.description,
              }))}
              viewAll={
                <Link to="/admin/classes/$classId/offlines" params={{ classId }} onClick={() => onOpenChange(false)}>
                  <Button variant="outline" size="sm">
                    {t("admin.sessions.manageRooms.viewAll")}
                  </Button>
                </Link>
              }
            />
          </TabsContent>

          <TabsContent value="quiz" className="min-h-0 overflow-y-auto">
            <RoomList
              isLoading={quizLoading}
              empty={t("admin.sessions.manageRooms.empty.quiz")}
              items={quizzesToShow.map((q) => ({
                id: q.id ?? "",
                primary: q.title ?? "—",
                secondary: q.description,
              }))}
              viewAll={
                <Link to="/admin/classes/$classId/quizzes" params={{ classId }} onClick={() => onOpenChange(false)}>
                  <Button variant="outline" size="sm">
                    {t("admin.sessions.manageRooms.viewAll")}
                  </Button>
                </Link>
              }
            />
          </TabsContent>

          <TabsContent value="practice" className="min-h-0 overflow-y-auto">
            <PracticeList
              isLoading={practiceLoading}
              empty={t("admin.sessions.manageRooms.empty.practice")}
              items={practiceRooms.map((r) => ({
                id: r.id ?? "",
                primary: r.title ?? "—",
                start_time: r.start_time,
              }))}
              viewAll={
                <Link to="/admin/classes/$classId/practices" params={{ classId }} onClick={() => onOpenChange(false)}>
                  <Button variant="outline" size="sm">
                    {t("admin.sessions.manageRooms.viewAll")}
                  </Button>
                </Link>
              }
            />
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  )
}

interface RoomListItem {
  id: string
  primary: string
  secondary?: string
}

function RoomList({
  isLoading,
  empty,
  items,
  viewAll,
}: {
  isLoading: boolean
  empty: string
  items: RoomListItem[]
  viewAll: React.ReactNode
}) {
  if (isLoading) {
    return <div className="text-muted-foreground py-6 text-center text-sm">…</div>
  }
  if (items.length === 0) {
    return (
      <div className="flex flex-col items-center gap-3 py-6">
        <div className="text-muted-foreground text-sm">{empty}</div>
        {viewAll}
      </div>
    )
  }
  return (
    <div className="flex flex-col gap-2">
      <ul className="divide-border divide-y rounded-md border">
        {items.map((it) => (
          <li key={it.id} className="flex items-center justify-between px-3 py-2">
            <div className="min-w-0">
              <div className="truncate text-sm font-medium">{it.primary}</div>
              {it.secondary && <div className="text-muted-foreground truncate text-xs">{it.secondary}</div>}
            </div>
          </li>
        ))}
      </ul>
      <div className="flex justify-end">{viewAll}</div>
    </div>
  )
}

function PracticeList({
  isLoading,
  empty,
  items,
  viewAll,
}: {
  isLoading: boolean
  empty: string
  items: { id: string; primary: string; start_time?: string }[]
  viewAll: React.ReactNode
}) {
  const formatDate = useFormatDate()
  if (isLoading) {
    return <div className="text-muted-foreground py-6 text-center text-sm">…</div>
  }
  if (items.length === 0) {
    return (
      <div className="flex flex-col items-center gap-3 py-6">
        <div className="text-muted-foreground text-sm">{empty}</div>
        {viewAll}
      </div>
    )
  }
  return (
    <div className="flex flex-col gap-2">
      <ul className="divide-border divide-y rounded-md border">
        {items.map((it) => (
          <li key={it.id} className="flex items-center justify-between px-3 py-2">
            <div className="min-w-0">
              <div className="truncate text-sm font-medium">{it.primary}</div>
              {it.start_time && (
                <div className="text-muted-foreground truncate text-xs">{formatDate(it.start_time)}</div>
              )}
            </div>
          </li>
        ))}
      </ul>
      <div className="flex justify-end">{viewAll}</div>
    </div>
  )
}
