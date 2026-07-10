import { AnimatePresence, motion } from "motion/react"
import { useTranslation } from "react-i18next"

import { typingCopy } from "./lib/typing"
import { useTyping } from "./use-typing"

interface TypingIndicatorProps {
  convId: string
}

/**
 * "X is typing…" strip that sits between the message list and the composer.
 * The wrapper is a FIXED height regardless of whether anyone's typing, so the
 * composer never jumps — only the label + dots fade in/out inside it. Renders
 * an empty (but height-reserving) shell when no one is typing.
 */
export function TypingIndicator({ convId }: TypingIndicatorProps) {
  const { t } = useTranslation()
  const typers = useTyping(convId)
  const copy = typingCopy(typers.map((typer) => typer.name))

  let label: string | null = null
  if (copy?.key === "conversations.typing.one") label = t(copy.key, copy.params)
  else if (copy?.key === "conversations.typing.two") label = t(copy.key, copy.params)
  else if (copy?.key === "conversations.typing.many") label = t(copy.key)

  return (
    <div className="flex h-5 shrink-0 items-center px-4" aria-live="polite" role="status">
      <AnimatePresence>
        {label && (
          <motion.div
            key={copy?.key}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.15 }}
            className="text-muted-foreground flex items-center gap-1.5 text-xs"
          >
            <span className="flex items-center gap-0.5" aria-hidden="true">
              <span className="bg-current animate-typing-bounce size-1 rounded-full" />
              <span className="bg-current animate-typing-bounce size-1 rounded-full [animation-delay:150ms]" />
              <span className="bg-current animate-typing-bounce size-1 rounded-full [animation-delay:300ms]" />
            </span>
            <span>{label}</span>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}
