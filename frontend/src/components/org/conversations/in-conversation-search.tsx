import type { GithubCom4H1RZooraInternalDomainConversationMessage as ConversationMessage } from "@/api/model"

import { ChevronDownIcon, ChevronUpIcon, SearchIcon, XIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { useDebounce } from "use-debounce"

import { useGetConversationsIdSearch } from "@/api/conversations/conversations"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

import { useJumpToMessage } from "./jump-context"
import { matchesKey, nextMatchIndex } from "./lib/search"

// Short queries add noise for an ILIKE substring scan; wait for a real term.
const MIN_QUERY = 2
const DEBOUNCE_MS = 300

interface InConversationSearchProps {
  convId: string
}

/**
 * Toggleable in-thread search, mounted in the thread header. A search icon opens
 * a debounced input; matches drive a `current / total` counter and prev/next
 * chevrons that cycle (with wraparound) through the matching messages, jumping +
 * highlighting each via the thread's jump channel. Esc closes and clears.
 */
export function InConversationSearch({ convId }: InConversationSearchProps) {
  const { t } = useTranslation()
  const jumpToMessage = useJumpToMessage()

  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState("")
  const [debounced] = useDebounce(query, DEBOUNCE_MS)
  const [current, setCurrent] = useState(0)

  const trimmed = debounced.trim()
  const enabled = open && trimmed.length >= MIN_QUERY

  const { data } = useGetConversationsIdSearch(convId, { q: trimmed }, { query: { enabled } })
  const matches: ConversationMessage[] = data?.status === 200 ? (data.data.data ?? []) : []
  const key = matchesKey(matches)

  // A fresh result set → reset the cursor to the first match and jump to it.
  useEffect(() => {
    if (!enabled) return
    if (matches.length === 0) {
      setCurrent(0)
      return
    }
    setCurrent(0)
    if (matches[0].id) jumpToMessage(matches[0].id)
    // Re-run only when the *set* of matches changes, not on every render.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [key, enabled])

  function cycle(dir: 1 | -1) {
    const next = nextMatchIndex(current, matches.length, dir)
    if (next < 0) return
    setCurrent(next)
    const id = matches[next].id
    if (id) jumpToMessage(id)
  }

  function close() {
    setOpen(false)
    setQuery("")
    setCurrent(0)
  }

  if (!open) {
    return (
      <Button
        variant="ghost"
        size="icon-sm"
        aria-label={t("conversations.search.inThread.open")}
        title={t("conversations.search.inThread.open")}
        onClick={() => setOpen(true)}
      >
        <SearchIcon />
      </Button>
    )
  }

  const hasQuery = trimmed.length >= MIN_QUERY
  const total = matches.length
  const counter = total > 0 ? t("conversations.search.inThread.count", { current: current + 1, total }) : ""

  return (
    // On mobile the open panel takes over the whole header row (avatar + title
    // don't fit alongside the input); on `sm`+ it sits inline in the actions slot.
    <div className="bg-background absolute inset-x-0 inset-y-0 z-10 flex items-center gap-1 px-4 sm:static sm:z-auto sm:bg-transparent sm:px-0">
      <div className="relative min-w-0 flex-1 sm:flex-none">
        <SearchIcon className="text-muted-foreground pointer-events-none absolute start-2.5 top-1/2 size-4 -translate-y-1/2" />
        <Input
          autoFocus
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Escape") {
              e.preventDefault()
              close()
            } else if (e.key === "Enter") {
              e.preventDefault()
              cycle(e.shiftKey ? -1 : 1)
            }
          }}
          placeholder={t("conversations.search.inThread.placeholder")}
          className="h-8 w-full ps-9 pe-16 sm:w-64"
        />
        <span className="text-muted-foreground pointer-events-none absolute end-2.5 top-1/2 -translate-y-1/2 font-mono text-xs tabular-nums">
          {hasQuery ? (total > 0 ? counter : t("conversations.search.inThread.empty")) : ""}
        </span>
      </div>

      <Button
        variant="ghost"
        size="icon-sm"
        aria-label={t("conversations.search.inThread.previous")}
        title={t("conversations.search.inThread.previous")}
        disabled={total === 0}
        onClick={() => cycle(-1)}
      >
        <ChevronUpIcon />
      </Button>
      <Button
        variant="ghost"
        size="icon-sm"
        aria-label={t("conversations.search.inThread.next")}
        title={t("conversations.search.inThread.next")}
        disabled={total === 0}
        onClick={() => cycle(1)}
      >
        <ChevronDownIcon />
      </Button>
      <Button
        variant="ghost"
        size="icon-sm"
        aria-label={t("conversations.search.inThread.close")}
        title={t("conversations.search.inThread.close")}
        onClick={close}
      >
        <XIcon />
      </Button>
    </div>
  )
}
