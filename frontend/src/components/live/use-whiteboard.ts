import { useDataChannel } from "@livekit/components-react"
import { useEffect, useRef, useState } from "react"
import { createTLStore, getSnapshot, loadSnapshot } from "tldraw"
import type { TLEditorSnapshot, TLRecord, TLStore } from "tldraw"

import {
  useGetLiveRoomsIdWhiteboard,
  usePutLiveRoomsIdWhiteboard,
} from "@/api/live-sessions/live-sessions"

// ---- encoding helpers -------------------------------------------------------

const textEncoder = new TextEncoder()
const textDecoder = new TextDecoder()

function bytesToSnapshot(bytes: number[]): TLEditorSnapshot | null {
  if (!bytes || bytes.length === 0) return null
  try {
    const json = textDecoder.decode(new Uint8Array(bytes))
    const parsed = JSON.parse(json)
    // Guard: a valid tldraw snapshot has a document field
    if (parsed && typeof parsed === "object" && "document" in parsed) {
      return parsed as TLEditorSnapshot
    }
  } catch {
    // corrupt snapshot — treat as empty
  }
  return null
}

function snapshotToBytes(snap: TLEditorSnapshot): number[] {
  return Array.from(textEncoder.encode(JSON.stringify(snap)))
}

// ---- diff wire format -------------------------------------------------------

// added/updated hold serialized TLRecord objects; removed holds string IDs
interface TldrawDiff {
  added: TLRecord[]
  updated: TLRecord[]
  removed: string[]
}

// entry.changes shape: { added: Record<id, R>, updated: Record<id, [from, to]>, removed: Record<id, R> }
// We avoid importing RecordsDiff from @tldraw/store directly; use unknown + cast.
function encodeDiff(changes: {
  added: Record<string, TLRecord>
  updated: Record<string, [TLRecord, TLRecord]>
  removed: Record<string, TLRecord>
}): Uint8Array {
  const wire: TldrawDiff = {
    added: Object.values(changes.added),
    updated: Object.values(changes.updated).map((pair) => pair[1]),
    removed: Object.keys(changes.removed),
  }
  return textEncoder.encode(JSON.stringify(wire))
}

function decodeDiff(payload: Uint8Array): TldrawDiff | null {
  try {
    const parsed = JSON.parse(textDecoder.decode(payload))
    if (parsed && Array.isArray(parsed.added) && Array.isArray(parsed.removed)) {
      return parsed as TldrawDiff
    }
  } catch {
    // ignore malformed packets
  }
  return null
}

// ---- hook -------------------------------------------------------------------

export interface UseWhiteboardResult {
  store: TLStore
}

export function useWhiteboard(liveId: string, canDraw: boolean): UseWhiteboardResult {
  // useState with an initializer function creates the store exactly once
  const [store] = useState<TLStore>(() => createTLStore())

  const { send } = useDataChannel<"tldraw">("tldraw", (msg) => {
    const diff = decodeDiff(msg.payload)
    if (!diff) return
    store.mergeRemoteChanges(() => {
      if (diff.added.length > 0 || diff.updated.length > 0) {
        store.put([...diff.added, ...diff.updated] as TLRecord[])
      }
      if (diff.removed.length > 0) {
        // store.remove expects readonly RecordId[], string IDs are compatible
        store.remove(diff.removed as Parameters<typeof store.remove>[0])
      }
    })
  })

  const sendRef = useRef(send)
  sendRef.current = send

  // Load persisted snapshot on mount
  const { data: whiteboardRes } = useGetLiveRoomsIdWhiteboard(liveId)
  const snapshotLoadedRef = useRef(false)

  useEffect(() => {
    if (snapshotLoadedRef.current) return
    const bytes = whiteboardRes?.status === 200 ? whiteboardRes.data.data?.snapshot : undefined
    if (!bytes || bytes.length === 0) return
    const snap = bytesToSnapshot(bytes)
    if (!snap) return
    snapshotLoadedRef.current = true
    store.mergeRemoteChanges(() => {
      loadSnapshot(store, snap)
    })
  }, [whiteboardRes, store])

  // Persistence mutation — only used when canDraw
  const saveMutation = usePutLiveRoomsIdWhiteboard()
  const saveMutationRef = useRef(saveMutation)
  saveMutationRef.current = saveMutation
  const saveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Outgoing diffs + debounced persistence — re-attach whenever canDraw/liveId changes
  useEffect(() => {
    // Outgoing diffs: only when canDraw (host/presenter)
    const unsubscribeOutgoing = canDraw
      ? store.listen(
          (entry) => {
            // entry.changes: { added, updated, removed }
            const changes = entry.changes as {
              added: Record<string, TLRecord>
              updated: Record<string, [TLRecord, TLRecord]>
              removed: Record<string, TLRecord>
            }
            sendRef.current(encodeDiff(changes), { reliable: true })
          },
          { source: "user", scope: "document" },
        )
      : undefined

    // Debounced persistence: trigger on ALL changes when canDraw
    const unsubscribePersist = canDraw
      ? store.listen(
          () => {
            if (saveTimerRef.current) clearTimeout(saveTimerRef.current)
            saveTimerRef.current = setTimeout(() => {
              saveTimerRef.current = null
              const snap = getSnapshot(store)
              const bytes = snapshotToBytes(snap)
              saveMutationRef.current.mutate({ id: liveId, data: { snapshot: bytes } })
            }, 1500)
          },
          { source: "all", scope: "document" },
        )
      : undefined

    return () => {
      unsubscribeOutgoing?.()
      unsubscribePersist?.()
      if (saveTimerRef.current) {
        clearTimeout(saveTimerRef.current)
        saveTimerRef.current = null
      }
    }
  }, [canDraw, liveId, store])

  return { store }
}
