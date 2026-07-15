import type { TLAssetStore } from "tldraw"

import { postLiveRoomsIdWhiteboardMediaPresign } from "@/api/live-sessions/live-sessions"

// Custom tldraw asset store for the live-room whiteboard.
//
// Without this, tldraw's default store base64-encodes every inserted image
// *inline* into the `asset:` record. That record then rides the tldraw diff we
// broadcast over the LiveKit reliable data channel (see use-whiteboard.ts), which
// has a ~15 KiB per-message cap — so any real image is silently dropped and never
// reaches other participants. Here we instead upload the binary to S3
// (whiteboards/<room>/ in the public bucket) and store only the short public URL
// in the record, so the synced diff stays tiny and peers see the image.
export function createWhiteboardAssetStore(liveId: string): TLAssetStore {
  return {
    async upload(_asset, file, abortSignal) {
      const res = await postLiveRoomsIdWhiteboardMediaPresign(
        liveId,
        {
          file_name: file.name,
          mime_type: file.type || "application/octet-stream",
          size: file.size,
        },
        { signal: abortSignal }
      )
      const presign = res.status === 200 ? res.data.data : undefined
      if (!presign?.upload_url || !presign.public_url) {
        throw new Error("whiteboard image presign failed")
      }

      const put = await fetch(presign.upload_url, {
        method: "PUT",
        body: file,
        headers: { "Content-Type": file.type || "application/octet-stream" },
        signal: abortSignal,
      })
      if (!put.ok) throw new Error(`whiteboard image upload failed: ${put.status}`)

      // resolve() defaults to asset.props.src, so returning the permanent public
      // URL here is all peers/late-joiners need — no re-signing on render.
      return { src: presign.public_url }
    },
  }
}
