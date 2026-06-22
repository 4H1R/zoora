import { useTranslation } from "react-i18next"

// slotForHour maps a 24h hour to a greeting slot.
function slotForHour(hour: number): "morning" | "afternoon" | "evening" | "night" {
  if (hour < 5) return "night"
  if (hour < 12) return "morning"
  if (hour < 17) return "afternoon"
  if (hour < 21) return "evening"
  return "night"
}

// useGreeting returns a time-aware greeting (e.g. "Good morning, Ali"),
// falling back to the generic welcome string when no name is available.
export function useGreeting(name: string): string {
  const { t } = useTranslation()
  const slot = slotForHour(new Date().getHours())
  if (name) return t(`org.dashboard.greetingName.${slot}`, { name })
  return t(`org.dashboard.greeting.${slot}`)
}
