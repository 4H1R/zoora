# Zoora Videos

Tutorial & marketing videos built with [Remotion](https://remotion.dev). Persian (RTL), Vazirmatn, brand tokens mirrored from `frontend/src/styles.css`.

## Commands

```bash
bun run studio        # Remotion Studio — live preview & scrubbing
bun run render:intro  # Render the intro to out/intro.mp4
bun run still -- --frame=300   # Render a single frame to out/still.png
bun run typecheck
```

Run from `videos/` (deps install from the repo root via the bun workspace).

## Structure

```
src/
  Root.tsx              # Composition registry
  lib/
    tokens.ts           # Colors / font / radius mirrored from the frontend
    anim.ts             # Spring & fade helpers
    Subtitles.tsx       # Burned-in RTL subtitle track
  components/
    Logo.tsx            # Z-mark (from frontend/public/favicon.svg)
    Icon.tsx            # Minimal SVG icon set (no emoji — headless Chrome has no emoji font)
    Cursor.tsx          # Animated pointer with click ripples
    mock.tsx            # Recreated app UI: BrowserFrame, Sidebar, cards, toasts…
  intro/
    script.ts           # Scene timings + narration lines (single source of truth)
    Intro.tsx           # TransitionSeries of the six scenes
    scenes/             # Logo, Classes, Live, Quiz, Reach, Outro
docs/
  voiceover-intro.md    # Voice-over script with timestamps (matches script.ts)
```

## Conventions

- **One `script.ts` per video** holds scene durations and subtitle lines. Subtitles double as the voice-over script; keep `docs/voiceover-*.md` in sync when editing.
- **No emoji** in any scene — the render Chrome has no emoji font (tofu boxes). Add an SVG to `Icon.tsx` instead.
- Persian digits via `toFa()` from `lib/tokens.ts`; wrap Persian blocks in `dir="rtl"`.
- Voice-over is recorded separately and mixed onto the rendered MP4 in editing.

## Adding a new tutorial video

1. Create `src/<name>/script.ts` (timings + lines) and `src/<name>/<Name>.tsx`.
2. Reuse `components/mock.tsx` primitives to recreate the relevant app screens.
3. Register a `<Composition>` in `src/Root.tsx`, add a `render:<name>` script.
4. Write `docs/voiceover-<name>.md` from the script lines.

## Gotcha: TypeScript 7

The monorepo hoists frontend's TypeScript 7 (native preview) which lacks `ts.sys` and crashes Remotion's esbuild-loader. `remotion.config.ts` injects `tsconfigRaw` into the webpack config so the loader never `require('typescript')`s. Don't remove that override.
