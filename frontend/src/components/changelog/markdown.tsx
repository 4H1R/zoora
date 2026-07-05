import ReactMarkdown from "react-markdown"
import remarkGfm from "remark-gfm"

const VIDEO_EXT = /\.(mp4|webm|mov|m4v)(\?.*)?$/i

/**
 * ChangelogMarkdown renders trusted-admin markdown. Raw HTML is disabled
 * (no rehype-raw) so untrusted-looking markup can never inject scripts. Image
 * syntax `![](url)` renders <img>, but a video URL becomes a <video> element —
 * admins insert both via the same "insert media" button.
 */
export function ChangelogMarkdown({ children }: { children: string }) {
  return (
    <div className="prose prose-sm dark:prose-invert max-w-none prose-img:rounded-lg prose-video:rounded-lg">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          img: ({ src, alt }) => {
            const url = typeof src === "string" ? src : ""
            if (VIDEO_EXT.test(url)) {
              return (
                <video
                  src={url}
                  controls
                  preload="metadata"
                  className="w-full rounded-lg"
                />
              )
            }
            return <img src={url} alt={alt ?? ""} loading="lazy" className="rounded-lg" />
          },
        }}
      >
        {children}
      </ReactMarkdown>
    </div>
  )
}
