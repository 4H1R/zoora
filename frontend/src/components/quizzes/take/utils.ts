import type {
  GithubCom4H1RZooraInternalDomainNegativeMarkConfig as NegativeMarkConfig,
  GithubCom4H1RZooraInternalDomainQuestion as Question,
  GithubCom4H1RZooraInternalDomainQuizRoom as QuizRoom,
} from "@/api/model"

import type { AnswerState } from "./types"

export function shuffleSeeded<T>(arr: T[], seed: string): T[] {
  let h = 2166136261
  for (let i = 0; i < seed.length; i++) {
    h = (h ^ seed.charCodeAt(i)) * 16777619
  }
  const out = arr.slice()
  for (let i = out.length - 1; i > 0; i--) {
    h = (h * 1664525 + 1013904223) | 0
    const j = Math.abs(h) % (i + 1)
    ;[out[i], out[j]] = [out[j], out[i]]
  }
  return out
}

export function formatClock(totalSeconds: number) {
  const t = Math.max(0, Math.floor(totalSeconds))
  const h = Math.floor(t / 3600)
  const m = Math.floor((t % 3600) / 60)
  const s = t % 60
  const pad = (n: number) => n.toString().padStart(2, "0")
  return h > 0 ? `${pad(h)}:${pad(m)}:${pad(s)}` : `${pad(m)}:${pad(s)}`
}

export function pickRoomForSession(rooms: QuizRoom[], classSessionId: string): QuizRoom | undefined {
  return rooms.find((r) => r.class_session_id === classSessionId) ?? rooms[0]
}

export function isRoomOpen(room: QuizRoom, nowMs: number): boolean {
  if (!room.started_at) return false
  const start = new Date(room.started_at).getTime()
  if (nowMs < start) return false
  if (!room.ended_at) return true
  return nowMs < new Date(room.ended_at).getTime()
}

export function computeDeadline(startedAtIso: string | undefined, durationMinutes: number, room: QuizRoom): number {
  const startedAt = startedAtIso ? new Date(startedAtIso).getTime() : Date.now()
  const durationMs = durationMinutes * 60_000
  const roomEnd = room.ended_at ? new Date(room.ended_at).getTime() : Infinity
  return Math.min(startedAt + durationMs, roomEnd)
}

export function countAnswered(
  answers: Record<string, AnswerState>,
  order: string[],
  questions: Question[],
): number {
  let n = 0
  for (const qid of order) {
    const q = questions.find((qq) => qq.id === qid)
    const a = answers[qid]
    if (!a || !q) continue
    if (q.type === "choice" && a.selected_option_ids.length > 0) n++
    else if ((q.type === "short_answer" || q.type === "descriptive") && a.value.trim().length > 0) n++
  }
  return n
}

export function hasAnswer(a: AnswerState | undefined): boolean {
  if (!a) return false
  return a.selected_option_ids.length > 0 || a.value.trim().length > 0
}

export function questionTypeKey(type: Question["type"]): "choice" | "short" | "descriptive" {
  if (type === "choice") return "choice"
  if (type === "short_answer") return "short"
  return "descriptive"
}

// formatFraction renders a per-wrong fraction for display; rounds to 2 decimals
// and trims trailing zeros (e.g. 0.5 -> "0.5", 0.33 -> "0.33", 0.2 -> "0.2").
export function formatFraction(value: number): string {
  return value.toFixed(2).replace(/\.?0+$/, "")
}

// penaltyText returns a human-readable per-question penalty string, or null
// when the config has no negative marking.
export function penaltyText(
  cfg: NegativeMarkConfig | undefined,
  t: (key: string, opts?: Record<string, unknown>) => string,
): string | null {
  if (!cfg || !cfg.mode || cfg.mode === "none") return null
  if (cfg.mode === "per_wrong") {
    return t("org.session.quizzes.take.penalty.perWrong", {
      fraction: formatFraction(cfg.fraction ?? cfg.negative_value ?? 0),
    })
  }
  if (cfg.mode === "accumulative") {
    return t("org.session.quizzes.take.penalty.accumulative", {
      count: cfg.wrongs_per_point ?? 0,
    })
  }
  return null
}
