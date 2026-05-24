export type AnswerState = {
  selected_option_ids: string[]
  value: string
  spent_seconds: number
}

export type PersistedQuizState = {
  answers: Record<string, AnswerState>
  order: string[]
  index: number
}

export function emptyAnswer(): AnswerState {
  return { selected_option_ids: [], value: "", spent_seconds: 0 }
}
