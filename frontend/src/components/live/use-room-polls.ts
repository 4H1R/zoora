import { useRoomContext } from "@livekit/components-react"
import { RoomEvent } from "livekit-client"
import { useEffect, useRef, useState } from "react"

import { decodeRoomEvent, encodeRoomEvent } from "./room-events"
import { useRoomChannel } from "./use-room-channel"

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

// Room-level poll session state. MUST be mounted for the whole room lifetime
// (in RoomShell), NOT inside a tab panel — otherwise the data-channel listener
// unmounts on tab switch and participants miss `poll_launched` broadcasts.
export function useRoomPolls(isHost: boolean) {
  const room = useRoomContext()
  const [activePoll, setActivePoll] = useState<LivePoll | null>(null)
  const [results, setResults] = useState<PollResults | null>(null)
  const [hasAnswered, setHasAnswered] = useState(false)

  // Latest values for the participant-connected listener (avoids re-subscribing).
  const activePollRef = useRef<LivePoll | null>(null)
  const resultsRef = useRef<PollResults | null>(null)
  activePollRef.current = activePoll
  resultsRef.current = results

  const { send } = useRoomChannel(undefined, (msg) => {
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

  // Late-join sync: when a new participant connects, the host re-broadcasts the
  // current poll (and revealed results) so latecomers and tab-switchers catch up.
  // Data-channel messages are ephemeral, so without this a student who joins mid
  // -poll — or whose panel was unmounted — never sees the vote prompt.
  useEffect(() => {
    if (!isHost || !room) return
    const onConnected = () => {
      const poll = activePollRef.current
      if (!poll) return
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
      const r = resultsRef.current
      if (r) {
        send(
          encodeRoomEvent({
            type: "poll_results",
            data: { pollId: poll.pollId, counts: r.counts, total: r.total },
          }),
          { reliable: true },
        )
      }
    }
    room.on(RoomEvent.ParticipantConnected, onConnected)
    return () => {
      room.off(RoomEvent.ParticipantConnected, onConnected)
    }
  }, [isHost, room, send])

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

export type RoomPolls = ReturnType<typeof useRoomPolls>
