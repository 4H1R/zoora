const PLACEHOLDER = "—"

export interface FormatScoreOptions {
  digits?: number
  fallback?: string
  trimZeros?: boolean
}

export function formatScore(
  value: number | null | undefined,
  { digits = 2, fallback = PLACEHOLDER, trimZeros = false }: FormatScoreOptions = {}
): string {
  if (value == null || !Number.isFinite(value)) return fallback
  const fixed = value.toFixed(digits)
  if (!trimZeros) return fixed
  return fixed.replace(/\.?0+$/, "")
}
