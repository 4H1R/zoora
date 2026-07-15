import type { $ZodErrorMap, $ZodRawIssue } from "zod/v4/core"

import i18n from "i18next"
import { z } from "zod"
import { en as zodEn, fa as zodFa } from "zod/v4/locales"

/**
 * Centralized Zod localization — Laravel-style, field-name-aware messages.
 *
 * Every validation error flows through `customError`, which resolves a key from
 * the `validation` i18n namespace and returns an already-translated string with
 * the field name injected as `{{attribute}}`. Messages mirror Laravel's
 * `validation.php`: per-type variants (`min.string` / `min.numeric` / `min.array`)
 * and a `validation.attributes` map of human-readable field names.
 *
 * Components render `error.message` directly — no per-field `t()`.
 *
 * Anything we don't explicitly handle falls through to Zod's bundled locale
 * (`localeError`) so rare codes still get a sane translated message.
 *
 * Custom `.refine()` checks opt in by passing `params: { i18n: "validation.x" }`
 * (plus any interpolation values), e.g.
 *   .refine(fn, { path: ["end_time"], params: { i18n: "validation.endAfterStart" } })
 */

const bundledLocales: Record<string, () => { localeError: $ZodErrorMap }> = {
  en: zodEn,
  fa: zodFa,
}

/** Resolve a human-readable field name from the issue path (Laravel `:attribute`). */
function attribute(path: PropertyKey[] | undefined): string {
  const field = String(path?.[path.length - 1] ?? "")
  const key = `validation.attributes.${field}`
  if (field && i18n.exists(key)) return i18n.t(key) as unknown as string
  // Fallback: humanize (drop trailing _id, underscores -> spaces).
  return field.replace(/_id$/, "").replace(/_/g, " ")
}

function tv(key: string, path: PropertyKey[] | undefined, values?: Record<string, unknown>): string {
  return i18n.t(`validation.${key}`, { attribute: attribute(path), ...values } as never) as unknown as string
}

/** Map a Zod numeric/collection origin to a Laravel size-rule variant. */
function sizeVariant(origin: string): "string" | "numeric" | "array" {
  if (origin === "string") return "string"
  if (origin === "array" || origin === "set") return "array"
  return "numeric"
}

const customError: $ZodErrorMap = (issue: $ZodRawIssue) => {
  const path = issue.path as PropertyKey[] | undefined

  switch (issue.code) {
    case "invalid_type": {
      if (issue.input === undefined || issue.input === null) return tv("required", path)
      if (issue.expected === "string") return tv("string", path)
      if (issue.expected === "number" || issue.expected === "bigint") return tv("numeric", path)
      if (issue.expected === "array") return tv("array", path)
      return tv("invalid", path)
    }

    case "too_small": {
      const min = Number(issue.minimum)
      // A string min of 1 (or 0) reads as "required" rather than a length rule.
      if (issue.origin === "string" && min <= 1) return tv("required", path)
      return tv(`min.${sizeVariant(String(issue.origin))}`, path, { min })
    }

    case "too_big": {
      const max = Number(issue.maximum)
      return tv(`max.${sizeVariant(String(issue.origin))}`, path, { max })
    }

    case "not_multiple_of":
      return Number(issue.divisor) === 1 ? tv("integer", path) : tv("multiple_of", path, { value: issue.divisor })

    case "invalid_format": {
      const format = issue.format
      if (format === "email") return tv("email", path)
      if (format === "url") return tv("url", path)
      if (format === "uuid" || format === "guid") return tv("uuid", path)
      return tv("regex", path)
    }

    case "invalid_value":
      return tv("enum", path)

    case "unrecognized_keys":
      return tv("invalid", path)

    case "custom": {
      // Opt-in translation for `.refine()` / `.superRefine()` checks.
      const key = (issue.params as { i18n?: string } | undefined)?.i18n
      if (key) return i18n.t(key, { attribute: attribute(path), ...issue.params } as never) as unknown as string
      return undefined
    }

    default:
      return undefined
  }
}

/** Apply the global Zod config for a given language. Safe to call repeatedly. */
export function configureZodLocale(lng: string | undefined): void {
  const base = (lng ?? "en").split("-")[0]
  const locale = (bundledLocales[base] ?? bundledLocales.en)()
  z.config({ customError, localeError: locale.localeError })
}
