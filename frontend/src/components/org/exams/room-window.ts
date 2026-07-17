import type {
  GithubCom4H1RZooraInternalDomainQuiz as Quiz,
  GithubCom4H1RZooraInternalDomainQuizRoom as QuizRoom,
} from "@/api/model"

// surfacedRoom picks the room whose window matters most right now:
// the open room, else the next upcoming one, else the most recent past one.
// Mirrors the backend's choice for the student list.
export function surfacedRoom(quiz: Quiz): QuizRoom | undefined {
  const rooms = quiz.rooms ?? []
  const now = Date.now()

  const started = (r: QuizRoom) => (r.started_at ? new Date(r.started_at).getTime() : undefined)
  const ended = (r: QuizRoom) => (r.ended_at ? new Date(r.ended_at).getTime() : undefined)

  const open = rooms.find((r) => {
    const s = started(r)
    if (s === undefined || s > now) return false
    const e = ended(r)
    return e === undefined || e > now
  })
  if (open) return open

  const upcoming = rooms
    .filter((r) => (started(r) ?? 0) > now)
    .sort((a, b) => (started(a) ?? 0) - (started(b) ?? 0))[0]
  if (upcoming) return upcoming

  return rooms
    .filter((r) => started(r) !== undefined)
    .sort((a, b) => (started(b) ?? 0) - (started(a) ?? 0))[0]
}
