import { useTranslation } from "react-i18next"

export function usePermissionLabel() {
  const { t } = useTranslation()
  return (name: string) => {
    const [resource, action] = name.split(":")
    const r = t(`permissions.resources.${resource}`, { defaultValue: resource })
    const a = t(`permissions.actions.${action}`, { defaultValue: action })
    return `${r}: ${a}`
  }
}

export function useRoleName() {
  const { t } = useTranslation()
  return (name: string) => t(`roles.presets.${name}`, { defaultValue: name })
}
