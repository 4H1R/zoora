import { useId } from "react"
import { useTranslation } from "react-i18next"

/**
 * Branded loading screen shown while shells resolve (auth org, admin guard).
 *
 * Self-contained: the brand mark, motion keyframes, and atmosphere all live here
 * so the component drops in anywhere without touching global styles. The "Z" mark
 * draws itself in with a stroke sweep, then a brand-gradient bar runs an
 * indeterminate sweep — no generic spinner. Honors prefers-reduced-motion,
 * dark mode (semantic tokens), and RTL (logical properties only).
 */
export function SplashScreen() {
  const { t } = useTranslation()
  const gradientId = useId()
  const scope = useId().replace(/[:]/g, "")

  return (
    <div
      data-splash={scope}
      className="splash-root bg-background relative flex h-screen w-full items-center justify-center overflow-hidden"
    >
      <div aria-hidden className="splash-glow splash-glow-a" />
      <div aria-hidden className="splash-glow splash-glow-b" />
      <div aria-hidden className="splash-grid" />

      <div className="splash-stack relative flex flex-col items-center gap-7">
        <div className="splash-mark">
          <svg
            viewBox="0 0 48 48"
            fill="none"
            role="img"
            aria-label={t("common.brandName")}
            className="size-20 drop-shadow-[0_8px_30px_oklch(0.627_0.194_149.214/0.35)]"
          >
            <defs>
              <linearGradient id={gradientId} x1="6" y1="4" x2="42" y2="46" gradientUnits="userSpaceOnUse">
                <stop offset="0" stopColor="#2fbd68" />
                <stop offset="0.55" stopColor="#16a34a" />
                <stop offset="1" stopColor="#15803d" />
              </linearGradient>
            </defs>
            <rect className="splash-tile" width="48" height="48" rx="13" fill={`url(#${gradientId})`} />
            <path
              className="splash-z"
              d="M14 16 H34 L14 32 H34"
              stroke="#ffffff"
              strokeWidth="5.4"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </div>

        <span className="splash-word inline-flex items-center text-2xl leading-none font-bold tracking-tight">
          {t("common.brandName")}
        </span>

        <div className="splash-bar" role="status" aria-label={t("common.loading", "Loading")}>
          <span className="splash-bar-fill" />
        </div>
      </div>

      <style>{`
        [data-splash="${scope}"] .splash-glow {
          position: absolute;
          border-radius: 9999px;
          filter: blur(80px);
          opacity: 0.5;
          pointer-events: none;
        }
        [data-splash="${scope}"] .splash-glow-a {
          inset-block-start: -12%;
          inset-inline-start: -8%;
          width: 38rem;
          height: 38rem;
          background: radial-gradient(circle, oklch(0.627 0.194 149.214 / 0.55), transparent 65%);
          animation: splashDriftA 14s ease-in-out infinite;
        }
        [data-splash="${scope}"] .splash-glow-b {
          inset-block-end: -16%;
          inset-inline-end: -10%;
          width: 32rem;
          height: 32rem;
          background: radial-gradient(circle, oklch(0.723 0.219 149.579 / 0.4), transparent 65%);
          animation: splashDriftB 17s ease-in-out infinite;
        }
        [data-splash="${scope}"] .splash-grid {
          position: absolute;
          inset: 0;
          pointer-events: none;
          background-image:
            linear-gradient(to right, oklch(0.627 0.194 149.214 / 0.05) 1px, transparent 1px),
            linear-gradient(to bottom, oklch(0.627 0.194 149.214 / 0.05) 1px, transparent 1px);
          background-size: 3rem 3rem;
          mask-image: radial-gradient(circle at center, black, transparent 70%);
        }

        [data-splash="${scope}"] .splash-mark {
          animation: splashRise 0.7s cubic-bezier(0.22, 1, 0.36, 1) both;
        }
        [data-splash="${scope}"] .splash-tile {
          transform-origin: center;
          animation: splashTile 0.7s cubic-bezier(0.22, 1, 0.36, 1) both;
        }
        [data-splash="${scope}"] .splash-z {
          stroke-dasharray: 78;
          stroke-dashoffset: 78;
          animation: splashDraw 0.9s cubic-bezier(0.65, 0, 0.35, 1) 0.25s forwards;
        }
        [data-splash="${scope}"] .splash-word {
          color: var(--foreground);
          animation: splashRise 0.7s cubic-bezier(0.22, 1, 0.36, 1) 0.35s both;
        }

        [data-splash="${scope}"] .splash-bar {
          position: relative;
          width: 9rem;
          height: 3px;
          border-radius: 9999px;
          overflow: hidden;
          background: oklch(0.627 0.194 149.214 / 0.12);
          animation: splashRise 0.7s cubic-bezier(0.22, 1, 0.36, 1) 0.5s both;
        }
        [data-splash="${scope}"] .splash-bar-fill {
          position: absolute;
          inset-block: 0;
          inline-size: 45%;
          border-radius: 9999px;
          background: linear-gradient(90deg, #2fbd68, #16a34a, #15803d);
          animation: splashSweep 1.4s cubic-bezier(0.65, 0, 0.35, 1) 0.9s infinite;
        }

        @keyframes splashRise {
          from { opacity: 0; transform: translateY(10px); }
          to { opacity: 1; transform: translateY(0); }
        }
        @keyframes splashTile {
          from { transform: scale(0.82); }
          to { transform: scale(1); }
        }
        @keyframes splashDraw {
          to { stroke-dashoffset: 0; }
        }
        @keyframes splashSweep {
          0% { inset-inline-start: -45%; }
          100% { inset-inline-start: 100%; }
        }
        @keyframes splashDriftA {
          0%, 100% { transform: translate(0, 0); }
          50% { transform: translate(4%, 6%); }
        }
        @keyframes splashDriftB {
          0%, 100% { transform: translate(0, 0); }
          50% { transform: translate(-5%, -4%); }
        }

        @media (prefers-reduced-motion: reduce) {
          [data-splash="${scope}"] .splash-mark,
          [data-splash="${scope}"] .splash-tile,
          [data-splash="${scope}"] .splash-word,
          [data-splash="${scope}"] .splash-bar,
          [data-splash="${scope}"] .splash-glow-a,
          [data-splash="${scope}"] .splash-glow-b { animation: none; }
          [data-splash="${scope}"] .splash-z {
            stroke-dashoffset: 0;
            animation: none;
          }
          [data-splash="${scope}"] .splash-bar-fill {
            position: relative;
            inline-size: 100%;
            animation: splashPulse 1.4s ease-in-out infinite;
          }
          @keyframes splashPulse {
            0%, 100% { opacity: 0.4; }
            50% { opacity: 1; }
          }
        }
      `}</style>
    </div>
  )
}
