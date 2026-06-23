/**
 * Shared atmospheric backdrop — aurora blobs, masked grid, grain and edge fades.
 * Powers the landing page and the status (404 / error) screens so they share one
 * visual language. `tone` swaps the accent: brand green for marketing/empty
 * states, a destructive red wash for failures.
 */

export const GRAIN =
  "url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='160' height='160'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E\")"

type Tone = "brand" | "alert"

const TONES: Record<Tone, { wash: string; blobA: string; blobB: string }> = {
  brand: {
    wash: "color-mix(in oklch, var(--primary) 14%, transparent)",
    blobA: "var(--green-500)",
    blobB: "var(--green-700)",
  },
  alert: {
    wash: "color-mix(in oklch, var(--destructive) 16%, transparent)",
    blobA: "color-mix(in oklch, var(--destructive) 78%, var(--green-600))",
    blobB: "var(--destructive)",
  },
}

export function BackgroundFX({ tone = "brand" }: { tone?: Tone }) {
  const c = TONES[tone]
  return (
    <div aria-hidden className="pointer-events-none absolute inset-0 overflow-hidden">
      <div
        className="absolute inset-0"
        style={{ background: `radial-gradient(120% 80% at 50% -10%, ${c.wash}, transparent 60%)` }}
      />
      <div
        className="animate-aurora absolute -top-40 start-[-10%] size-[40rem] rounded-full opacity-50 blur-3xl"
        style={{ background: `radial-gradient(circle, ${c.blobA}, transparent 65%)` }}
      />
      <div
        className="animate-aurora-slow absolute -bottom-52 end-[-8%] size-[36rem] rounded-full opacity-40 blur-3xl"
        style={{ background: `radial-gradient(circle, ${c.blobB}, transparent 65%)` }}
      />
      <div
        className="absolute inset-0"
        style={{
          backgroundImage:
            "linear-gradient(to right, color-mix(in oklch, var(--foreground) 5%, transparent) 1px, transparent 1px), linear-gradient(to bottom, color-mix(in oklch, var(--foreground) 5%, transparent) 1px, transparent 1px)",
          backgroundSize: "64px 64px",
          maskImage: "radial-gradient(120% 90% at 50% 0%, black, transparent 75%)",
          WebkitMaskImage: "radial-gradient(120% 90% at 50% 0%, black, transparent 75%)",
        }}
      />
      <div
        className="absolute inset-0 opacity-[0.04] mix-blend-overlay dark:opacity-[0.06]"
        style={{ backgroundImage: GRAIN }}
      />
      <div
        className="absolute inset-x-0 bottom-0 h-40"
        style={{ background: "linear-gradient(to top, var(--background), transparent)" }}
      />
    </div>
  )
}
