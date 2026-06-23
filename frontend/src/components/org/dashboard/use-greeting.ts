import { useTranslation } from "react-i18next"

function slotForHour(hour: number): "morning" | "afternoon" | "evening" | "night" {
  if (hour < 5) return "night"
  if (hour < 12) return "morning"
  if (hour < 17) return "afternoon"
  if (hour < 21) return "evening"
  return "night"
}

export function useGreeting(name: string): string {
  const { t } = useTranslation()
  const slot = slotForHour(new Date().getHours())
  if (name) return t(`org.dashboard.greetingName.${slot}`, { name })
  return t(`org.dashboard.greeting.${slot}`)
}
