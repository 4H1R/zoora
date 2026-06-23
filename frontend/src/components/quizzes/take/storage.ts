import type { PersistedQuizState } from "./types"

const STORAGE_PREFIX = "zoora.quiz.take."

function key(submissionId: string) {
  return STORAGE_PREFIX + submissionId
}

export function loadPersistedState(submissionId: string): PersistedQuizState | null {
  try {
    const raw = localStorage.getItem(key(submissionId))
    if (!raw) return null
    return JSON.parse(raw) as PersistedQuizState
  } catch {
    return null
  }
}

export function savePersistedState(submissionId: string, state: PersistedQuizState) {
  try {
    localStorage.setItem(key(submissionId), JSON.stringify(state))
  } catch {
    // quota exceeded or storage unavailable
  }
}

export function clearPersistedState(submissionId: string) {
  try {
    localStorage.removeItem(key(submissionId))
  } catch {
    // storage unavailable — ignore
  }
}
