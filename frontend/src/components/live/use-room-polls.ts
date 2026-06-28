import { useDataChannel } from "@livekit/components-react"
import { useState } from "react"

import { decodeRoomEvent, encodeRoomEvent } from "./room-events"

export interface LivePoll {
  pollId: string
  name: string
  options: { label: string; value: string }[]
  allowedAnswersCount: number
}

export interface PollResults {
  counts: Record<string, number>
  total: number
}

export function useRoomPolls() {
  const [activePoll, setActivePoll] = useState<LivePoll | null>(null)
  const [results, setResults] = useState<PollResults | null>(null)
  const [hasAnswered, setHasAnswered] = useState(false)

  const { send } = useDataChannel((msg) => {
    const event = decodeRoomEvent(msg.payload)
    if (!event) return

    if (event.type === "poll_launched") {
      setActivePoll({
        pollId: event.data.pollId,
        name: event.data.name,
        options: event.data.options,
        allowedAnswersCount: event.data.allowedAnswersCount,
      })
      setResults(null)
      setHasAnswered(false)
    } else if (event.type === "poll_results") {
      setResults({ counts: event.data.counts, total: event.data.total })
    } else if (event.type === "poll_closed") {
      setActivePoll(null)
      setResults(null)
      setHasAnswered(false)
    }
  })

  function launchPoll(poll: LivePoll) {
    setActivePoll(poll)
    setResults(null)
    setHasAnswered(false)
    send(
      encodeRoomEvent({
        type: "poll_launched",
        data: {
          pollId: poll.pollId,
          name: poll.name,
          options: poll.options,
          allowedAnswersCount: poll.allowedAnswersCount,
        },
      }),
      { reliable: true },
    )
  }

  function revealResults(r: PollResults) {
    if (!activePoll) return
    setResults(r)
    send(
      encodeRoomEvent({
        type: "poll_results",
        data: { pollId: activePoll.pollId, counts: r.counts, total: r.total },
      }),
      { reliable: true },
    )
  }

  function closePoll() {
    if (!activePoll) return
    const id = activePoll.pollId
    setActivePoll(null)
    setResults(null)
    setHasAnswered(false)
    send(encodeRoomEvent({ type: "poll_closed", data: { pollId: id } }), { reliable: true })
  }

  function markAnswered() {
    setHasAnswered(true)
  }

  return { activePoll, results, hasAnswered, launchPoll, revealResults, closePoll, markAnswered }
}
