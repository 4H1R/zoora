import type { GithubCom4H1RZooraInternalDomainClassSession } from "@/api/model"
import type { ErrorType } from "@/api/mutator/custom-instance"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute, useNavigate } from "@tanstack/react-router"

import { orgHead } from "@/lib/org-head"
import { useOrgGuard } from "@/lib/access"
import { useState } from "react"
import { useAccess } from "react-access-engine"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { z } from "zod"

import {
  getGetClassesIdSessionsQueryKey,
  useGetClassesId,
  useGetClassesIdSessions,
  usePostClassesIdSessions,
} from "@/api/classes/classes"
import { getLiveRooms, usePostLiveRooms } from "@/api/live-sessions/live-sessions"

export const Route = createFileRoute("/_auth/org/$orgId/classes/$classId")({
  head: () => orgHead("org.nav.classes"),
  component: RouteComponent,
})

const sessionSchema = z.object({
  name: z.string().min(2),
  start_time: z.string().min(1),
})

type SessionFormValues = z.infer<typeof sessionSchema>

function LiveSessionButton({ session }: { session: GithubCom4H1RZooraInternalDomainClassSession }) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { can } = useAccess()

  const canStart = can("livesessions:create") || can("livesessions:create_any")
  const canJoin = can("livesessions:join") || can("livesessions:join_any")

  const joinMutation = usePostLiveRooms({
    mutation: {
      onSuccess: (result) => {
        const room = (result.status === 201 && result.data.data) || undefined
        if (room?.id) navigate({ to: "/live/$liveId", params: { liveId: room.id } })
      },
      onError: async (err, variables) => {
        if ((err as ErrorType<unknown>).response?.status === 409) {
          try {
            const rooms = await getLiveRooms()
            const roomsData = (rooms.status === 200 && rooms.data.data) || undefined
            const items = roomsData?.items ?? []
            const room = items.find((r) => r.class_session_id === variables.data.class_session_id)
            if (room?.id) navigate({ to: "/live/$liveId", params: { liveId: room.id } })
          } catch {
            // ignore
          }
        }
      },
    },
  })

  if (!canStart && !canJoin) return null

  return (
    <button
      type="button"
      disabled={joinMutation.isPending}
      onClick={() => joinMutation.mutate({ data: { class_session_id: session.id! } })}
    >
      {canStart ? t("classes.sessions.start") : t("classes.sessions.join")}
    </button>
  )
}

function RouteComponent() {
  const { t } = useTranslation()
  const { classId } = Route.useParams()
  const queryClient = useQueryClient()
  const [showForm, setShowForm] = useState(false)
  const { can } = useAccess()
  const canCreateSession = can("classes:update") || can("classes:update_any")
  const allowed = useOrgGuard(["classes:view", "classes:view_any"])

  const { data: classData, isPending: classPending } = useGetClassesId(classId)
  const { data: sessionsData, isPending: sessionsPending } = useGetClassesIdSessions(classId, undefined)

  const cls = (classData?.status === 200 && classData.data.data) || undefined
  const sessionsResult = (sessionsData?.status === 200 && sessionsData.data.data) || undefined
  const sessions = sessionsResult?.items ?? []

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<SessionFormValues>({
    resolver: zodResolver(sessionSchema),
    defaultValues: { name: "", start_time: "" },
  })

  const createMutation = usePostClassesIdSessions({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: getGetClassesIdSessionsQueryKey(classId) })
        reset()
        setShowForm(false)
      },
    },
  })

  const onSubmit = handleSubmit((values) => {
    createMutation.mutate({
      id: classId,
      data: {
        name: values.name,
        start_time: new Date(values.start_time).toISOString(),
      },
    })
  })

  if (!allowed) return null
  if (classPending) return <div>Loading...</div>

  return (
    <div>
      <h1>{cls?.name}</h1>

      {canCreateSession && (
        <button type="button" onClick={() => setShowForm((v) => !v)}>
          {t("classes.sessions.newSession")}
        </button>
      )}

      {canCreateSession && showForm && (
        <form onSubmit={onSubmit}>
          <div>
            <label>{t("classes.sessions.form.name")}</label>
            <input {...register("name")} />
            {errors.name && <span>{errors.name.message}</span>}
          </div>
          <div>
            <label>{t("classes.sessions.form.startTime")}</label>
            <input type="datetime-local" {...register("start_time")} />
            {errors.start_time && <span>{errors.start_time.message}</span>}
          </div>
          <button type="submit" disabled={createMutation.isPending}>
            {t("common.create")}
          </button>
          <button type="button" onClick={() => setShowForm(false)}>
            {t("common.cancel")}
          </button>
        </form>
      )}

      {sessionsPending ? (
        <div>Loading...</div>
      ) : sessions.length === 0 ? (
        <div>{t("classes.sessions.empty")}</div>
      ) : (
        <ul>
          {sessions.map((s) => (
            <li key={s.id}>
              <strong>{s.name}</strong> — {s.start_time}
              <LiveSessionButton session={s} />
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
