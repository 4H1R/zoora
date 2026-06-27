import type { GithubCom4H1RZooraInternalDomainQuestion as Question } from "@/api/model"

import { useTranslation } from "react-i18next"

import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"

import { OptionTile } from "./option-tile"
import type { AnswerState } from "./types"

interface QuestionInputProps {
  question: Question
  answer: AnswerState
  onChange: (updater: (prev: AnswerState) => AnswerState) => void
}

export function QuestionInput({ question, answer, onChange }: QuestionInputProps) {
  const { t } = useTranslation()
  const opts = question.options ?? []

  if (question.type === "choice") {
    const isMulti = question.is_multi_select ?? false
    return (
      <div className="flex flex-col gap-2">
        {opts.map((opt, i) => {
          const id = opt.id ?? String(i)
          const checked = answer.selected_option_ids.includes(id)
          return (
            <OptionTile
              key={id}
              index={i}
              label={opt.value ?? ""}
              checked={checked}
              imageMediaID={opt.image_media_id}
              onClick={() => onChange((prev) => toggleSelection(prev, id, isMulti))}
            />
          )
        })}
        {isMulti && (
          <span className="text-muted-foreground mt-1 font-mono text-[10px] tracking-[0.25em] uppercase">
            {t("org.session.quizzes.take.multiSelectHint")}
          </span>
        )}
      </div>
    )
  }

  if (question.type === "short_answer") {
    return (
      <Input
        value={answer.value}
        onChange={(e) => onChange((prev) => ({ ...prev, value: e.target.value }))}
        placeholder={t("org.session.quizzes.take.shortPlaceholder")}
        className="max-w-2xl text-base"
      />
    )
  }

  return (
    <Textarea
      value={answer.value}
      onChange={(e) => onChange((prev) => ({ ...prev, value: e.target.value }))}
      placeholder={t("org.session.quizzes.take.descriptivePlaceholder")}
      rows={8}
      className="max-w-3xl text-base"
    />
  )
}

function toggleSelection(prev: AnswerState, id: string, isMulti: boolean): AnswerState {
  const checked = prev.selected_option_ids.includes(id)
  if (isMulti) {
    return {
      ...prev,
      selected_option_ids: checked
        ? prev.selected_option_ids.filter((x) => x !== id)
        : [...prev.selected_option_ids, id],
    }
  }
  return { ...prev, selected_option_ids: checked ? [] : [id] }
}
