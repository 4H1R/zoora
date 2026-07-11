import { cn } from "@/lib/utils"

/**
 * Telegram-style tiled "doodle" wallpaper for the chat surface, themed around
 * learning science: a seamless SVG pattern of education glyphs (graduation cap,
 * atom, open book, flask, lightbulb, DNA, ruler, pencil) mixed with math marks
 * (π, Σ, √, ∞, ∫). Drawn inline with `currentColor` so it inherits the wrapper's
 * `text-foreground` and re-tints for light/dark automatically; kept faint via
 * opacity so message bubbles stay readable on top.
 *
 * Decorative only — `aria-hidden` + `pointer-events-none`. Mount as an absolute
 * layer inside a `relative isolate` parent so its negative z-index paints behind
 * the in-flow message list without slipping under the card background.
 */

interface ChatBackgroundProps {
  className?: string
}

export function ChatBackground({ className }: ChatBackgroundProps) {
  return (
    <svg
      aria-hidden
      className={cn(
        "pointer-events-none absolute inset-0 -z-10 h-full w-full",
        "text-foreground opacity-[0.04] dark:opacity-[0.06]",
        className
      )}
      xmlns="http://www.w3.org/2000/svg"
    >
      <defs>
        {/* Glyph definitions — stroked icons + a math-symbol group. */}
        <g id="cb-cap" fill="none" stroke="currentColor" strokeWidth={2} strokeLinejoin="round">
          <path d="M2 15 L20 7 L38 15 L20 23 Z" />
          <path d="M31 18 v8 c0 3 -22 3 -22 0 v-8" />
          <path d="M38 15 v9" />
        </g>
        <g id="cb-atom" fill="none" stroke="currentColor" strokeWidth={2}>
          <circle cx="20" cy="20" r="2.5" fill="currentColor" stroke="none" />
          <ellipse cx="20" cy="20" rx="16" ry="6" />
          <ellipse cx="20" cy="20" rx="16" ry="6" transform="rotate(60 20 20)" />
          <ellipse cx="20" cy="20" rx="16" ry="6" transform="rotate(120 20 20)" />
        </g>
        <g id="cb-book" fill="none" stroke="currentColor" strokeWidth={2} strokeLinejoin="round">
          <path d="M20 8 c-4 -3 -10 -3 -16 0 v22 c6 -3 12 -3 16 0 c4 -3 10 -3 16 0 V8 c-6 -3 -12 -3 -16 0 Z" />
          <path d="M20 8 v22" />
        </g>
        <g id="cb-flask" fill="none" stroke="currentColor" strokeWidth={2} strokeLinejoin="round">
          <path d="M16 5 h8 M17 5 v9 L8 32 a3 3 0 0 0 3 4 h18 a3 3 0 0 0 3 -4 L23 14 V5" />
          <path d="M12 26 h16" />
        </g>
        <g
          id="cb-bulb"
          fill="none"
          stroke="currentColor"
          strokeWidth={2}
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <path d="M20 4 a11 11 0 0 0 -7 19 c1 1 2 2 2 4 h10 c0 -2 1 -3 2 -4 a11 11 0 0 0 -7 -19 Z" />
          <path d="M16 32 h8 M17 36 h6" />
        </g>
        <g id="cb-dna" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round">
          <path d="M14 4 q12 8 0 16 q-12 8 0 16" />
          <path d="M26 4 q-12 8 0 16 q12 8 0 16" />
          <path d="M16 8 h8 M15 16 h10 M15 24 h10 M16 32 h8" />
        </g>
        <g id="cb-ruler" fill="none" stroke="currentColor" strokeWidth={2} strokeLinejoin="round">
          <rect x="4" y="15" width="32" height="10" rx="1" transform="rotate(-25 20 20)" />
          <path d="M11 12 l2 3 M16 9 l2 3 M21 7 l2 3 M26 4 l2 3" transform="rotate(-25 20 20)" />
        </g>
        <g id="cb-pencil" fill="none" stroke="currentColor" strokeWidth={2} strokeLinejoin="round">
          <path d="M6 34 l3 -9 18 -18 6 6 -18 18 -9 3 Z" />
          <path d="M24 7 l6 6 M9 25 l6 6" />
        </g>

        {/* One 300×300 doodle tile — icons scattered with math marks, rows offset
            so repeats read as a continuous field. */}
        <pattern id="cb-tile" patternUnits="userSpaceOnUse" width="300" height="300">
          <use href="#cb-cap" x="18" y="24" />
          <use href="#cb-atom" x="150" y="14" />
          <use href="#cb-book" x="248" y="30" />
          <use href="#cb-flask" x="86" y="96" />
          <use href="#cb-dna" x="210" y="90" />
          <use href="#cb-bulb" x="20" y="150" />
          <use href="#cb-ruler" x="130" y="152" />
          <use href="#cb-pencil" x="252" y="158" />
          <use href="#cb-atom" x="70" y="228" />
          <use href="#cb-cap" x="196" y="238" />
          <use href="#cb-book" x="10" y="256" />
          <use href="#cb-flask" x="256" y="250" />
          <g fill="currentColor" fontSize="26" fontFamily="serif" fontStyle="italic">
            <text x="118" y="82">
              π
            </text>
            <text x="250" y="132">
              Σ
            </text>
            <text x="60" y="130">
              √
            </text>
            <text x="182" y="200">
              ∞
            </text>
            <text x="18" y="216">
              ∫
            </text>
            <text x="118" y="288">
              Σ
            </text>
          </g>
        </pattern>
      </defs>

      <rect width="100%" height="100%" fill="url(#cb-tile)" />
    </svg>
  )
}
