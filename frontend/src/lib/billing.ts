import { useTranslation } from "react-i18next"

// Backend stores money in Rial (minor units). The whole UI shows Toman, which is
// Rial / 10. These helpers are the single conversion point so pages never scatter
// the ÷10 / ×10 factor.
export function rialToToman(amountRial?: number): number {
  return Math.round((amountRial ?? 0) / 10)
}

export function tomanToRial(amountToman: number): number {
  return Math.round(amountToman * 10)
}

/**
 * Hook bound to the active language. Returns a formatter that converts a Rial
 * amount to a grouped Toman string (localized digits, no unit suffix).
 */
export function useFormatToman() {
  const { i18n } = useTranslation()
  return (amountRial?: number) => rialToToman(amountRial).toLocaleString(i18n.language)
}
