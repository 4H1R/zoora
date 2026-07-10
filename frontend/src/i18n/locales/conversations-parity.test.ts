import { describe, expect, it } from "vitest"

import en from "./en.json"
import fa from "./fa.json"

// Flatten a nested translation object into dot-separated leaf keys so the two
// locales' `conversations` subtrees can be compared as flat key sets. Every
// user-facing key added to one locale must exist in the other.
function flattenKeys(value: unknown, prefix = ""): string[] {
  if (value === null || typeof value !== "object") return [prefix]
  if (Array.isArray(value)) {
    return value.flatMap((item, index) => flattenKeys(item, `${prefix}[${index}]`))
  }
  return Object.entries(value as Record<string, unknown>).flatMap(([key, child]) =>
    flattenKeys(child, prefix ? `${prefix}.${key}` : key)
  )
}

describe("conversations i18n parity", () => {
  const enKeys = new Set(flattenKeys((en as Record<string, unknown>).conversations))
  const faKeys = new Set(flattenKeys((fa as Record<string, unknown>).conversations))

  it("has matching en/fa key sets under `conversations`", () => {
    const missingInFa = [...enKeys].filter((k) => !faKeys.has(k)).sort()
    const missingInEn = [...faKeys].filter((k) => !enKeys.has(k)).sort()

    expect({ missingInFa, missingInEn }).toEqual({ missingInFa: [], missingInEn: [] })
  })
})
