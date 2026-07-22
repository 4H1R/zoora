import type { ComponentProps } from "react"

import ReactMarkdown from "react-markdown"
import remarkGfm from "remark-gfm"

const VIDEO_EXT = /\.(mp4|webm|mov|m4v)(\?.*)?$/i

// Custom renderer for markdown `![](url)`: a video URL becomes a <video>, any
// other URL an <img>. Hoisted to module scope so it isn't redefined per render.
function MarkdownImage({ src, alt }: ComponentProps<"img">) {
  const url = typeof src === "string" ? src : ""
  if (VIDEO_EXT.test(url)) {
    return <video src={url} controls preload="metadata" className="w-full rounded-lg" />
  }
  return <img src={url} alt={alt ?? ""} loading="lazy" className="rounded-lg" />
}

/**
 * ChangelogMarkdown renders trusted-admin markdown. Raw HTML is disabled
 * (no rehype-raw) so untrusted-looking markup can never inject scripts. Image
 * syntax `![](url)` renders <img>, but a video URL becomes a <video> element —
 * admins insert both via the same "insert media" button.
 */
export function ChangelogMarkdown({ children }: { children: string }) {
  return (
    <div className="prose prose-sm dark:prose-invert prose-img:rounded-lg prose-video:rounded-lg max-w-none">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{ img: MarkdownImage }}
      >
        {children}
      </ReactMarkdown>
    </div>
  )
}
