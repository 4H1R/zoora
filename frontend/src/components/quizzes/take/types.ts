export type AnswerState = {
  selected_option_ids: string[]
  value: string
  spent_seconds: number
}

export type PersistedQuizState = {
  answers: Record<string, AnswerState>
  order: string[]
  index: number
  // Anti-cheat tab-switch totals — persisted so a mid-quiz refresh doesn't
  // reset them (they otherwise live only in an in-memory ref). Omitted when
  // track_tab_switches is off.
  tabHidden?: { count: number; seconds: number }
}

export function emptyAnswer(): AnswerState {
  return { selected_option_ids: [], value: "", spent_seconds: 0 }
}
