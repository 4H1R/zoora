import type { TFunction } from "i18next"

// Preset poll options (the Yes/No template) carry stable string values ("yes" /
// "no"), while custom options use numeric indices ("0", "1", ...). The host bakes
// its own locale into the stored label, so each viewer re-resolves preset labels
// from the stable value to render Yes/No in their own language.
export function resolveOptionLabel(value: string, fallback: string, t: TFunction): string {
  if (value === "yes") return t("liveRoom.polls.yes")
  if (value === "no") return t("liveRoom.polls.no")
  return fallback
}
